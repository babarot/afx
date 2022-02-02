package config

import (
	"bytes"
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
	Link    []*Link           `yaml:"link"`
	Env     map[string]string `yaml:"env"`
	Alias   map[string]string `yaml:"alias"`
	Snippet string            `yaml:"snippet"`
}

// Build is
type Build struct {
	Env   map[string]string `yaml:"env"`
	Steps []string          `yaml:"steps"`
}

// Link is
type Link struct {
	From string `yaml:"from"`
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
		return err
	}

	l.From = tmp.From
	l.To = expandTilda(os.ExpandEnv(tmp.To))

	return nil
}

// GetLink is
func (c Command) GetLink(pkg Package) ([]Link, error) {
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

	var links []Link
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
			return links, errors.Wrapf(err, "%s: failed to get links (%#v)", pkg.GetName(), link)
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
	if len(c.Link) == 0 {
		_, err := exec.LookPath(pkg.GetName())
		return err == nil
	}

	links, err := c.GetLink(pkg)
	if len(links) == 0 || err != nil {
		return false
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
		log.Printf("[DEBUG] build command block...\n")
		err := c.build(pkg)
		if err != nil {
			return errors.Wrapf(err, "failed to build: %s", pkg.GetName())
		}
	}

	links, err := c.GetLink(pkg)
	if len(links) == 0 {
		log.Printf("[ERROR] no links: %s\n", pkg.GetName())
		return err
	}

	var errs errors.Errors
	for _, link := range links {
		// Create base dir if not exists when creating symlink
		pdir := filepath.Dir(link.To)
		if _, err := os.Stat(pdir); os.IsNotExist(err) {
			log.Printf("[DEBUG] create directory to install path: %s", pdir)
			os.MkdirAll(pdir, 0755)
		}

		fi, err := os.Stat(link.From)
		if err != nil {
			log.Printf("[ERROR] link.from %q: no such file or directory\n", link.From)
			continue
		}
		switch fi.Mode() {
		case 0755:
			// ok
		default:
			os.Chmod(link.From, 0755)
		}

		log.Printf("[DEBUG] create symlink %s to %s", link.From, link.To)
		if err := os.Symlink(link.From, link.To); err != nil {
			log.Printf("[ERROR] failed to create symlink: %v", err)
			errs.Append(err)
		}
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

	for k, v := range c.Env {
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
