package tests

import (
	"io/ioutil"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCorrection_BadFunction(t *testing.T) {
	expectation, err := ioutil.ReadFile(`bad_function.expected`)
	require.NoError(t, err)

	out, err := exec.Command(`gokey`).CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Empty(t, out)

	bytes, err := ioutil.ReadFile(`bad_function.go`)
	require.NoError(t, err)
	assert.Equal(t, string(expectation), string(bytes))
}

func TestCorrection_StructDefs(t *testing.T) {
	expectation, err := ioutil.ReadFile(`types.expected`)
	require.NoError(t, err)

	out, err := exec.Command(`gokey`).CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Empty(t, out)

	bytes, err := ioutil.ReadFile(`types.go`)
	require.NoError(t, err)
	assert.Equal(t, string(expectation), string(bytes))
}
