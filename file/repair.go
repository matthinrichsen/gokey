package file

import (
	"go/ast"
	"go/token"
	"log"
	"path/filepath"
	"reflect"
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

	parentStructure := map[ast.Node]ast.Node{}

	offset := 0
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		parentStructure[c.Node()] = c.Parent()
		baseOffset := nudgeTokenPositions(c.Node(), int64(offset))

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

			offset += len(names[c.Index()]) + 2
			baseOffset = nudgeTokenPositions(expr, int64(len(names[c.Index()])+2))

			c.Replace(&ast.KeyValueExpr{
				Value: expr,
				Key: &ast.Ident{
					Name: names[c.Index()],
				},
				Colon: token.Pos(baseOffset + int64(len(names[c.Index()]))),
			})

			for cur := c.Parent(); cur != nil; cur = parentStructure[cur] {
				nudgeRightBrace(cur, int64(len(names[c.Index()])+2))
			}
		}

		return true
	}, nil)

	return nodesRepared
}

func nudgeTokenPositions(i interface{}, offset int64) (baseOff int64) {
	defer func() {
		_ = recover()
	}()

	e := reflect.ValueOf(i).Elem()
	for i := 0; i < e.NumField(); i++ {
		func() {
			defer func() {
				_ = recover()
			}()
			f := e.Field(i)

			switch f.Type().String() {
			case `token.Pos`:
				v := f.Int()
				if v > 0 {
					if v < baseOff || baseOff == 0 {
						baseOff = v
					}
					f.SetInt(v + offset)
				}
			}
		}()
	}
	return baseOff
}

func nudgeRightBrace(i interface{}, offset int64) {
	defer func() {
		_ = recover()
	}()

	f := reflect.ValueOf(i).Elem().FieldByName(`Rbrace`)
	f.SetInt(f.Int() + offset)
}
