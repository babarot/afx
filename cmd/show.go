package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/b4b4r07/afx/pkg/helpers/templates"
	"github.com/b4b4r07/afx/pkg/printers"
	"github.com/b4b4r07/afx/pkg/state"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

type showCmd struct {
	metaCmd

	opt showOpt
}

type showOpt struct {
	output string
}

var (
	// showLong is long description of show command
	showLong = templates.LongDesc(``)

	// showExample is examples for show command
	showExample = templates.Examples(`
		$ afx show
		$ afx show -o json | jq .github
	`)
)

// newShowCmd creates a new show command
func (m metaCmd) newShowCmd() *cobra.Command {
	c := &showCmd{metaCmd: m}

	showCmd := &cobra.Command{
		Use:                   "show",
		Short:                 "Show packages managed by afx",
		Long:                  showLong,
		Example:               showExample,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		ValidArgs:             state.Keys(m.state.NoChanges),
		Args:                  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := m.GetConfig()
			if len(args) > 0 {
				cfg = cfg.Contains(args...)
			}
			b, err := yaml.Marshal(cfg)
			if err != nil {
				return err
			}
			switch c.opt.output {
			case "default":
				return c.run(args)
			case "json":
				b, err := yaml.YAMLToJSON(b)
				if err != nil {
					return err
				}
				fmt.Println(string(b))
			case "yaml":
				fmt.Println(string(b))
			case "path":
				for _, pkg := range c.GetPackages(c.state.NoChanges) {
					fmt.Println(pkg.GetHome())
				}
			case "name":
				for _, pkg := range c.GetPackages(c.state.NoChanges) {
					fmt.Println(pkg.GetName())
				}
			default:
				return fmt.Errorf("%s: not supported output style", c.opt.output)
			}
			return nil
		},
	}

	flag := showCmd.Flags()
	flag.StringVarP(&c.opt.output, "output", "o", "default", "Output style [default,json,yaml,path,name]")

	showCmd.RegisterFlagCompletionFunc("output",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			out := []string{"default", "json", "yaml", "path", "name"}
			return out, cobra.ShellCompDirectiveNoFileComp
		})

	return showCmd
}

func (c *showCmd) run(args []string) error {
	w := printers.GetNewTabWriter(os.Stdout)
	headers := []string{"NAME", "TYPE", "STATUS"}

	type Item struct {
		Name   string
		Type   string
		Status string
	}
	type Items []Item

	filter := func(items []Item, input string) []Item {
		var tmp []Item
		for _, item := range items {
			if strings.Contains(item.Name, input) {
				tmp = append(tmp, item)
			}
		}
		return tmp
	}

	var items []Item
	for _, pkg := range c.state.Additions {
		items = append(items, Item{
			Name:   pkg.Name,
			Type:   pkg.Type,
			Status: "WaitingInstall",
		})
	}
	for _, pkg := range c.state.Changes {
		items = append(items, Item{
			Name:   pkg.Name,
			Type:   pkg.Type,
			Status: "WaitingUpdate",
		})
	}
	for _, pkg := range c.state.Deletions {
		items = append(items, Item{
			Name:   pkg.Name,
			Type:   pkg.Type,
			Status: "WaitingUninstall",
		})
	}
	for _, pkg := range c.state.NoChanges {
		items = append(items, Item{
			Name:   pkg.Name,
			Type:   pkg.Type,
			Status: "Installed",
		})
	}

	if len(args) > 0 {
		var tmp []Item
		for _, arg := range args {
			tmp = append(tmp, filter(items, arg)...)
		}
		items = tmp
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	fmt.Fprintf(w, strings.Join(headers, "\t")+"\n")
	for _, item := range items {
		fields := []string{
			item.Name, item.Type, item.Status,
		}
		fmt.Fprintf(w, strings.Join(fields, "\t")+"\n")
	}

	return w.Flush()
}
