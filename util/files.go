package util

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
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
	pkgs, err := parser.ParseDir(fset, directory, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	files := map[string]*ast.File{}
	for _, p := range pkgs {
		for filename, astFile := range p.Files {
			files[filename] = astFile
		}
	}
	return files, nil
}
