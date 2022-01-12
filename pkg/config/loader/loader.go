package loader

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/b4b4r07/afx/pkg/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// Parser represents mainly HCL parser
type Parser struct {
	p *hclparse.Parser
}

func (p *Parser) loadHCLFile(path string) (hcl.Body, hcl.Diagnostics) {
	log.Printf("[TRACE] parsing as HCL body: %s\n", path)
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "failed to read file",
				Detail:   fmt.Sprintf("%q could not be read.", path),
			},
		}
	}

	var file *hcl.File
	var diags hcl.Diagnostics
	switch {
	case strings.HasSuffix(path, ".json"):
		file, diags = p.p.ParseJSON(src, path)
	default:
		file, diags = p.p.ParseHCL(src, path)
	}

	// If the returned file or body is nil, then we'll return a non-nil empty
	// body so we'll meet our contract that nil means an error reading the file.
	if file == nil || file.Body == nil {
		return hcl.EmptyBody(), diags
	}

	return file.Body, diags
}

func visitHCL(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch filepath.Ext(path) {
		case ".hcl":
			*files = append(*files, path)
		}
		return nil
	}
}

// getConfigFiles walks the given path and returns the files ending with HCL
// Also, it returns the path if the path is just a file and a HCL file
func getConfigFiles(path string) ([]string, error) {
	var files []string
	fi, err := os.Stat(path)
	if err != nil {
		return files, errors.Detail{
			Head:    path,
			Summary: "No such file or directory",
			Details: []string{},
		}
	}
	if fi.IsDir() {
		return files, filepath.Walk(path, visitHCL(&files))
	}
	switch filepath.Ext(path) {
	case ".hcl":
		files = append(files, path)
	default:
		log.Printf("[WARN] %s: found but cannot be loaded. HCL is only allowed\n", path)
	}
	return files, nil
}

// Load reads the files and converts them to Config object
func Load(path string) (schema.Data, error) {
	parser := &Parser{hclparse.NewParser()}

	var diags hcl.Diagnostics
	var bodies []hcl.Body

	files, err := getConfigFiles(path)
	if err != nil {
		return schema.Data{}, err
	}

	if len(files) == 0 {
		return schema.Data{}, errors.Detail{
			Head:    `"loader.Load()" failed`,
			Summary: "No HCL files found",
			Details: []string{
				fmt.Sprintf("Config root path is %q.", path),
				"But it doesn't have any HCL files at all",
				"For more usage, see https://github.com/b4b4r07/afx",
			},
		}
	}

	for _, file := range files {
		body, fDiags := parser.loadHCLFile(file)
		bodies = append(bodies, body)
		diags = append(diags, fDiags...)
	}

	if diags.HasErrors() {
		err = errors.New(diags, parser.p.Files())
	}

	return schema.Data{
		Body:  hcl.MergeBodies(bodies),
		Files: parser.p.Files(),
	}, err
}
