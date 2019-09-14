package file

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"testing"

	"github.com/matthinrichsen/gokey/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAstBeforeAndAfter(t *testing.T, expected, inputFile string) {
	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, `testfile.go`, []byte(inputFile), parser.ParseComments)
	require.NoError(t, err)

	sm := util.NewStructManager()
	info, err := util.CompileFiles(`tests`, fset, ast)
	require.NoError(t, err)

	sm.AddPackage(`github.com/matthinrichsen/anotherPackage`, info)
	assert.True(t, Repair(ast, `github.com/matthinrichsen/anotherPackage`, sm))

	b := &bytes.Buffer{}
	printer.Fprint(b, fset, ast)
	assert.Equal(t, expected, b.String())
}

func TestFileRepair_SimpleStruct(t *testing.T) {
	input := `package anotherPackage

type LastStruct struct {
	Int int
}

var abc = LastStruct{1}
`

	expectation := `package anotherPackage

type LastStruct struct {
	Int int
}

var abc = LastStruct{Int: 1}
`

	testAstBeforeAndAfter(t, expectation, input)
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

	expectation := `package anotherPackage

type StructOne struct {
	Name	string
	A	NestedStruct
}
type NestedStruct struct {
	One	int
	Two	int
	Three	int
}

func NewStructOne() StructOne {
	return StructOne{Name: "ThisIsMyName", A: NestedStruct{One: 1, Two: 2, Three: 3}}
}
`

	testAstBeforeAndAfter(t, expectation, input)
}
