package main

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/matthinrichsen/gokey/tests"
	wat "github.com/matthinrichsen/gokey/tests"
)

var _ tests.MyStruct
var _ = tests.MyStruct2{tests.MyStruct3{}}
var _ = wat.MyStruct2{wat.MyStruct3{}}

func main() {
	fixDirectory(``)
}

type brokenFile struct {
	dir           string
	filename      string
	mode          os.FileMode
	f             *ast.File
	nodesToRepair []*ast.CompositeLit
}

func fixDirectory(path string) {

	path, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	toFix := []brokenFile{}
	fileSet := token.NewFileSet()
	packages := map[string]*types.Info{}

	_ = filepath.Walk(path, func(filename string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		log.Println(filename)
		allFiles, err := parseAllGoFilesInDir(path, fileSet, false)
		if err != nil {
			return nil
		}

		buildOutImports(allFiles, fileSet, packages)

		for _, f := range allFiles {
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

			if len(nodesToRepair) > 0 {
				toFix = append(toFix, brokenFile{
					f:             f,
					dir:           path,
					filename:      filename,
					mode:          info.Mode(),
					nodesToRepair: nodesToRepair,
				})
			}
		}
		return nil
	})

	structFieldNames := map[string][]string{}
	for k, v := range packages {
		for i := range v.Defs {
			if i.Obj != nil {
				ts, ok := i.Obj.Decl.(*ast.TypeSpec)
				if ok {
					structFieldNames[strings.TrimSuffix(strings.TrimPrefix(k, `"`), `"`)+"."+i.Name] = membersFromTypeSpec(ts)
				}
			}
		}
	}
	spew.Dump(structFieldNames)

	for _, brokenFile := range toFix {
		importsToPaths := map[string]string{}
		for _, i := range brokenFile.f.Imports {
			if i.Name != nil {
				importsToPaths[i.Name.String()] = i.Path.Value
			}
			importsToPaths[i.Path.Value] = i.Path.Value
		}

		for _, n := range brokenFile.nodesToRepair {
			names := []string{}
			switch t := n.Type.(type) {
			case *ast.SelectorExpr:
				structName := t.Sel.String()

				pkg, ok := t.X.(*ast.Ident)
				if ok {
					structName := removeQuotes(importsToPaths[removeQuotes(pkg.String())]) + "." + structName
					log.Println(structName)
					names = structFieldNames[structName]
				}
			}
			log.Println(names)
			//spew.Dump(n)
			_ = n
		}
	}
}

// this should take in the package and struct name instead and either compute the struct field names
// or return a cached copy of it
func getMemberNames(dir string, a *ast.CompositeLit, infos map[string]*types.Info) []string {
	switch t := a.Type.(type) {
	case *ast.Ident:
		for k, v := range infos {
			_, foundIt := v.Defs[t]
			if foundIt {
				log.Println(`FOUND IT`, k, t)
			}
			for k, v := range v.Defs {
				log.Println(k, v)
			}
		}
	case *ast.SelectorExpr:
		for k, v := range infos {
			_, foundIt := v.Defs[t.Sel]
			if foundIt {
				log.Println(`FOUND IT SEL`, k, t)
			}
		}
	}
	return nil

	id, ok := a.Type.(*ast.Ident)
	if !ok || id == nil || id.Obj == nil {
		sel, ok := a.Type.(*ast.SelectorExpr)
		if ok {
			log.Printf("sel")
			log.Printf("%#v %#v", sel.X, sel.Sel)
		} else {

			log.Printf("not through %#v", a.Type)
		}
		return nil
	}

	ts, ok := id.Obj.Decl.(*ast.TypeSpec)
	if !ok || ts == nil {
		return nil
	}

	return membersFromTypeSpec(ts)
}

func membersFromTypeSpec(ts *ast.TypeSpec) []string {
	st, ok := ts.Type.(*ast.StructType)
	if !ok || st == nil || st.Fields == nil {
		return nil
	}

	names := []string{}
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			id, ok := field.Type.(*ast.Ident)
			if ok && id.Obj != nil {
				names = append(names, id.Obj.Name)
				continue
			}

			se, ok := field.Type.(*ast.SelectorExpr)
			if ok {
				names = append(names, se.Sel.Name)
				continue
			}
		}

		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return names
}

func removeQuotes(s string) string {
	return strings.TrimSuffix(strings.TrimPrefix(s, `"`), `"`)
}

func compile(importPath string, fset *token.FileSet) (*types.Info, []*ast.File, error) {
	p := filepath.Join(os.Getenv("GOPATH"), "src", strings.TrimSuffix(strings.TrimPrefix(importPath, `"`), `"`))

	files, err := parseAllGoFilesInDir(p, fset, false)
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

	pkgs, err := tc.Check(importPath, fset, files, info)
	if err != nil {
		return nil, nil, err
	}
	_ = pkgs
	return info, files, nil
}

func buildOutImports(files []*ast.File, fileSet *token.FileSet, packages map[string]*types.Info) {
	for _, f := range files {
		for _, i := range f.Imports {
			_, ok := packages[i.Path.Value]
			if ok {
				continue
			}

			info, nextRoundOfFiles, err := compile(i.Path.Value, fileSet)
			if err == nil {
				packages[i.Path.Value] = info
				buildOutImports(nextRoundOfFiles, fileSet, packages)
			}
		}
	}
}

func parseAllGoFilesInDir(dir string, fset *token.FileSet, recurse bool) ([]*ast.File, error) {
	files := []*ast.File{}
	_ = filepath.Walk(dir, func(filename string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if info.IsDir() {
			if !recurse && dir != filename {
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

		files = append(files, f)
		return nil
	})

	return files, nil
}
