package utils

import "fmt"

// ExampleStringWithCharset показывает, как создать строку из заданного набора символов указанной длины.
func ExampleStringWithCharset() {
	s := StringWithCharset(5, "01")
	fmt.Println(len(s))
	// Output: 5
}

// ExampleRandomString показывает, как получить случайную строку заданной длины.
func ExampleRandomString() {
	s := RandomString(8)
	fmt.Println(len(s))
	// Output: 8
}
