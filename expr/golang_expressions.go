package expr

import "github.com/samber/lo"

func IfElse[T any](cond bool, trueVal, falseVal T) T {
	if cond {
		return trueVal
	}
	return falseVal
}

func Contains[T comparable](elems []T, v T) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func Unique[T comparable](elems []T) []T {
	return lo.Uniq(elems)
}
