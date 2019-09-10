package tests

import (
	"errors"
	"log"
)

func lookAtMyTerribleFunction() {
	s := MyStruct{1, 2, 3, map[string]string{}, []int{}, MyStruct2{MyStruct3{errors.New(`implements Error()`)}}}
	log.Println(s)
}
