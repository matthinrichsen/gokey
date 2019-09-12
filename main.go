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
	"strings"
	"sync"
)

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

	fileSet := token.NewFileSet()
	packages := map[string]*types.Info{}
	structFieldNames := map[string][]string{}

	_ = filepath.Walk(path, func(directory string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		astinfo, allFiles, err := compile(directory, fileSet)
		if err != nil {
			return nil
		}

		importDir, err := filepath.Rel(filepath.Join(os.Getenv(`GOPATH`), `src`), directory)
		if err != nil {
			importDir = directory
		}

		packages[importDir] = astinfo
		buildOutImports(allFiles, fileSet, packages)

		for filename, f := range allFiles {

			importsToPaths := map[string]string{}
			once := sync.Once{}
			nodesToRepaired := false
			buildOutImports := func() {
				nodesToRepaired = true
				for _, i := range f.Imports {
					if i.Name != nil {
						importsToPaths[removeQuotes(i.Name.String())] = removeQuotes(i.Path.Value)
					}
					_, value := filepath.Split(removeQuotes(i.Path.Value))
					importsToPaths[value] = removeQuotes(i.Path.Value)
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
								importReference = importsToPaths[removeQuotes(pkg.String())]
								structReference = importReference + `.` + structName
							}
						case *ast.Ident: // this struct declaration is local to the package
							importReference = importReference
							structReference = importReference + `.` + t.Name
						}

						names, ok := structFieldNames[structReference]
						if !ok {
							for i := range packages[importReference].Defs {
								if i.Obj != nil {
									ts, ok := i.Obj.Decl.(*ast.TypeSpec)
									if ok {
										structFieldNames[removeQuotes(importReference)+"."+i.Name] = membersFromTypeSpec(ts)
									}
								}
							}
						}

						if len(names) < i {
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

			if nodesToRepaired {
				fmt.Println(filename)
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

func buildOutImports(files map[string]*ast.File, fileSet *token.FileSet, packages map[string]*types.Info) {
	for _, f := range files {
		for _, i := range f.Imports {
			_, ok := packages[i.Path.Value]
			if ok {
				continue
			}

			info, nextRoundOfFiles, err := compile(filepath.Join(os.Getenv("GOPATH"), "src", removeQuotes(i.Path.Value)), fileSet)
			if err == nil {
				packages[i.Path.Value] = info
				buildOutImports(nextRoundOfFiles, fileSet, packages)
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
