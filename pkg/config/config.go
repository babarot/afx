package config

import "path/filepath"

// Config structure for file describing deployment. This includes the module source, inputs
// dependencies, backend etc. One config element is connected to a single deployment
type Config struct {
	GitHub []*GitHub `yaml:"github,block"`
	Gist   []*Gist   `yaml:"gist,block"`
	Local  []*Local  `yaml:"local,block"`
	HTTP   []*HTTP   `yaml:"http,block"`
}

// Load is
type Load struct {
	Scripts []string `yaml:"scripts,optional"`
}

// Merge all sources into current configuration struct.
// Should just call merge on all blocks / attributes of config struct.
func (c *Config) Merge(srcs []*Config) error {
	// if err := mergeModules(c, srcs); err != nil {
	// 	return err
	// }
	//
	// if err := mergeInputs(c, srcs); err != nil {
	// 	return err
	// }

	return nil
}

// PostProcess is called after merging all configurations together to perform additional
// processing after config is read. Can modify config elements
// func (c *Config) PostProcess(file *File) {
// 	// for _, hook := range c.Hooks {
// 	// 	if hook.Command != nil && strings.HasPrefix(*hook.Command, ".") {
// 	// 		fileDir := filepath.Dir(file.FullPath)
// 	// 		absCommand := filepath.Join(fileDir, *hook.Command)
// 	// 		hook.Command = &absCommand
// 	// 	}
// 	// }
// }

// Validate that the configuration is correct. Calls validation on all parts of the struct.
// This assumes merge is already done and this is a complete configuration. If it is just a
// partial configuration from a child config it can fail as required blocks might not have
// been set.
func (c Config) Validate() (bool, error) {
	return true, nil
}

func ParseYAML(cfg Config) ([]Package, error) {
	var pkgs []Package
	for _, pkg := range cfg.GitHub {
		// TODO: Remove?
		if pkg.HasReleaseBlock() && !pkg.HasCommandBlock() {
			pkg.Command = &Command{
				Link: []*Link{
					{From: filepath.Join("**", pkg.Release.Name)},
				},
			}
		}
		pkgs = append(pkgs, pkg)
	}
	for _, pkg := range cfg.Gist {
		pkgs = append(pkgs, pkg)
	}
	for _, pkg := range cfg.Local {
		pkgs = append(pkgs, pkg)
	}
	for _, pkg := range cfg.HTTP {
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}
