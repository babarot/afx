package config

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/goccy/go-yaml"
	"github.com/mattn/go-shellwords"
	"github.com/mattn/go-zglob"
)

// Command is
type Command struct {
	Build   *Build            `yaml:"build"`
	Link    []*Link           `yaml:"link"` // validate:"required"
	Env     map[string]string `yaml:"env"`
	Alias   map[string]string `yaml:"alias"`
	Snippet string            `yaml:"snippet"`
	If      string            `yaml:"if"`
}

// Build is
type Build struct {
	Steps []string          `yaml:"steps" validate:"required"`
	Env   map[string]string `yaml:"env"`
}

// Link is
type Link struct {
	From string `yaml:"from" validate:"required"`
	To   string `yaml:"to"`
}

func (l *Link) MarshalYAML() ([]byte, error) {
	type alias Link

	return yaml.Marshal(struct {
		*alias
		To string `yaml:"to"`
	}{
		alias: (*alias)(l),
		To:    os.ExpandEnv(l.To),
	})
}

func (l *Link) UnmarshalYAML(b []byte) error {
	type alias Link

	tmp := struct {
		*alias
		From string `yaml:"from"`
		To   string `yaml:"to"`
	}{
		alias: (*alias)(l),
	}

	if err := yaml.Unmarshal(b, &tmp); err != nil {
		return errors.Wrap(err, "failed to unmarshal YAML")
	}

	l.From = tmp.From
	l.To = expandTilda(os.ExpandEnv(tmp.To))

	return nil
}

// GetLink is
func (c Command) GetLink(pkg Package) ([]Link, error) {
	var links []Link

	if _, err := os.Stat(pkg.GetHome()); err != nil {
		return links, fmt.Errorf(
			"%s: still not exists. this method should have been called after install was done",
			pkg.GetHome(),
		)
	}

	getTo := func(link *Link) string {
		dest := link.To
		if link.To == "" {
			dest = filepath.Base(link.From)
		}
		if !filepath.IsAbs(link.To) {
			dest = filepath.Join(os.Getenv("AFX_COMMAND_PATH"), dest)
		}
		return dest
	}

	for _, link := range c.Link {
		if link.From == "." {
			links = append(links, Link{
				From: pkg.GetHome(),
				To:   getTo(link),
			})
			continue
		}
		file := filepath.Join(pkg.GetHome(), link.From)
		matches, err := zglob.Glob(file)
		if err != nil {
			return links, errors.Wrapf(err, "%s: failed to get links", pkg.GetName())
		}
		var src string
		switch len(matches) {
		case 0:
			log.Printf("[ERROR] %s: no matches\n", file)
			continue
		case 1:
			// OK pattern: matches should be only one
			src = matches[0]
		case 2:
			// TODO: Update this with more flexiblities
			msg := fmt.Sprintf("[ERROR] %s: %d files matched: %#v\n", pkg.GetName(), len(matches), matches)
			return links, errors.New(msg)
		default:
			return links, errors.New("unknown error occured")
		}
		links = append(links, Link{
			From: src,
			To:   getTo(link),
		})
	}

	return links, nil
}

// Installed returns true ...
func (c Command) Installed(pkg Package) bool {
	links, err := c.GetLink(pkg)
	if err != nil {
		log.Printf("[ERROR] %s: command.Installed(): cannot get link section", pkg.GetName())
		return false
	}

	if len(links) == 0 {
		// regard as installed if home dir exists
		// even if link section is not specified
		_, err := os.Stat(pkg.GetHome())
		return err == nil
	}

	for _, link := range links {
		fi, err := os.Lstat(link.To)
		if err != nil {
			return false
		}
		if fi.Mode()&os.ModeSymlink != os.ModeSymlink {
			return false
		}
		orig, err := os.Readlink(link.To)
		if err != nil {
			return false
		}
		if _, err := os.Stat(orig); err != nil {
			log.Printf("[DEBUG] %v does no longer exist (%s)", orig, link.To)
			return false
		}
	}

	return true
}

// buildRequired is
func (c Command) buildRequired() bool {
	return c.Build != nil && len(c.Build.Steps) > 0
}

