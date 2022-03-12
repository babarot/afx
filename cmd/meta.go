package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/b4b4r07/afx/pkg/config"
	"github.com/b4b4r07/afx/pkg/env"
	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/github"
	"github.com/b4b4r07/afx/pkg/helpers/shell"
	"github.com/b4b4r07/afx/pkg/printers"
	"github.com/b4b4r07/afx/pkg/state"
	"github.com/b4b4r07/afx/pkg/update"
	"github.com/fatih/color"
	"github.com/mattn/go-shellwords"
)

type metaCmd struct {
	env      *env.Config
	packages []config.Package
	main     *config.Main
	state    *state.State
	configs  map[string]config.Config

	updateMessageChan chan *update.ReleaseInfo
}

func (m *metaCmd) init() error {
	m.updateMessageChan = make(chan *update.ReleaseInfo)
	go func() {
		log.Printf("[DEBUG] (goroutine): checking new updates...")
		release, err := checkForUpdate(Version)
		if err != nil {
			log.Printf("[ERROR] (goroutine): cannot check for new updates: %s", err)
		}
		m.updateMessageChan <- release
	}()

	root := filepath.Join(os.Getenv("HOME"), ".afx")
	cfgRoot := filepath.Join(os.Getenv("HOME"), ".config", "afx")
	cache := filepath.Join(root, "cache.json")

	files, err := config.WalkDir(cfgRoot)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to walk dir", cfgRoot)
	}

	var pkgs []config.Package
	m.configs = map[string]config.Config{}

	cfg, err := config.Read(filepath.Join(cfgRoot, "main.yaml"))
	if err != nil {
		log.Printf("[INFO] main.yaml is not set")
	}
	m.main = cfg.Main

	for _, file := range files {
		if !m.main.Loadable(file) {
			continue
		}
		cfg, err := config.Read(file)
		if err != nil {
			return errors.Wrapf(err, "%s: failed to read config", file)
		}
		parsed, err := cfg.Parse()
		if err != nil {
			return errors.Wrapf(err, "%s: failed to parse config", file)
		}
		pkgs = append(pkgs, parsed...)
		m.configs[file] = cfg
	}

	if err := config.Validate(pkgs); err != nil {
		return errors.Wrap(err, "failed to validate packages")
	}

	pkgs, err = config.Sort(pkgs)
	if err != nil {
		return errors.Wrap(err, "failed to resolve dependencies between packages")
	}

	m.packages = pkgs

	m.env = env.New(cache)
	m.env.Add(env.Variables{
		"AFX_CONFIG_PATH":  env.Variable{Value: cfgRoot},
		"AFX_LOG":          env.Variable{},
		"AFX_LOG_PATH":     env.Variable{},
		"AFX_COMMAND_PATH": env.Variable{Default: filepath.Join(os.Getenv("HOME"), "bin")},
		"AFX_SHELL":        env.Variable{Default: m.main.Shell},
		"AFX_SUDO_PASSWORD": env.Variable{
			Input: env.Input{
				When:    config.HasSudoInCommandBuildSteps(m.packages),
				Message: "Please enter sudo command password",
				Help:    "Some packages build steps requires sudo command",
			},
		},
		"GITHUB_TOKEN": env.Variable{
			Input: env.Input{
				When:    config.HasGitHubReleaseBlock(m.packages),
				Message: "Please type your GITHUB_TOKEN",
				Help:    "To fetch GitHub Releases, GitHub token is required",
			},
		},
		"AFX_NO_UPDATE_NOTIFIER": env.Variable{},
	})

	for k, v := range m.main.Env {
		log.Printf("[DEBUG] main: set env: %s=%s", k, v)
		os.Setenv(k, v)
	}

	log.Printf("[DEBUG] mkdir %s\n", root)
	os.MkdirAll(root, os.ModePerm)

	log.Printf("[DEBUG] mkdir %s\n", os.Getenv("AFX_COMMAND_PATH"))
	os.MkdirAll(os.Getenv("AFX_COMMAND_PATH"), os.ModePerm)

	resourcers := make([]state.Resourcer, len(m.packages))
	for i, pkg := range m.packages {
		resourcers[i] = pkg
	}

	s, err := state.Open(filepath.Join(root, "state.json"), resourcers)
	if err != nil {
		return errors.Wrap(err, "faield to open state file")
	}
	m.state = s

	log.Printf("[INFO] state additions: (%d) %#v", len(s.Additions), state.Keys(s.Additions))
	log.Printf("[INFO] state deletions: (%d) %#v", len(s.Deletions), state.Keys(s.Deletions))
	log.Printf("[INFO] state changes: (%d) %#v", len(s.Changes), state.Keys(s.Changes))
	log.Printf("[INFO] state unchanges: (%d) []string{...skip...}", len(s.NoChanges))

	return nil
}

