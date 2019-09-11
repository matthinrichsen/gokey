package tests

import (
	"io/ioutil"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCorrection(t *testing.T) {
	expectation := `package tests

import (
	"errors"
	"log"
)

func lookAtMyTerribleFunction() {
	s := MyStruct{a: 1, b: 2, c: 3, m: map[string]string{}, arr: []int{}, MyStruct2: MyStruct2{AnotherStruct: MyStruct3{errors.New("implements Error()")}}, LastStruct: anotherPackage.LastStruct{}}
	log.Println(s)
}
`

	out, err := exec.Command(`gokey`).CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Empty(t, out)

	bytes, err := ioutil.ReadFile(`bad_function.go`)
	require.NoError(t, err)
	assert.Equal(t, expectation, string(bytes))
}
