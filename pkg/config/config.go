package config

// Config structure for file describing deployment. This includes the module source, inputs
// dependencies, backend etc. One config element is connected to a single deployment
type Config struct {
	GitHub []*GitHub `hcl:"github,block"`
	Gist   []*Gist   `hcl:"gist,block"`
	Local  []*Local  `hcl:"local,block"`
	HTTP   []*HTTP   `hcl:"http,block"`
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
