package file

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertPrint(t *testing.T, expected, input string) {
	fset := token.NewFileSet()
	a, err := parser.ParseFile(fset, `testfile.go`, []byte(input), parser.ParseComments)
	require.NoError(t, err)

	tokenFile := fset.File(a.Pos())
	linePositions := getLineTokenPositions(tokenFile)
	lines := getLineOffsets(linePositions, tokenFile)

	actual, err := PrintRepair(a, RepairInfo{Lines: lines})
	require.NoError(t, err)
	assert.Equal(t, expected, b.String(), "%s\n---------------------- VS --------------------\n\n%s", string(actual), expected)
}

func TestPrint_ShouldFormat(t *testing.T) {
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

	expected := `package anotherPackage

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
	assertPrint(t, expected, input)
}
