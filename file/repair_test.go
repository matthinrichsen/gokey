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

func TestFileRepair_SimpleStruct(t *testing.T) {
	file := `package anotherPackage

type LastStruct struct {
	Int int
}

var abc = LastStruct{1}
`
	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, `testfile.go`, []byte(file), parser.ParseComments)
	require.NoError(t, err)

	sm := util.NewStructManager()
	info, err := util.CompileFiles(`tests`, fset, ast)
	require.NoError(t, err)

	sm.AddPackage(`github.com/matthinrichsen/anotherPackage`, info)
	assert.True(t, Repair(ast, `github.com/matthinrichsen/anotherPackage`, sm))

	b := &bytes.Buffer{}
	printer.Fprint(b, fset, ast)

	expectedString := `package anotherPackage

type LastStruct struct {
	Int int
}

var abc = LastStruct{Int: 1}
`
	assert.Equal(t, expectedString, b.String())
}
