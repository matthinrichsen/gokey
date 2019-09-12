package file

import (
	"go/ast"
	"path/filepath"
	"sync"

	"github.com/matthinrichsen/gokey/util"
)

func Repair(f *ast.File, importDir string, sn util.StructManager) bool {
	importsToPaths := map[string]string{}
	once := sync.Once{}
	nodesRepared := false

	buildOutImports := func() {
		nodesRepared = true
		for _, i := range f.Imports {
			if i.Name != nil {
				importsToPaths[util.RemoveQuotes(i.Name.String())] = util.RemoveQuotes(i.Path.Value)
			}
			_, value := filepath.Split(util.RemoveQuotes(i.Path.Value))
			importsToPaths[value] = util.RemoveQuotes(i.Path.Value)
		}
	}

	ast.Inspect(f, func(n ast.Node) bool {
		a, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		for i, e := range a.Elts {
			switch e.(type) {
			case *ast.KeyValueExpr:
			default:
				once.Do(buildOutImports)

				var importReference string
				var structReference string

				switch t := a.Type.(type) {
				case *ast.SelectorExpr: // this struct declaration is import from another package
					structName := t.Sel.String()

					pkg, ok := t.X.(*ast.Ident)
					if ok {
						importReference = importsToPaths[util.RemoveQuotes(pkg.String())]
						structReference = structName
					}
				case *ast.Ident: // this struct declaration is local to the package
					importReference = importDir
					structReference = t.Name
				}

				names, ok := sn.Get(importReference, structReference)
				if !ok || len(names) <= i {
					return true
				}

				a.Elts[i] = &ast.KeyValueExpr{
					Value: e,
					Key: &ast.Ident{
						Name: names[i],
					},
					Colon: f.End(),
				}
			}
		}

		return false
	})
	return nodesRepared
}
