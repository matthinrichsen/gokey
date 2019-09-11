package tests

import (
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
