package tests

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
}

type MyStruct2 struct {
	AnotherStruct MyStruct3
}

type MyStruct3 struct {
	someInterface
}
