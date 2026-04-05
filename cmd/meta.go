package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"

	"github.com/babarot/afx/internal/env"
	"github.com/babarot/afx/internal/errors"
	"github.com/babarot/afx/internal/github"
	afxpkg "github.com/babarot/afx/internal/pkg"
	"github.com/babarot/afx/internal/printers"
	"github.com/babarot/afx/internal/state"
	"github.com/babarot/afx/internal/update"
)

type metaCmd struct {
	env      *env.Config
	packages []afxpkg.Package
	main     *afxpkg.Main
	state    *state.State
	configs  map[string]afxpkg.Config

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

	if err := m.loadConfigs(); err != nil {
		return err
	}
	if err := m.initPackages(); err != nil {
		return err
	}
	if err := m.initEnv(); err != nil {
		return err
	}
	return m.initState()
}

// loadConfigs reads and parses all YAML config files from the config directory.
func (m *metaCmd) loadConfigs() error {
	cfgRoot := afxpkg.ConfigDir()

	if err := afxpkg.CreateDirIfNotExist(cfgRoot); err != nil {
		return errors.Wrapf(err, "%s: failed to create dir", cfgRoot)
	}
	files, err := afxpkg.WalkDir(cfgRoot)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to walk dir", cfgRoot)
	}

	var pkgs []afxpkg.Package
	app := &afxpkg.DefaultMain
	m.configs = map[string]afxpkg.Config{}
	for _, file := range files {
		cfg, err := afxpkg.Read(file)
		if err != nil {
			return errors.Wrapf(err, "%s: failed to read config", file)
		}
		parsed, err := cfg.Parse()
		if err != nil {
			return errors.Wrapf(err, "%s: failed to parse config", file)
		}
		pkgs = append(pkgs, parsed...)
		m.configs[file] = cfg
		if cfg.Main != nil {
			app = cfg.Main
		}
	}

	m.main = app
	m.packages = pkgs
	return nil
}

// initPackages validates and sorts packages by dependency order.
func (m *metaCmd) initPackages() error {
	if err := afxpkg.Validate(m.packages); err != nil {
		return errors.Wrap(err, "failed to validate packages")
	}
	sorted, err := afxpkg.Sort(m.packages)
	if err != nil {
		return errors.Wrap(err, "failed to resolve dependencies between packages")
	}
	m.packages = sorted
	return nil
}

// initEnv sets up environment variables and creates required directories.
func (m *metaCmd) initEnv() error {
	root := afxpkg.DataDir()
	cfgRoot := afxpkg.ConfigDir()
	cache := filepath.Join(root, "cache.json")

	m.env = env.New(cache)
	_ = m.env.Add(env.Variables{
		"AFX_CONFIG_PATH":  env.Variable{Value: cfgRoot},
		"AFX_LOG":          env.Variable{},
		"AFX_LOG_PATH":     env.Variable{},
		"AFX_COMMAND_PATH": env.Variable{Default: afxpkg.BinDir()},
		"AFX_SHELL":        env.Variable{Default: m.main.Shell},
		"AFX_SUDO_PASSWORD": env.Variable{
			Input: env.Input{
				When:    afxpkg.HasSudoInCommandBuildSteps(m.packages),
				Message: "Please enter sudo command password",
				Help:    "Some packages build steps requires sudo command",
			},
		},
		"GITHUB_TOKEN": env.Variable{
			Input: env.Input{
				When:    afxpkg.HasGitHubReleaseBlock(m.packages),
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

	_ = os.MkdirAll(root, os.ModePerm)
	_ = os.MkdirAll(os.Getenv("AFX_COMMAND_PATH"), os.ModePerm)
	return nil
}

// initState opens the state file and logs the current state summary.
func (m *metaCmd) initState() error {
	root := afxpkg.DataDir()

	resourcers := make([]state.Resourcer, len(m.packages))
	for i, pkg := range m.packages {
		resourcers[i] = pkg
	}

	s, err := state.Open(filepath.Join(root, "state.json"), resourcers)
	if err != nil {
		return errors.Wrap(err, "failed to open state file")
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

func (m *metaCmd) askRunCommand(op any, pkgs []string) (bool, error) {
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
		var sb strings.Builder
		sb.WriteString("\n")
		sort.Strings(pkgs)
		for _, pkg := range pkgs {
			fmt.Fprintf(&sb, "- %s\n", pkg)
		}
		confirm.Help = sb.String()
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
	stateFilePath := filepath.Join(afxpkg.DataDir(), "version.json")
	return update.CheckForUpdate(client, stateFilePath, Repository, Version)
}

func (m metaCmd) GetPackage(resource state.Resource) afxpkg.Package {
	for _, pkg := range m.packages {
		if pkg.GetName() == resource.Name {
			return pkg
		}
	}
	return nil
}

func (m metaCmd) GetPackages(resources []state.Resource) []afxpkg.Package {
	var pkgs []afxpkg.Package
	for _, resource := range resources {
		pkgs = append(pkgs, m.GetPackage(resource))
	}
	return pkgs
}

func (m metaCmd) GetConfig() afxpkg.Config {
	var all afxpkg.Config
	for _, cfg := range m.configs {
		if cfg.Main != nil {
			all.Main = cfg.Main
		}
		all.GitHub = append(all.GitHub, cfg.GitHub...)
		all.Gist = append(all.Gist, cfg.Gist...)
		all.HTTP = append(all.HTTP, cfg.HTTP...)
		all.Local = append(all.Local, cfg.Local...)
	}
	return all
}
