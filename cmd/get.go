package cmd

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/b4b4r07/afx/pkg/config"
	"github.com/b4b4r07/afx/pkg/helpers/shell"
	"github.com/b4b4r07/afx/pkg/helpers/spin"
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hclwrite"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

type getCmd struct {
	meta

	spin *spin.Spinner
	pkg  config.Package

	verbose bool
}

var (
	// getLong is long description of get command
	getLong = templates.LongDesc(`adfasfadsfsds`)

	// getExample is examples for get command
	getExample = templates.Examples(`
		# This command gets the definition based on given link
		pkg get https://github.com/b4b4r07/enhancd
	`)
)

// newGetCmd creates a new get command
func newGetCmd() *cobra.Command {
	c := &getCmd{}

	getCmd := &cobra.Command{
		Use:                   "get <URL>",
		Short:                 "Get package from given URL",
		Long:                  getLong,
		Example:               getExample,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.meta.init(args); err != nil {
				return err
			}
			if c.parseErr != nil {
				return c.parseErr
			}
			c.Env.Ask("GITHUB_TOKEN")
			c.spin = spin.New("%s")
			c.spin.Start()
			defer c.spin.Stop()
			return c.run(args)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			c.spin.Stop()
			c.pkg.GetType()
			path, err := getConfigPath()
			if err != nil {
				return err
			}
			var fp io.Writer
			fp, err = os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return err
			}
			if c.verbose {
				fp = io.MultiWriter(os.Stdout, fp)
			}
			f := hclwrite.NewEmptyFile()
			cfg := config.ConvertsFrom(c.pkg)
			gohcl.EncodeIntoBody(&cfg, f.Body())
			fmt.Fprintf(fp, "%s\n", f.Bytes())
			vim := shell.New("vim", path)
			return vim.Run(context.Background())
		},
	}

	f := getCmd.Flags()
	f.BoolVar(&c.verbose, "verbose", false, "output verbosely")

	return getCmd
}

func (c *getCmd) run(args []string) error {
	var pkg config.Package

	arg := args[0]
	u, err := url.Parse(arg)
	if err != nil {
		return err
	}

	switch u.Host {
	case "github.com":
		e := strings.Split(u.Path[1:len(u.Path)], "/")
		github, err := config.NewGitHub(e[0], e[1])
		if err != nil {
			c.Env.Refresh()
			return err
		}
		pkg = github
	case "gist.github.com":
		e := strings.Split(u.Path[1:len(u.Path)], "/")
		gist, err := config.NewGist(e[0], e[1])
		if err != nil {
			c.Env.Refresh()
			return err
		}
		pkg = gist
	default:
		return fmt.Errorf("%s: currently github is only allowed", arg)
	}

	if config.Defined(c.Packages, pkg) {
		return fmt.Errorf("%s: already installed", pkg.GetSlug())
	}

	pathsCh := make(chan []string)
	errCh := make(chan error)
	go func() {
		paths, err := pkg.Objects()
		pathsCh <- paths
		errCh <- err
	}()

	result, err := c.prompt(promptui.Select{
		Label: "Select package type",
		Items: []string{"plugin", "command"},
	})
	if err != nil {
		return err
	}

	paths := <-pathsCh
	if err := <-errCh; err != nil {
		return err
	}

	switch result {
	case "command":
		if len(pkg.GetCommandBlock().Link) == 0 {
			from, _ := c.prompt(promptui.Select{
				Label: "Select command file (source)",
				Items: paths,
			})
			to, _ := c.prompt(promptui.Select{
				Label: "Select command file (destination)",
				Items: paths,
			}, "(Rename to)")
			pkg = pkg.SetCommand(config.Command{
				Link: []*config.Link{&config.Link{From: from, To: to}},
			})
		}
	case "plugin":
		if len(paths) > 0 {
			source, _ := c.prompt(promptui.Select{
				Label: "Select source file",
				Items: paths,
			}, "(Others)")
			pkg = pkg.SetPlugin(config.Plugin{Sources: []string{source}})
		}
	}

	c.pkg = pkg
	return nil
}

func (c getCmd) prompt(s promptui.Select, others ...string) (string, error) {
	c.spin.Stop()
	defer c.spin.Start()
	s.HideSelected = true

	if len(others) > 0 {
		p := promptui.SelectWithAdd{
			Label:    s.Label.(string),
			Items:    s.Items.([]string),
			AddLabel: others[0],
		}
		_, result, err := p.Run()
		return result, err
	}

	switch items := s.Items.(type) {
	case []string:
		s.Searcher = func(input string, index int) bool {
			item := items[index]
			name := strings.Replace(strings.ToLower(item), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)
			return strings.Contains(name, input)
		}
	}

	_, result, err := s.Run()
	return result, err
}
