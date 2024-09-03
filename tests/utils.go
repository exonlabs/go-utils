package tests

import (
	"fmt"
)

func GreenText(msg string) string {
	return fmt.Sprintf("\033[0;32m%v\033[0m", msg)
}

func RedText(msg string) string {
	return fmt.Sprintf("\033[0;31m%v\033[0m", msg)
}

func ValidMsg() string {
	return GreenText("VALID")
}

func FailMsg() string {
	return RedText("FAILED")
}

// Reverse reverses the elements of the slice in place.
func Reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// Reverse the elements of the slice and return a copy
func RevCopy[S ~[]T, T any](s S) S {
	b := make([]T, len(s))
	copy(b, s)
	Reverse(b)
	return b
}

func PrintData(d any) string {
	return fmt.Sprintf(
		"\n%#v\n-----------------------------------------------", d)
}
