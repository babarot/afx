package tmpl

import (
	"bytes"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/b4b4r07/afx/pkg/data"
)

// Template holds data that can be applied to a template string.
type Template struct {
	data   *data.Data
	fields Fields
}

// Fields that will be available to the template engine.
type Fields map[string]interface{}

const (
	pkgName = "Name"
	pkgHome = "Home"
	dir     = "Dir"
	goos    = "OS"
	goarch  = "Arch"
	env     = "Env"
	release = "Release"

	// release
	releaseName = "Name"
	releaseTag  = "Tag"
)

// New Template.
func New(d *data.Data) *Template {
	return &Template{
		data: d,
		fields: Fields{
			env:     d.Env,
			pkgName: d.Package.Name,
			pkgHome: d.Package.Home,
			dir:     d.Package.Home,
			goos:    d.Runtime.Goos,
			goarch:  d.Runtime.Goarch,
			release: map[string]string{
				releaseName: d.Release.Name,
				releaseTag:  d.Release.Tag,
			},
		},
	}
}

// Apply applies the given string against the Fields stored in the template.
func (t *Template) Apply(s string) (string, error) {
	var out bytes.Buffer
	tmpl, err := template.New("tmpl").
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"replace": strings.ReplaceAll,
			"time": func(s string) string {
				return time.Now().UTC().Format(s)
			},
			"tolower":    strings.ToLower,
			"toupper":    strings.ToUpper,
			"trim":       strings.TrimSpace,
			"trimprefix": strings.TrimPrefix,
			"trimsuffix": strings.TrimSuffix,
			"dir":        filepath.Dir,
			"abs":        filepath.Abs,
		}).
		Parse(s)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(&out, t.fields)
	return out.String(), err
}

// Replace populates Fields from the artifact and replacements.
func (t *Template) Replace(replacements map[string]string) *Template {
	t.fields[goos] = replace(replacements, t.data.Runtime.Goos)
	t.fields[goarch] = replace(replacements, t.data.Runtime.Goarch)
	return t
}

func replace(replacements map[string]string, original string) string {
	result := replacements[original]
	if result == "" {
		return original
	}
	return result
}
