package file

import (
	"go/ast"
	"path/filepath"
	"sync"

	"golang.org/x/tools/go/ast/astutil"

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

	astutil.Apply(f, func(c *astutil.Cursor) bool {
		a, ok := c.Parent().(*ast.CompositeLit)
		if !ok {
			return true
		}

		expr, ok := c.Node().(ast.Expr)
		if !ok {
			return true
		}

		switch expr.(type) {
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
			default:
				log.Printf("%T what is this? %#v %#v", t, expr, c.Parent())
			}

			names, ok := sn.Get(importReference, structReference)
			if !ok || len(names) <= c.Index() || c.Index() == -1 {
				return true
			}

			c.Replace(&ast.KeyValueExpr{
				Value: expr,
				Key: &ast.Ident{
					Name: names[c.Index()],
				},
				Colon: f.End(),
			})
		}

		return true
	}, nil)

	return nodesRepared
}
