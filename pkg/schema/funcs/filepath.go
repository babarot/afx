package funcs

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// GlobFunc returns a list of files matching a given pattern
var GlobFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "pattern",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.List(cty.String)),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		pattern := args[0].AsString()

		pattern, err = expand(pattern)
		if err != nil {
			pattern = args[0].AsString()
		}

		files, err := filepath.Glob(pattern)
		if err != nil {
			return cty.NilVal, err
		}

		vals := make([]cty.Value, len(files))
		for i, file := range files {
			vals[i] = cty.StringVal(file)
		}

		if len(vals) == 0 {
			return cty.ListValEmpty(cty.String), nil
		}
		return cty.ListVal(vals), nil
	},
})

// Expand ...
var Expand = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "path",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		path := args[0].AsString()

		path, err = expand(path)
		if err != nil {
			path = args[0].AsString()
		}

		return cty.StringVal(path), nil
	},
})

func expand(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := getHomeDir()
	if err != nil {
		return "", err
	}

	return home + path[1:], nil
}

func getHomeDir() (string, error) {
	home := ""

	switch runtime.GOOS {
	case "windows":
		home = filepath.Join(os.Getenv("HomeDrive"), os.Getenv("HomePath"))
		if home == "" {
			home = os.Getenv("UserProfile")
		}

	default:
		home = os.Getenv("HOME")
	}

	if home == "" {
		return "", errors.New("no home found")
	}
	return home, nil
}
