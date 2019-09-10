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

type myStruct struct {
	a int
	b int
	c int

	myStruct2
	m   map[string]string
	arr []string
}

type myStruct2 struct {
	a int
	b int
	c int
}

func main() {
	fixDirectory(`.`)
}

func fixDirectory(path string) {
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

		fileSet := token.NewFileSet()
		f, err := parser.ParseFile(fileSet, filename, bytes, parser.ParseComments)
		if err != nil {
			return nil
		}

		ast.Inspect(f, func(n ast.Node) bool {
			a, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}

			names := getMemberNames(a)
			_ = names
			for i, e := range a.Elts {
				switch t := e.(type) {
				case *ast.BasicLit:
					a.Elts[i] = &ast.KeyValueExpr{
						Key: &ast.Ident{
							Name: names[i],
						},
						Value: t,
					}
				case *ast.CompositeLit:
					a.Elts[i] = &ast.KeyValueExpr{
						Key: &ast.Ident{
							Name: names[i],
						},
						Value: t,
					}
				}
			}
			return false
		})

		printer.Fprint(os.Stdout, fileSet, f)
		return nil
	})
}

func getMemberNames(a *ast.CompositeLit) []string {
	id, ok := a.Type.(*ast.Ident)
	if !ok {
		return nil
	}

	ts, ok := id.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		return nil
	}

	st, ok := ts.Type.(*ast.StructType)
	if !ok {
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
