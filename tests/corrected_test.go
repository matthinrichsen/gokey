package tests

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCorrection(t *testing.T) {
	out, err := exec.Command(`gokey`).CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Empty(t, out)

	err = filepath.Walk(`.`, func(fp string, info os.FileInfo, err error) error {
		require.NoError(t, err)

		ext := filepath.Ext(fp)
		if info.IsDir() || ext != `.go` || strings.HasSuffix(fp, `_test.go`) {
			return nil
		}

		t.Log(`checking ` + fp)

		expectedFile := strings.TrimSuffix(fp, ext) + `.expected`

		expectedBytes, err := ioutil.ReadFile(expectedFile)
		require.NoError(t, err)

		actualBytes, err := ioutil.ReadFile(fp)
		require.NoError(t, err)
		assert.Equal(t, string(expectedBytes), string(actualBytes))

		return nil
	})
	require.NoError(t, err)
}
