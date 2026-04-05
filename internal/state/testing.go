package state

import (
	"bytes"
	"io"
	"os"
)

func stubState(m map[string]string) func() {
	origRead := ReadStateFile
	origSave := SaveStateFile
	ReadStateFile = func(filename string) ([]byte, error) {
		content, ok := m[filename]
		if !ok {
			return []byte(nil), os.ErrNotExist
		}
		return []byte(content), nil
	}
	SaveStateFile = func(fn string) (io.Writer, error) {
		// override with this buffer to prevent creating
		// actual files in testing
		var b bytes.Buffer
		return &b, nil
	}
	return func() {
		ReadStateFile = origRead
		SaveStateFile = origSave
	}
}

type testConfig struct {
	pkgs []testPackage
}

type testPackage struct {
	r Resource
}

func (p testPackage) GetResource() Resource {
	return p.r
}

func stubPackages(resources []Resource) []Resourcer {
	var cfg testConfig
	for _, resource := range resources {
		cfg.pkgs = append(cfg.pkgs, testPackage{r: resource})
	}

	resourcers := make([]Resourcer, len(cfg.pkgs))
	for i, pkg := range cfg.pkgs {
		resourcers[i] = pkg
	}

	return resourcers
}