func printForUpdate(uriCh chan *update.ReleaseInfo) {
	switch Version {
	case "unset":
		return
	}
	log.Printf("[DEBUG] checking updates on afx repo...")
	newRelease := <-uriCh
	if newRelease != nil {
		fmt.Fprintf(os.Stdout, "\n\n%s %s -> %s\n",
			color.YellowString("A new release of afx is available:"),
			color.CyanString("v"+Version),
			color.CyanString(newRelease.Version))
		fmt.Fprintf(os.Stdout, "%s\n\n", color.YellowString(newRelease.URL))
		fmt.Fprintf(os.Stdout, "To upgrade, run: afx self-update\n")
	}
}

func (m *metaCmd) printForUpdate() error {
	if m.updateMessageChan == nil {
		return errors.New("update message chan is not set")
	}
	printForUpdate(m.updateMessageChan)
	return nil
}

func (m *metaCmd) prompt() (config.Package, error) {
	if m.main.FilterCmd == "" {
		return nil, errors.New("filter_command is not set")
	}

	var stdin, stdout bytes.Buffer

	p := shellwords.NewParser()
	p.ParseEnv = true
	p.ParseBacktick = true

	args, err := p.Parse(m.main.FilterCmd)
	if err != nil {
		return nil, errors.New("failed to parse filter command in main config")
	}

	cmd := shell.Shell{
		Stdin:   &stdin,
		Stdout:  &stdout,
		Stderr:  os.Stderr,
		Command: args[0],
		Args:    args[1:],
	}

	for _, pkg := range m.packages {
		fmt.Fprintln(&stdin, pkg.GetName())
	}

	if err := cmd.Run(context.Background()); err != nil {
		return nil, err
	}

	search := func(name string) config.Package {
		for _, pkg := range m.packages {
			if pkg.GetName() == name {
				return pkg
			}
		}
		return nil
	}

	for _, line := range strings.Split(stdout.String(), "\n") {
		if pkg := search(line); pkg != nil {
			return pkg, nil
		}
	}

	return nil, errors.New("pkg not found")
}

func (m *metaCmd) askRunCommand(op interface{}, pkgs []string) (bool, error) {
	var do string
	switch op.(type) {
	case installCmd:
		do = "install"
	case uninstallCmd:
		do = "uninstall"
	case updateCmd:
		do = "update"
	case checkCmd:
		do = "check"
	default:
		return false, errors.New("unsupported command type")
	}

	length := 3
	target := strings.Join(pkgs, ", ")
	if len(pkgs) > length {
		target = fmt.Sprintf("%s, ... (%d packages)", strings.Join(pkgs[:length], ", "), len(pkgs))
	}

	yes := false
	confirm := survey.Confirm{
		Message: fmt.Sprintf("OK to %s these packages? %s", do, color.YellowString(target)),
	}

	if len(pkgs) > length {
		helpMessage := "\n"
		sort.Strings(pkgs)
		for _, pkg := range pkgs {
			helpMessage += fmt.Sprintf("- %s\n", pkg)
		}
		confirm.Help = helpMessage
	}

	if err := survey.AskOne(&confirm, &yes); err != nil {
		return false, errors.Wrap(err, "failed to get input from console")
	}
	return yes, nil
}

func shouldCheckForUpdate() bool {
	if os.Getenv("AFX_NO_UPDATE_NOTIFIER") != "" {
		return false
	}
	return !isCI() && printers.IsTerminal(os.Stdout) && printers.IsTerminal(os.Stderr)
}

// based on https://github.com/watson/ci-info/blob/HEAD/index.js
func isCI() bool {
	return os.Getenv("CI") != "" || // GitHub Actions, Travis CI, CircleCI, Cirrus CI, GitLab CI, AppVeyor, CodeShip, dsari
		os.Getenv("BUILD_NUMBER") != "" || // Jenkins, TeamCity
		os.Getenv("RUN_ID") != "" // TaskCluster, dsari
}

func checkForUpdate(currentVersion string) (*update.ReleaseInfo, error) {
	if !shouldCheckForUpdate() {
		return nil, nil
	}
	client := github.NewClient()
	stateFilePath := filepath.Join(os.Getenv("HOME"), ".afx", "version.json")
	return update.CheckForUpdate(client, stateFilePath, Repository, Version)
}

func (m metaCmd) GetPackage(resource state.Resource) config.Package {
	for _, pkg := range m.packages {
		if pkg.GetName() == resource.Name {
			return pkg
		}
	}
	return nil
}

func (m metaCmd) GetPackages(resources []state.Resource) []config.Package {
	var pkgs []config.Package
	for _, resource := range resources {
		pkgs = append(pkgs, m.GetPackage(resource))
	}
	return pkgs
}

func (m metaCmd) GetConfig() config.Config {
	var all config.Config
	for _, config := range m.configs {
		if config.Main != nil {
			all.Main = config.Main
		}
		all.GitHub = append(all.GitHub, config.GitHub...)
		all.Gist = append(all.Gist, config.Gist...)
		all.HTTP = append(all.HTTP, config.HTTP...)
		all.Local = append(all.Local, config.Local...)
	}
	return all
}
