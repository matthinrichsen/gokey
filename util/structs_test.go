package util

import (
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructInFile(t *testing.T) {

	fset := token.NewFileSet()
	info, _, err := CompileFilesInDirectory(`../tests`, fset)
	require.NoError(t, err)

	sm := NewStructManager()
	sm.AddPackage(`github.com/matthinrichsen/gokey/tests`, info)

	names, ok := sm.Get(`github.com/matthinrichsen/gokey/tests`, `MyStruct`)
	require.True(t, ok)
	assert.Equal(t, []string{`a`, `b`, `c`, `m`, `arr`, `MyStruct2`, `LastStruct`, `LS`}, names)

	names, ok = sm.Get(`github.com/matthinrichsen/gokey/tests`, `MyStruct2`)
	require.True(t, ok)
	assert.Equal(t, []string{`AnotherStruct`}, names)

	names, ok = sm.Get(`github.com/matthinrichsen/gokey/tests`, `MyStruct3`)
	require.True(t, ok)
	assert.Equal(t, []string{`someInterface`}, names)

	names, ok = sm.Get(`github.com/matthinrichsen/gokey/tests`, `MyStruct4`)
	require.True(t, ok)
	assert.Equal(t, []string{`a`, `error`}, names)

	names, ok = sm.Get(`github.com/matthinrichsen/gokey/tests`, `dne`)
	require.False(t, ok)
	assert.Empty(t, names)
}
