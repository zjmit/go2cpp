// SPDX-License-Identifier: Apache-2.0

package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/akupila/go-wasm"
	"golang.org/x/tools/go/packages"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func identifierFromString(str string) string {
	// TODO: Remove two consecutive '_' characters.
	str = strings.ReplaceAll(str, ".", "·")
	str = strings.ReplaceAll(str, "-", "_")
	return str
}

func namespaceFromPkgName(pkgname string) (string, error) {
	pkgs, err := packages.Load(nil, pkgname)
	if err != nil {
		return "", err
	}
	ts := strings.Split(pkgs[0].PkgPath, "/")
	for i, t := range ts {
		ts[i] = identifierFromString(t)
	}
	return strings.Join(ts, "."), nil
}

type Func struct {
	Name string
}

func run() error {
	tmp, err := ioutil.TempDir("", "go2dotnet")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	wasmpath := filepath.Join(tmp, "tmp.wasm")

	// TODO: Detect the last argument is path or not
	pkgname := os.Args[len(os.Args)-1]

	args := append([]string{"build"}, os.Args[1:]...)
	args = append(args[:len(args)-1], "-o", wasmpath, pkgname)
	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	if err := cmd.Run(); err != nil {
		return err
	}

	f, err := os.Open(wasmpath)
	if err != nil {
		return err
	}
	defer f.Close()

	
	mod, err := wasm.Parse(f)
	if err != nil {
		return err
	}

	var fs []*Func
	for _, s := range mod.Sections {
		switch s := s.(type) {
		case *wasm.SectionName:
			for _, f := range s.Functions.Names {
				fs = append(fs, &Func{
					Name: identifierFromString(f.Name),
				})
			}
		}
	}

	namespace, err := namespaceFromPkgName(pkgname)
	if err != nil {
		return err
	}

	t := template.Must(template.New("out.cs").Parse(tmpl))
	if err := t.Execute(os.Stdout, struct {
		Namespace string
		Funcs     []*Func
	}{
		Namespace: namespace,
		Funcs:     fs,
	}); err != nil {
		return err
	}

	return nil
}

const tmpl = `// Code generated by go2dotnet. DO NOT EDIT.

namespace {{.Namespace}}
{
{{range $value := .Funcs}}    private void {{$value.Name}}
    {
        // TODO: Implement this.
    }

{{end -}}
}
`
