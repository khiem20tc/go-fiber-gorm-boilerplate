package utils

import (
	"github.com/samber/lo"
)

func Or[T interface{ comparable }](args ...T) T {
	var result T
	for _, args := range args {
		result = args
		if !lo.IsEmpty[T](args) {
			return args
		}
	}

	return result
}

func And[T interface{ comparable }](args ...T) T {
	var result T
	for _, args := range args {
		result = args
		if lo.IsEmpty[T](args) {
			return args
		}
	}

	return result
}
