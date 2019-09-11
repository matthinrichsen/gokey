package tests

import (
	"io/ioutil"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCorrection_BadFunction(t *testing.T) {
	expectation := "package tests

import (
	\"errors\"
	\"log\"
)

func lookAtMyTerribleFunction() {
	s := MyStruct{a: 1, b: 2, c: 3, m: map[string]string{}, arr: []int{}, MyStruct2: MyStruct2{AnotherStruct: MyStruct3{errors.New(\"implements Error()\")}}, LastStruct: anotherPackage.LastStruct{}}
	log.Println(s)
}"

	out, err := exec.Command(`gokey`).CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Empty(t, out)

	bytes, err := ioutil.ReadFile(`bad_function.go`)
	require.NoError(t, err)
	assert.Equal(t, expectation, string(bytes))
}

func TestCorrection_StructDefs(t *testing.T) {
	expectation :="package tests

import (
	\"errors\"

	\"github.com/matthinrichsen/gokey/tests/anotherPackage\"
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

var withinTheSameFile = MyStruct2{AnotherStruct: MyStruct3{someInterface: errors.New(`wazzup`)}}
var withinTheSameFile2 = MyStruct3{someInterface: errors.New(`wazzup`)}
var withinTheSameFile3 = MyStruct4{1, errors.New(`wazzup`)}"

	out, err := exec.Command(`gokey`).CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Empty(t, out)

	bytes, err := ioutil.ReadFile(`bad_function.go`)
	require.NoError(t, err)
	assert.Equal(t, expectation, string(bytes))
}
