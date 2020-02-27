// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-interpreter/wagon/wasm"
	"golang.org/x/tools/go/packages"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func identifierFromString(str string) string {
	var ident string
	for _, r := range []rune(str) {
		if r > 0xff {
			panic("identifiers cannot include non-Latin1 characters")
		}
		if '0' <= r && r <= '9' {
			ident += string(r)
			continue
		}
		if 'a' <= r && r <= 'z' {
			ident += string(r)
			continue
		}
		if 'A' <= r && r <= 'Z' {
			ident += string(r)
			continue
		}
		ident += fmt.Sprintf("_%02x", r)
	}
	return ident
}

func namespaceFromPkg(pkg *packages.Package) string {
	ts := strings.Split(pkg.PkgPath, "/")
	for i, t := range ts {
		ts[i] = identifierFromString(t)
	}
	return strings.Join(ts, ".")
}

type Func struct {
	Wasm  wasm.Function
	Index int
	Name  string
}

func wasmTypeToCSharpType(v wasm.ValueType) string {
	switch v {
	case wasm.ValueTypeI32:
		return "int"
	case wasm.ValueTypeI64:
		return "long"
	case wasm.ValueTypeF32:
		return "float"
	case wasm.ValueTypeF64:
		return "double"
	default:
		panic("not reached")
	}
}

func (f *Func) CSharp(indent string) string {
	var ret string
	var retType string
	switch ts := f.Wasm.Sig.ReturnTypes; len(ts) {
	case 0:
		retType = "void"
	case 1:
		retType = wasmTypeToCSharpType(ts[0])
		ret = "return 0;"
	default:
		panic("the number of return values should be 0 or 1 so far")
	}

	str := fmt.Sprintf(`// OriginalName: %s
// Index:        %d
internal %s %s()
{
    // TODO: Implement this.
    %s
}`, f.Name, f.Index, retType, identifierFromString(f.Name), ret)

	// Add indentations
	var lines []string
	for _, line := range strings.Split(str, "\n") {
		lines = append(lines, indent+line)
	}
	return strings.Join(lines, "\n") + "\n"
}

func run() error {
	tmp, err := ioutil.TempDir("", "go2dotnet-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	wasmpath := filepath.Join(tmp, "tmp.wasm")

	// TODO: Detect the last argument is path or not
	pkgname := os.Args[len(os.Args)-1]

	args := append([]string{"build"}, os.Args[1:]...)
	args = append(args[:len(args)-1], "-o="+wasmpath, pkgname)
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

	mod, err := wasm.ReadModule(f, nil)
	if err != nil {
		return err
	}

	var ifs []*Func
	var fs []*Func
	for i, f := range mod.FunctionIndexSpace {
		if f.Name == "" {
			name := mod.Import.Entries[i].FieldName
			ifs = append(ifs, &Func{
				Wasm:  f,
				Index: i,
				Name:  name,
			})
			continue
		}
		name := f.Name
		fs = append(fs, &Func{
			Wasm:  f,
			Index: i,
			Name:  name,
		})
	}

	pkgs, err := packages.Load(nil, pkgname)
	if err != nil {
		return err
	}

	namespace := namespaceFromPkg(pkgs[0])
	class := identifierFromString(pkgs[0].Name)

	t := template.Must(template.New("out.cs").Parse(csTmpl))
	if err := t.Execute(os.Stdout, struct {
		Namespace   string
		Class       string
		ImportFuncs []*Func
		Funcs       []*Func
	}{
		Namespace:   namespace,
		Class:       class,
		ImportFuncs: ifs,
		Funcs:       fs,
	}); err != nil {
		return err
	}

	return nil
}

const csTmpl = `// Code generated by go2dotnet. DO NOT EDIT.

namespace {{.Namespace}}
{
    sealed class Import
    {
{{- range $value := .ImportFuncs}}
{{$value.CSharp "        "}}{{end}}    }

    sealed class Go_{{.Class}}
    {
{{- range $value := .Funcs}}
{{$value.CSharp "        "}}{{end}}    }
}
`
