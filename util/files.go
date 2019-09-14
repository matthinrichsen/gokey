package util

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"path/filepath"
)

func CompileFilesInDirectory(directory string, fset *token.FileSet) (*types.Info, map[string]*ast.File, error) {
	files, err := ParseAllGoFilesInDirectory(directory, fset)
	if err != nil {
		return nil, nil, err
	}

	fileList := []*ast.File{}
	for _, f := range files {
		fileList = append(fileList, f)
	}

	info, err := CompileFiles(directory, fset, fileList...)
	if err != nil {
		return nil, nil, err
	}
	return info, files, nil
}

func CompileFiles(directory string, fset *token.FileSet, files ...*ast.File) (*types.Info, error) {
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
	_, err := tc.Check(directory, fset, files, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func ParseAllGoFilesInDirectory(directory string, fset *token.FileSet) (map[string]*ast.File, error) {
	files := map[string]*ast.File{}
	_ = filepath.Walk(directory, func(filename string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if info.IsDir() {
			if directory != filename {
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
