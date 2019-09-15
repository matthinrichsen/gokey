package file

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthinrichsen/gokey/util"
)

func assertAST(t *testing.T, expected, inputFile string, definitions []string) {
	fset := token.NewFileSet()
	sm := util.NewStructManager()

	a, err := parser.ParseFile(fset, `testfile.go`, []byte(inputFile), parser.ParseComments)
	require.NoError(t, err)

	for _, dir := range definitions {
		pathDir := filepath.Join(os.Getenv(`GOPATH`), `src`, dir)
		info, _, err := util.CompileFilesInDirectory(pathDir, fset)
		require.NoError(t, err)

		sm.AddPackage(dir, info)
	}

	info, err := util.CompileFiles(`somepkg`, fset, a)
	require.NoError(t, err)

	sm.AddPackage(`github.com/matthinrichsen/anotherPackage`, info)
	repairInfo, repaired := Repair(a, `github.com/matthinrichsen/anotherPackage`, sm, fset)
	assert.True(t, repaired)

	actual, err := PrintRepair(a, repairInfo)
	require.NoError(t, err)
	assert.Equal(t, expected, string(actual), "%s\n---------------------- VS --------------------\n\n%s", string(actual), expected)
}

func TestFileRepair_SimpleStruct(t *testing.T) {
	input := `package anotherPackage

type LastStruct struct {
	Int int
}

var abc = LastStruct{1}
`

	expectation := "package anotherPackage\n\ntype LastStruct struct {\n\tInt int\n}\n\nvar abc = LastStruct{Int: 1}\n"

	assertAST(t, expectation, input, nil)
}

func TestFileRepair_ComplexStruct(t *testing.T) {
	input := `package anotherPackage

type StructOne struct {
	Name string
	A	NestedStruct
}
type NestedStruct struct {
	One	int
	Two	int
	Three	int
}

func NewStructOne() StructOne {
	return StructOne{"ThisIsMyName", NestedStruct{1,2,3}}
}
`

	expectation := "package anotherPackage\n\ntype StructOne struct {\n\tName string\n\tA    NestedStruct\n}\ntype NestedStruct struct {\n\tOne   int\n\tTwo   int\n\tThree int\n}\n\nfunc NewStructOne() StructOne {\n\treturn StructOne{Name: \"ThisIsMyName\", A: NestedStruct{One: 1, Two: 2, Three: 3}}\n}\n"
	assertAST(t, expectation, input, nil)
}

func TestFileRepair_ComplexReferenceStruct(t *testing.T) {
	input := `package anotherPackage

import "github.com/matthinrichsen/gokey/tests"

func NewStructOne() tests.AllExportedFields {
	return tests.AllExportedFields{"A", tests.AnotherExpectedFieldStruct{1,2,3}}
}
`

	expectation := "package anotherPackage\n\nimport \"github.com/matthinrichsen/gokey/tests\"\n\nfunc NewStructOne() tests.AllExportedFields {\n\treturn tests.AllExportedFields{A: \"A\", Two: tests.AnotherExpectedFieldStruct{One: 1, Two: 2, Three: 3}}\n}\n"

	assertAST(t, expectation, input, []string{`github.com/matthinrichsen/gokey/tests`})
}

func TestFileRepair_ComplexReferenceStruct_Newlines(t *testing.T) {
	input := `package anotherPackage

import "github.com/matthinrichsen/gokey/tests"

func NewStructOne() tests.AllExportedFields {
	return tests.AllExportedFields{
		"A",
		tests.AnotherExpectedFieldStruct{
			1,
			2,
			3,
		},
	}
}
`

	expectation := "package anotherPackage\n\nimport \"github.com/matthinrichsen/gokey/tests\"\n\nfunc NewStructOne() tests.AllExportedFields {\n\treturn tests.AllExportedFields{\n\t\tA: \"A\",\n\t\tTwo: tests.AnotherExpectedFieldStruct{\n\t\t\tOne:   1,\n\t\t\tTwo:   2,\n\t\t\tThree: 3,\n\t\t},\n\t}\n}\n"

	assertAST(t, expectation, input, []string{`github.com/matthinrichsen/gokey/tests`})
}

func TestNudging(t *testing.T) {
	tests := []struct {
		input    ast.Node
		expected ast.Node
		nudge    int64
	}{{
		input: &ast.SelectorExpr{
			X: &ast.Ident{
				NamePos: 12,
				Name:    `123`,
			},
			Sel: &ast.Ident{
				NamePos: 13,
				Name:    `123`,
			},
		},
		expected: &ast.SelectorExpr{
			X: &ast.Ident{
				NamePos: 17,
				Name:    `123`,
			},
			Sel: &ast.Ident{
				NamePos: 18,
				Name:    `123`,
			},
		},
		nudge: 5,
	}, {
		input: &ast.Ident{
			NamePos: 13,
			Name:    `123`,
		},
		expected: &ast.Ident{
			NamePos: 18,
			Name:    `123`,
		},
		nudge: 5,
	}}
	for _, tc := range tests {
		nudgeTokenPositions(tc.input, tc.nudge)
		assert.Equal(t, tc.expected.Pos(), tc.input.Pos())
		assert.Equal(t, tc.expected.End(), tc.input.End())
		assert.Equal(t, tc.expected, tc.input)
	}
}
