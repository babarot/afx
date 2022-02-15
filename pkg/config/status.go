package config

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"
)

// Progress is
type Progress struct {
	Status map[string]Status
}

// NewProgress is
func NewProgress(pkgs []Package) Progress {
	status := make(map[string]Status)
	for _, pkg := range pkgs {
		status[pkg.GetName()] = Status{
			Name: pkg.GetName(),
			Done: false,
			Err:  false,
		}
	}
	return Progress{Status: status}
}

// Print is
func (p Progress) Print(completion chan Status) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	fadedOutput := color.New(color.FgCyan)
	for {
		s := <-completion
		fmt.Printf("\x1b[2K")
		if s.Err {
			fmt.Println(red("✖"), white(s.Name))
		} else {
			fmt.Println(green("✔"), white(s.Name))
		}
		p.Status[s.Name] = s
		count, repos := countRemaining(p.Status)

		if count == len(p.Status) {
			break
		}

		_, width := getTerminalSize()
		width = int(math.Min(float64(width), 100))
		finalOutput := strconv.Itoa(len(p.Status)-count) + "| " + strings.Join(repos, ", ")
		if width < 5 {
			finalOutput = ""
		} else if len(finalOutput) > width {
			finalOutput = finalOutput[:width-4] + "..."
		}
		fadedOutput.Printf(finalOutput + "\r")
	}
}

// Status is
type Status struct {
	Name string
	Done bool
	Err  bool
}

func countRemaining(status map[string]Status) (int, []string) {
	count := 0
	var repos []string
	for _, s := range status {
		if s.Done {
			count++
		} else {
			repos = append(repos, s.Name)
		}
	}
	return count, repos
}

func getTerminalSize() (int, int) {
	id := int(os.Stdout.Fd())
	width, height, err := terminal.GetSize(id)
	if err != nil {
		log.Printf("[ERROR]: getTerminalSize(): %v\n", err)
	}
	return height, width
}
