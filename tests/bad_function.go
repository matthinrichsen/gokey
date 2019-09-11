package tests

import (
	"errors"
	"log"

	"github.com/matthinrichsen/gokey/tests/anotherPackage"
)

func lookAtMyTerribleFunction() {
	a := anotherPackage.LastStruct{2}
	s := MyStruct{1, 2, 3, map[string]string{}, []int{}, MyStruct2{MyStruct3{errors.New(`implements Error()`)}}, anotherPackage.LastStruct{1}, a}
	log.Println(s, a)
}
