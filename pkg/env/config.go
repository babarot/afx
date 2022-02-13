package env

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/AlecAivazis/survey/v2"
)

// Config represents data of environment variables and cache file path
type Config struct {
	Path string              `json:"path"`
	Env  map[string]Variable `json:"env"`
}

// Variables is a collection of Variable and its name
type Variables map[string]Variable

// Variable represents environment variable
type Variable struct {
	Value   string `json:"value,omitempty"`
	Default string `json:"default,omitempty"`
	Input   Input  `json:"input,omitempty"`
}

// Input represents value input from terminal
type Input struct {
	When    bool   `json:"when,omitempty"`
	Message string `json:"message,omitempty"`
	Help    string `json:"help,omitempty"`
}

// New creates Config instance
func New(path string) *Config {
	cfg := &Config{
		Path: path,
		Env:  map[string]Variable{},
	}
	if _, err := os.Stat(path); err == nil {
		// already exist
		cfg.read()
	}
	return cfg
}

// Add adds environment variable with given key and given value
func (c *Config) Add(args ...interface{}) error {
	switch len(args) {
	case 0:
		return errors.New("one or two args required")
	case 1:
		switch args[0].(type) {
		case Variables:
			variables := args[0].(Variables)
			for name, v := range variables {
				c.add(name, v)
			}
			return nil
		default:
			return errors.New("args type should be Variables")
		}
	case 2:
		name, ok := args[0].(string)
		if !ok {
			return errors.New("args[0] type should be string")
		}
		v, ok := args[1].(Variable)
		if !ok {
			return errors.New("args[1] type should be Variable")
		}
		c.add(name, v)
		return nil
	default:
		return errors.New("too many arguments")
	}
}

func (c *Config) add(name string, v Variable) {
	defer c.save()

	existing, exist := c.Env[name]
	if exist {
		v.Value = existing.Value
	}

	if v.Value != os.Getenv(name) && os.Getenv(name) != "" {
		v.Value = os.Getenv(name)
	}
	if v.Value == "" {
		v.Value = os.Getenv(name)
	}
	if v.Value == "" {
		v.Value = v.Default
	}

	os.Setenv(name, v.Value)
	c.Env[name] = v
}

// Refresh deletes existing file cache
func (c *Config) Refresh() error {
	return c.delete()
}

// Ask asks the user for input using the given query
func (c *Config) Ask(keys ...string) {
	var update bool
	for _, key := range keys {
		v, found := c.Env[key]
		if !found {
			continue
		}
		if len(v.Value) > 0 {
			continue
		}
		if !v.Input.When {
			continue
		}
		var opts []survey.AskOpt
		opts = append(opts, survey.WithValidator(survey.Required))
		survey.AskOne(&survey.Password{
			Message: v.Input.Message,
			Help:    v.Input.Help,
		}, &v.Value, opts...)
		c.Env[key] = v
		os.Setenv(key, v.Value)
		update = true
	}
	if update {
		c.save()
	}
}

func (c *Config) AskWhen(env map[string]bool) {
	var update bool
	for key, when := range env {
		v, found := c.Env[key]
		if !found {
			continue
		}
		if len(v.Value) > 0 {
			continue
		}
		if !when {
			continue
		}
		var opts []survey.AskOpt
		opts = append(opts, survey.WithValidator(survey.Required))
		survey.AskOne(&survey.Password{
			Message: v.Input.Message,
			Help:    v.Input.Help,
		}, &v.Value, opts...)
		c.Env[key] = v
		os.Setenv(key, v.Value)
		update = true
	}
	if update {
		c.save()
	}
}

func (c *Config) read() error {
	_, err := os.Stat(c.Path)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(c.Path)
	if err != nil {
		return err
	}

	return json.Unmarshal(content, &c)
}

func (c *Config) save() error {
	cfg := Config{Path: c.Path, Env: make(map[string]Variable)}
	// Remove empty variable from c.Env
	// to avoid adding empty item to cache
	for name, v := range c.Env {
		if v.Value == "" && v.Default == "" {
			continue
		}
		cfg.Env[name] = v
	}

	f, err := os.Create(c.Path)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(cfg)
}

func (c *Config) delete() error {
	return os.Remove(c.Path)
}
