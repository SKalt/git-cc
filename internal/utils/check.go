package utils

import "strings"

func Check(err error) {
	if err != nil {
		panic(err)
	}
}
func Must[T any](t T, err error) T {
	Check(err)
	return t
}

func Render(r func(*strings.Builder)) string {
	var b strings.Builder
	r(&b)
	return b.String()
}

func Coalesce[T any](a *T, b T) T {
	if a == nil {
		return b
	}
	return *a
}
