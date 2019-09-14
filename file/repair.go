package file

import (
	"go/ast"
	"go/token"
	"log"
	"path/filepath"
	"reflect"
	"sync"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/davecgh/go-spew/spew"
	"github.com/matthinrichsen/gokey/util"
)

func Repair(f *ast.File, importDir string, sn util.StructManager, fset *token.FileSet) ([]int, bool) {
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

	lines := map[int]token.Pos{}
	linePos := []token.Pos{}
	//pos := map[ast.Node]token.Pos{}
	fsetFile := fset.File(f.Pos())
	for i := 0; i < fsetFile.Size(); i++ {
		l := fsetFile.Line(token.Pos(fsetFile.Base() + i))
		if _, ok := lines[l]; !ok {
			lines[l] = token.Pos(fsetFile.Base() + i)
			linePos = append(linePos, lines[l])
		}
	}

	adjustedLines := []int{}
	for _, l := range linePos {
		adjustedLines = append(adjustedLines, int(l)-fsetFile.Base())
	}

	log.Println(adjustedLines)

	parentStructure := map[ast.Node]ast.Node{}

	spew.Dump(lines)
	spew.Dump(linePos)
	defer spew.Dump(linePos)

	offset := 0
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		parentStructure[c.Node()] = c.Parent()
		_ = nudgeTokenPositions(c.Node(), int64(offset))

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
			}

			names, ok := sn.Get(importReference, structReference)
			if !ok || len(names) <= c.Index() || c.Index() == -1 {
				return true
			}

			nudge := len(names[c.Index()]) + 2
			for i, l := range linePos {
				if l > c.Node().Pos() {
					log.Println(i, int(l)-fsetFile.Base(), `nudging`, nudge)
					linePos[i] = linePos[i] + token.Pos(nudge)
				}
			}

			offset += nudge
			_ = nudgeTokenPositions(expr, int64(nudge))
			k := &ast.KeyValueExpr{
				Value: expr,
				Key: &ast.Ident{
					Name: names[c.Index()],
				},
				Colon: expr.Pos() - 2,
			}
			c.Replace(k)

			for cur := c.Parent(); cur != nil; cur = parentStructure[cur] {
				nudgeRightBrace(cur, int64(len(names[c.Index()])+2))
			}
		}

		return true
	}, nil)

	adjustedLines = []int{}
	for _, l := range linePos {
		adjustedLines = append(adjustedLines, int(l)-fsetFile.Base())
	}
	return adjustedLines, nodesRepared
}

func recordOriginalPositions() {
	_ = map[interface{}]int64{}
}

func nudgeTokenPositions(i interface{}, offset int64) (baseOff int64) {
	//	return
	defer func() {
		_ = recover()
	}()

	e := reflect.ValueOf(i).Elem()
	for i := 0; i < e.NumField(); i++ {
		baseOff = nudgeTokenPos(e.Field(i), offset, baseOff)
	}
	return baseOff
}

func nudgeTokenPos(f reflect.Value, offset, baseOff int64) (r int64) {
	//	return
	defer func() {
		_ = recover()
		r = baseOff
	}()

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
	return baseOff
}

func nudgeRightBrace(i interface{}, offset int64) {
	//	return
	defer func() {
		_ = recover()
	}()

	f := reflect.ValueOf(i).Elem().FieldByName(`Rbrace`)
	f.SetInt(f.Int() + offset)
}
