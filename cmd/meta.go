package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/b4b4r07/afx/pkg/config"
	"github.com/b4b4r07/afx/pkg/config/loader"
	"github.com/b4b4r07/afx/pkg/env"
	"github.com/b4b4r07/afx/pkg/schema"
	"github.com/manifoldco/promptui"
)

// meta is
type meta struct {
	// UI cli.Ui
	data schema.Data

	Env      *env.Config
	Packages []config.Package
	// Loader   *loader.Loader

	paths    []string
	parseErr error
}

func (m *meta) init(args []string) error {
	root := filepath.Join(os.Getenv("HOME"), ".afx")
	cfg := filepath.Join(os.Getenv("HOME"), ".config", "afx")
	cache := filepath.Join(root, ".cache.json")

	m.Env = env.New(cache)
	m.Env.Add("AFX_ROOT", env.Variable{Default: root})

	data, err := loader.Load(cfg)
	if err != nil {
		return err
	}
	m.data = data
	for file := range data.Files {
		m.paths = append(m.paths, file)
	}

	pkgs, err := config.Parse(data)
	if err != nil {
		m.parseErr = err
	}
	m.Packages = pkgs

	m.Env.Add(env.Variables{
		"AFX_CONFIG_ROOT":  env.Variable{Value: cfg},
		"AFX_LOG":          env.Variable{},
		"AFX_LOG_PATH":     env.Variable{},
		"AFX_COMMAND_PATH": env.Variable{Default: filepath.Join(os.Getenv("HOME"), "bin")},
		"AFX_SUDO_PASSWORD": env.Variable{
			Input: env.Input{
				When:    config.HasSudoInCommandBuildSteps(pkgs),
				Message: "Please enter sudo command password",
				Help:    "Some packages build steps requires sudo command",
			},
		},
		"GITHUB_TOKEN": env.Variable{
			Input: env.Input{
				When:    config.HasGitHubReleaseBlock(pkgs),
				Message: "Please type your GITHUB_TOKEN",
				Help:    "To fetch GitHub Releases, GitHub token is required",
			},
		},
	})

	log.Printf("[DEBUG] mkdir %s\n", os.Getenv("AFX_ROOT"))
	os.MkdirAll(os.Getenv("AFX_ROOT"), os.ModePerm)

	log.Printf("[DEBUG] mkdir %s\n", os.Getenv("AFX_COMMAND_PATH"))
	os.MkdirAll(os.Getenv("AFX_COMMAND_PATH"), os.ModePerm)

	return nil
}

func (m *meta) Prompt() (config.Package, error) {
	// https://github.com/manifoldco/promptui
	// https://github.com/iwittkau/mage-select
	type item struct {
		Package config.Package
		Plugin  bool
		Command bool
		Name    string
		Type    string
		Home    string
		Slug    string
	}

	var items []item
	for _, pkg := range m.Packages {
		if !pkg.Installed() {
			continue
		}
		items = append(items, item{
			Package: pkg,
			Plugin:  pkg.HasPluginBlock(),
			Command: pkg.HasCommandBlock(),
			Name:    pkg.GetName(),
			Type:    pkg.GetType(),
			Home:    pkg.GetHome(),
			Slug:    pkg.GetSlug(),
		})
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   promptui.IconSelect + " {{ .Slug | cyan }}",
		Inactive: "  {{ .Slug | faint }}",
		Selected: promptui.IconGood + " {{ .Slug }}",
		Details: `
{{ "Type:" | faint }}	{{ .Type }}
{{ "Command:" | faint }}	{{ .Command }}
{{ "Plugin:" | faint }}	{{ .Plugin }}
`,
		// FuncMap: template.FuncMap{ // TODO: do not overwrite
		// 	"toupper": strings.ToUpper,
		// },
	}

	size := 5
	if len(items) < size {
		size = len(items)
	}

	searcher := func(input string, index int) bool {
		item := items[index]
		name := strings.Replace(strings.ToLower(item.Slug), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:             "Select a pacakge:",
		Items:             items,
		Templates:         templates,
		Size:              size,
		Searcher:          searcher,
		StartInSearchMode: true,
		HideSelected:      true,
	}

	i, _, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			// TODO: do not regard this as error
			err = fmt.Errorf("prompt cancelled")
		}
		return nil, err
	}

	return items[i].Package, nil
}

func getConfigPath() (string, error) {
	root := os.Getenv("AFX_CONFIG_ROOT")
	var paths []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		switch filepath.Ext(path) {
		case ".hcl":
			paths = append(paths, filepath.Base(path))
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   promptui.IconSelect + " {{ . | cyan }}",
		Inactive: "  {{ . | faint }}",
	}

	prompt := promptui.Select{
		Label:        "Select config file you want to open",
		Items:        paths,
		Templates:    templates,
		HideSelected: true,
	}
	_, file, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, file), nil
}
