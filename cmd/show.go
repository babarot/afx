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

func (c *showCmd) run(args []string) error {
	w := printers.GetNewTabWriter(os.Stdout)
	headers := []string{"NAME", "TYPE"}
	fmt.Fprintf(w, strings.Join(headers, "\t")+"\n")

	type Line struct {
		Name string
		Type string
	}

	type Package interface {
		GetName() string
	}

	var pkgs []Package
	for _, pkg := range c.Packages {
		pkgs = append(pkgs, pkg)
	}
	for _, resource := range c.State.Deletions {
		pkgs = append(pkgs, resource)
	}

	var lines []Line
	for _, pkg := range pkgs {
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
		lines = append(lines, Line{
			Name: pkg.GetName(),
			Type: ty,
		})
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].Name < lines[j].Name
	})
	for _, line := range lines {
		fields := []string{
			line.Name, line.Type,
		}
		fmt.Fprintf(w, strings.Join(fields, "\t")+"\n")
	}

	return w.Flush()
}
