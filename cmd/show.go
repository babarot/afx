package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/b4b4r07/afx/pkg/config"
	"github.com/b4b4r07/afx/pkg/printers"
	"github.com/b4b4r07/afx/pkg/templates"
	"github.com/k0kubun/pp"
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
		Short:                 "Show packages",
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

type Line struct {
	Name string
	Type string
	// status  Status
}
type Lines []Line

// type Status string

// const (
// 	NeedInstall   = Status("NeedInstall")
// 	NeedReinstall = Status("NeedReinstall")
// 	NeedUpdate    = Status("NeedUpdate")
// 	NeedUninstall = Status("NeedUninstall")
// 	NoChanges     = Status("")
// )

type Package interface {
	GetName() string
}

func (c *showCmd) run(args []string) error {
	w := printers.GetNewTabWriter(os.Stdout)
	headers := []string{"NAME", "TYPE"}
	fmt.Fprintf(w, strings.Join(headers, "\t")+"\n")

	var pkgs []Package
	for _, pkg := range c.Packages {
		pkgs = append(pkgs, pkg)
	}
	for _, resource := range c.State.Deletions {
		pkgs = append(pkgs, resource)
	}
	pp.Println(pkgs)

	var lines Lines
	for _, pkg := range c.Packages {
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
			ty = "unknown"
		}
		lines = append(lines, Line{
			Name: pkg.GetName(),
			Type: ty,
		})
	}
	// 	var s Status
	// 	r := c.State.Validate(pkg.GetName())
	// 	switch r.(type) {
	// 	case state.Addition:
	// 		s = NeedInstall
	// 	case state.Readdition:
	// 		s = NeedReinstall
	// 	case state.Change:
	// 		s = NeedUpdate
	// 		// case state.Deletion:
	// 		// 	s = NeedUninstall
	// 	}
	// 	fmt.Printf("%T\t%s\n", r, pkg.GetName())
	// 	lines = append(lines, Line{
	// 		name:    pkg.GetName(),
	// 		pkgType: ty,
	// 		status:  s,
	// 	})
	// 	// fields := []string{pkg.GetName(), pkgType, stateType}
	// 	// fmt.Fprintf(w, strings.Join(fields, "\t")+"\n")
	// }
	// for _, resource := range c.State.Deletions {
	// 	lines = append(lines, Line{
	// 		name: resource.Name,
	// 		// pkgType: ty,
	// 		status: NeedUninstall,
	// 	})
	// }

	// lines = lines.filter(func(line Line) bool {
	// 	return line.status != NoChanges
	// })
	for _, line := range lines {
		fields := []string{
			line.Name, line.Type,
		}
		fmt.Fprintf(w, strings.Join(fields, "\t")+"\n")
	}

	return w.Flush()
}

// func (l Lines) filter(fn func(Line) bool) Lines {
// 	var lines []Line
// 	for _, line := range l {
// 		if fn(line) {
// 			lines = append(lines, line)
// 		}
// 	}
// 	return lines
// }