func (c Command) build(pkg Package) error {
	p := shellwords.NewParser()
	p.ParseEnv = true
	p.ParseBacktick = true
	p.Dir = pkg.GetHome()

	for _, step := range c.Build.Steps {
		args, err := p.Parse(step)
		if err != nil {
			continue
		}
		var stdin io.Reader = os.Stdin
		var stdout, stderr bytes.Buffer
		switch args[0] {
		case "sudo":
			sudo := []string{"sudo", "-S"}
			args = append(sudo, args[1:]...)
			stdin = strings.NewReader(os.Getenv("AFX_SUDO_PASSWORD") + "\n")
		}
		log.Printf("[DEBUG] run command: %#v\n", args)
		cmd := exec.Command(args[0], args[1:]...)
		for k, v := range c.Build.Env {
			cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Stdin = stdin
		cmd.Stdout = &stdout
		cmd.Stdout = os.Stdout // TODO: remove
		cmd.Stderr = &stderr
		log.Printf("[INFO] cd %s\n", pkg.GetHome())
		cmd.Dir = pkg.GetHome()
		if err := cmd.Run(); err != nil {
			return errors.New(stderr.String())
		}
	}
	return nil
}

// Install is
func (c Command) Install(pkg Package) error {
	if c.buildRequired() {
		err := c.build(pkg)
		if err != nil {
			return errors.Wrapf(err, "%s: failed to build", pkg.GetName())
		}
	}

	links, err := c.GetLink(pkg)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to get command.link", pkg.GetName())
	}

	if len(links) == 0 {
		log.Printf("[ERROR] %s: no links", pkg.GetName())
		return fmt.Errorf("%s: GetLink() returns nothing", pkg.GetName())
	}

	var errs errors.Errors
	for _, link := range links {
		// Create base dir if not exists when creating symlink
		pdir := filepath.Dir(link.To)
		if _, err := os.Stat(pdir); os.IsNotExist(err) {
			log.Printf("[DEBUG] %s: created directory to install path", pdir)
			os.MkdirAll(pdir, 0755)
		}

		fi, err := os.Stat(link.From)
		if err != nil {
			log.Printf("[ERROR] %s: no such file or directory\n", link.From)
			continue
		}
		switch fi.Mode() {
		case 0755:
			// ok
		default:
			os.Chmod(link.From, 0755)
		}

		if _, err := os.Lstat(link.To); err == nil {
			log.Printf("[DEBUG] %s: removed because already exists before linking", link.To)
			os.Remove(link.To)
		}

		log.Printf("[DEBUG] created symlink %s to %s", link.From, link.To)
		if err := os.Symlink(link.From, link.To); err != nil {
			log.Printf("[ERROR] failed to create symlink: %v", err)
			errs.Append(err)
		}
	}

	return errs.ErrorOrNil()
}

func (c Command) Unlink(pkg Package) error {
	links, err := c.GetLink(pkg)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to get command.link", pkg.GetName())
	}

	var errs errors.Errors
	for _, link := range links {
		log.Printf("[DEBUG] %s: unlinked %s", pkg.GetName(), link.To)
		errs.Append(os.Remove(link.To))
	}
	return errs.ErrorOrNil()
}

// Init returns necessary things which should be loaded when executing commands
func (c Command) Init(pkg Package) error {
	if !pkg.Installed() {
		msg := fmt.Sprintf("package %s is not installed, so skip to init", pkg.GetName())
		fmt.Printf("## %s\n", msg)
		return errors.New(msg)
	}

	shell := os.Getenv("AFX_SHELL")
	if shell == "" {
		shell = "bash"
	}

	if len(c.If) > 0 {
		cmd := exec.CommandContext(context.Background(), shell, "-c", c.If)
		err := cmd.Run()
		switch cmd.ProcessState.ExitCode() {
		case 0:
		default:
			log.Printf("[ERROR] %s: command.if returns not zero so unlink package", pkg.GetName())
			c.Unlink(pkg)
			return fmt.Errorf("%s: failed to run command.if: %w", pkg.GetName(), err)
		}
	}

	for k, v := range c.Env {
		switch k {
		case "PATH":
			// avoid overwriting PATH
			v = fmt.Sprintf("$PATH:%s", expandTilda(v))
		default:
			// through
		}
		fmt.Printf("export %s=%q\n", k, v)
	}

	for k, v := range c.Alias {
		fmt.Printf("alias %s=%q\n", k, v)
	}

	if s := c.Snippet; s != "" {
		fmt.Printf("%s", s)
	}

	return nil
}
