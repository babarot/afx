package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/b4b4r07/afx/pkg/config"
	"github.com/b4b4r07/afx/pkg/printers"
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/spf13/cobra"
)

type showCmd struct {
	meta
}

var (
	// showLong is long description of show command
	showLong = templates.LongDesc(``)

	// showExample is examples for show command
	showExample = templates.Examples(`
		afx show
	`)
)

// newShowCmd creates a new show command
func newShowCmd() *cobra.Command {
	c := &showCmd{}

	showCmd := &cobra.Command{
		Use:                   "show",
		Short:                 "Show packages managed by afx",
		Long:                  showLong,
		Example:               showExample,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.meta.init(args); err != nil {
				return err
			}
			return c.run(args)
		},
	}

	return showCmd
}

type Item struct {
	Name   string
	Type   string
	Status string
}

func (c *showCmd) run(args []string) error {
	w := printers.GetNewTabWriter(os.Stdout)
	headers := []string{"NAME", "TYPE", "STATUS"}

	var items []Item
	for _, pkg := range append(c.State.Additions, c.State.Readditions...) {
		items = append(items, Item{
			Name:   pkg.GetName(),
			Type:   getType(pkg),
			Status: "WaitingInstall",
		})
	}
	for _, pkg := range c.State.Changes {
		items = append(items, Item{
			Name:   pkg.GetName(),
			Type:   getType(pkg),
			Status: "WaitingUpdate",
		})
	}
	for _, resource := range c.State.Deletions {
		items = append(items, Item{
			Name:   resource.Name,
			Type:   resource.Type,
			Status: "WaitingUninstall",
		})
	}
	for _, pkg := range c.State.NoChanges {
		items = append(items, Item{
			Name:   pkg.GetName(),
			Type:   getType(pkg),
			Status: "Installed",
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	fmt.Fprintf(w, strings.Join(headers, "\t")+"\n")
	for _, line := range items {
		fields := []string{
			line.Name, line.Type, line.Status,
		}
		fmt.Fprintf(w, strings.Join(fields, "\t")+"\n")
	}

	return w.Flush()
}

func getType(pkg config.Package) string {
	var ty string
	switch pkg := pkg.(type) {
	case *config.GitHub:
		ty = "GitHub"
		if pkg.HasReleaseBlock() {
			ty = "GitHub Release"
		}
	case *config.Gist:
		ty = "Gist"
	case *config.Local:
		ty = "Local"
	case *config.HTTP:
		ty = "HTTP"
	default:
		ty = "Unknown"
	}
	return ty
}
