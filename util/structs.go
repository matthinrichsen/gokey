package util

import (
	"go/ast"
	"go/types"
)

type StructManager struct {
	packages    map[string]*types.Info
	defsToNames map[string][]string
}

func NewStructManager() StructManager {
	return StructManager{
		packages:    map[string]*types.Info{},
		defsToNames: map[string][]string{},
	}
}

func (s StructManager) AddPackage(importDir string, pkg *types.Info) {
	s.packages[RemoveQuotes(importDir)] = pkg
}

func (s StructManager) HasPackage(importDir string) bool {
	_, ok := s.packages[RemoveQuotes(importDir)]
	return ok
}

func (s StructManager) Get(pkg, structName string) ([]string, bool) {
	pkg = RemoveQuotes(pkg)
	structName = RemoveQuotes(structName)
	ref := pkg + `.` + structName

	_, ok := s.defsToNames[ref]
	if !ok {
		info, ok := s.packages[pkg]
		if !ok {
			return nil, false
		}

		for i := range info.Defs {
			if i.Obj != nil {
				ts, ok := i.Obj.Decl.(*ast.TypeSpec)
				if ok {
					s.defsToNames[pkg+"."+i.Name] = membersFromTypeSpec(ts)
				}
			}
		}
	}

	names, ok := s.defsToNames[ref]
	return names, ok
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
			if ok {
				if id.Obj != nil {
					names = append(names, id.Obj.Name)
				} else {
					names = append(names, id.Name)
				}
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
