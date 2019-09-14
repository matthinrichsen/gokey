package file

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/matthinrichsen/gokey/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertAST(t *testing.T, expected, inputFile string, definitions []string) {
	fset := token.NewFileSet()
	sm := util.NewStructManager()

	for _, dir := range definitions {
		pathDir := filepath.Join(os.Getenv(`GOPATH`), `src`, dir)
		info, _, err := util.CompileFilesInDirectory(pathDir, fset)
		require.NoError(t, err)

		sm.AddPackage(dir, info)
	}

	a, err := parser.ParseFile(fset, `testfile.go`, []byte(inputFile), parser.ParseComments)
	require.NoError(t, err)

	info, err := util.CompileFiles(`somepkg`, fset, a)
	require.NoError(t, err)

	sm.AddPackage(`github.com/matthinrichsen/anotherPackage`, info)
	assert.True(t, Repair(a, `github.com/matthinrichsen/anotherPackage`, sm))

	b := &bytes.Buffer{}
	printer.Fprint(b, fset, a)
	assert.Equal(t, expected, b.String(), "%s%s%s", b.String(), "\n---------------------- VS --------------------\n\n", expected)
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

	assertAST(t, expectation, input, nil)
}

func TestFileRepair_ComplexReferenceStruct(t *testing.T) {
	input := `package anotherPackage

import "github.com/matthinrichsen/gokey/tests"

func NewStructOne() tests.AllExportedFields {
	return tests.AllExportedFields{"A", tests.AnotherExpectedFieldStruct{1,2,3}}
}
`

	expectation := `package anotherPackage

import "github.com/matthinrichsen/gokey/tests"

func NewStructOne() tests.AllExportedFields {
	return tests.AllExportedFields{A: "A", Two: tests.AnotherExpectedFieldStruct{One: 1, Two: 2, Three: 3}}
}
`

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

	expectation := `package anotherPackage

import "github.com/matthinrichsen/gokey/tests"

func NewStructOne() tests.AllExportedFields {
	return tests.AllExportedFields{
		A: "A",
		Two: tests.AnotherExpectedFieldStruct{
			One: 1,
			Two: 2,
			Three: 3,
		},
	}
}
`

	assertAST(t, expectation, input, []string{`github.com/matthinrichsen/gokey/tests`})
}
