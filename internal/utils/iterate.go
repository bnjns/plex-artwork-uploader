package utils

import (
	"maps"
	"slices"
	"sort"
)

func IterateOrdered[T any](m map[int]T, fn func(int, T)) {
	keys := slices.Collect(maps.Keys(m))
	sort.Ints(keys)
	for _, key := range keys {
		fn(key, m[key])
	}
}
