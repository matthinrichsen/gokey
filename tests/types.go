package tests

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

var withinTheSameFile = MyStruct2{MyStruct3{errors.New(`wazzup`)}}
var withinTheSameFile2 = MyStruct3{errors.New(`wazzup`)}
var withinTheSameFile3 = MyStruct4{1, errors.New(`wazzup`)}
