package file

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/ast/astutil"

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

func TestFileRepair_InlineInternals(t *testing.T) {
	input := `package tests

import (
	"errors"

	"github.com/matthinrichsen/gokey/tests/anotherPackage"
)

type someInterface interface {
	Error() string
}

type MyStruct struct {
	a int
	b int
	c int

	m   map[string]string
	arr []int

	MyStruct2

	anotherPackage.LastStruct
	LS anotherPackage.LastStruct
}

type MyStruct2 struct {
	AnotherStruct MyStruct3
}

type MyStruct3 struct {
	someInterface
}

type MyStruct4 struct {
	a     int16
	error // embedding native go type
}

var withinTheSameFile = MyStruct2{MyStruct3{errors.New("wazzup")}}
var withinTheSameFile2 = MyStruct3{errors.New("wazzup")}
var withinTheSameFile3 = MyStruct4{1, errors.New("wazzup")}

type AllExportedFields struct {
	A   string
	Two AnotherExpectedFieldStruct
}

type AnotherExpectedFieldStruct struct {
	One        int
	Two, Three int
}
`

	expected := "package tests\n\nimport (\n\t\"errors\"\n\n\t\"github.com/matthinrichsen/gokey/tests/anotherPackage\"\n)\n\ntype someInterface interface {\n\tError() string\n}\n\ntype MyStruct struct {\n\ta int\n\tb int\n\tc int\n\n\tm   map[string]string\n\tarr []int\n\n\tMyStruct2\n\n\tanotherPackage.LastStruct\n\tLS anotherPackage.LastStruct\n}\n\ntype MyStruct2 struct {\n\tAnotherStruct MyStruct3\n}\n\ntype MyStruct3 struct {\n\tsomeInterface\n}\n\ntype MyStruct4 struct {\n\ta     int16\n\terror // embedding native go type\n}\n\nvar withinTheSameFile = MyStruct2{AnotherStruct: MyStruct3{someInterface: errors.New(\"wazzup\")}}\nvar withinTheSameFile2 = MyStruct3{someInterface: errors.New(\"wazzup\")}\nvar withinTheSameFile3 = MyStruct4{a: 1, error: errors.New(\"wazzup\")}\n\ntype AllExportedFields struct {\n\tA   string\n\tTwo AnotherExpectedFieldStruct\n}\ntype AnotherExpectedFieldStruct struct {\n\tOne        int\n\tTwo, Three int\n}\n"
	assertAST(t, expected, input, []string{"github.com/matthinrichsen/gokey/tests/anotherPackage"})
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

func debug(t *testing.T, input string) {

	fset := token.NewFileSet()

	a, err := parser.ParseFile(fset, `testfile.go`, []byte(input), parser.ParseComments)
	require.NoError(t, err)
	debugPositions(a)
}

func debugPositions(a *ast.File) {
	spew.Dump(a)
	astutil.Apply(a, func(c *astutil.Cursor) bool {
		n := c.Node()
		if n != nil {
			//		fmt.Println(n.Pos())
		}
		return true
	}, nil)
}
