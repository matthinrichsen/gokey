package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/matthinrichsen/gokey/file"
	"github.com/matthinrichsen/gokey/util"
)

func main() {
	fixDirectory(``)
}

func fixDirectory(path string) {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	fileSet := token.NewFileSet()
	sn := util.NewStructFieldNames()

	_ = filepath.Walk(path, func(directory string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		importDir, err := filepath.Rel(filepath.Join(os.Getenv(`GOPATH`), `src`), directory)
		if err != nil {
			importDir = directory
		}

		astinfo, allFiles, err := compile(directory, fileSet)
		if err != nil {
			return nil
		}

		sn.AddPackage(importDir, astinfo)
		buildOutImports(allFiles, fileSet, sn)

		for filename, f := range allFiles {
			if file.Repair(f, importDir, sn) {
				wd, _ := os.Getwd()
				reportFile, err := filepath.Rel(wd, filename)
				if err != nil {
					reportFile = filename
				}
				fmt.Println(reportFile)
				b := &bytes.Buffer{}
				printer.Fprint(b, fileSet, f)

				formatted, err := format.Source(b.Bytes())
				if err != nil {
					formatted = b.Bytes()
				}

				ioutil.WriteFile(filename, formatted, info.Mode())
			}
		}
		return nil
	})
}

func compile(p string, fset *token.FileSet) (*types.Info, map[string]*ast.File, error) {
	files, err := parseAllGoFilesInDir(p, fset)
	if err != nil {
		return nil, nil, err
	}

	tc := &types.Config{
		Importer: importer.Default(),
	}
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Scopes:     make(map[ast.Node]*types.Scope),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	fileList := []*ast.File{}
	for _, f := range files {
		fileList = append(fileList, f)
	}

	_, err = tc.Check(p, fset, fileList, info)
	if err != nil {
		return nil, nil, err
	}
	return info, files, nil
}

func buildOutImports(files map[string]*ast.File, fileSet *token.FileSet, sn util.StructFieldNames) {
	for _, f := range files {
		for _, i := range f.Imports {
			if sn.HasPackage(i.Path.Value) {
				continue
			}

			info, nextRoundOfFiles, err := compile(filepath.Join(os.Getenv("GOPATH"), "src", removeQuotes(i.Path.Value)), fileSet)
			if err == nil {
				sn.AddPackage(i.Path.Value, info)
				buildOutImports(nextRoundOfFiles, fileSet, sn)
			}
		}
	}
}

func parseAllGoFilesInDir(dir string, fset *token.FileSet) (map[string]*ast.File, error) {
	files := map[string]*ast.File{}
	_ = filepath.Walk(dir, func(filename string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if info.IsDir() {
			if dir != filename {
				return filepath.SkipDir
			}
			return nil
		}
		if err != nil {
			return err
		}

		if filepath.Ext(filename) != `.go` {
			return nil
		}

		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		f, err := parser.ParseFile(fset, filename, bytes, parser.ParseComments)
		if err != nil {
			return nil
		}

		files[filename] = f
		return nil
	})

	return files, nil
}
