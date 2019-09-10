package main

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
)

func main() {
	fixDirectory(`.`)
}

func fixDirectory(path string) {
	fileSet := token.NewFileSet()
	_ = filepath.Walk(path, func(filename string, info os.FileInfo, err error) error {
		if info.IsDir() {
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

		fp, err := os.OpenFile(filename, os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer fp.Close()

		f, err := parser.ParseFile(fileSet, filename, bytes, parser.ParseComments)
		if err != nil {
			return nil
		}

		nodesToRepair := []*ast.CompositeLit{}
		ast.Inspect(f, func(n ast.Node) bool {
			a, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}

			needsRepair := false
			for _, e := range a.Elts {
				switch e.(type) {
				case *ast.BasicLit:
					needsRepair = true
				case *ast.CompositeLit:
					needsRepair = true
				}
			}
			if needsRepair {
				nodesToRepair = append(nodesToRepair, a)
			}
			return false
		})

		for _, a := range nodesToRepair {
			names := getMemberNames(a)
			if len(names) < len(a.Elts) {
				continue
			}

			for i, e := range a.Elts {
				switch e.(type) {
				case *ast.BasicLit:
					a.Elts[i] = &ast.KeyValueExpr{
						Key: &ast.Ident{
							Name: names[i],
						},
						Value: e,
					}
				case *ast.CompositeLit:
					a.Elts[i] = &ast.KeyValueExpr{
						Key: &ast.Ident{
							Name: names[i],
						},
						Value: e,
					}
				}
			}
		}

		if len(nodesToRepair) > 0 {
			printer.Fprint(fp, fileSet, f)
		}
		return nil
	})
}

func getMemberNames(a *ast.CompositeLit) []string {
	id, ok := a.Type.(*ast.Ident)
	if !ok || id == nil || id.Obj == nil {
		return nil
	}

	ts, ok := id.Obj.Decl.(*ast.TypeSpec)
	if !ok || ts == nil {
		return nil
	}

	st, ok := ts.Type.(*ast.StructType)
	if !ok || st == nil || st.Fields == nil {
		return nil
	}

	names := []string{}
	for _, field := range st.Fields.List {
		id, ok := field.Type.(*ast.Ident)
		if ok && id.Obj != nil {
			names = append(names, id.Obj.Name)
		}
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return names
}
