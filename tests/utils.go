package tests

import (
	"fmt"
	"slices"
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

// Reverse the elements of the slice and return a copy
func RevCopy[S ~[]T, T any](s S) S {
	b := make([]T, len(s))
	copy(b, s)
	slices.Reverse(b)
	return b
}

func PrintData(d any) string {
	return fmt.Sprintf(
		"\n%#v\n-----------------------------------------------", d)
}
