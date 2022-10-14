package util

import (
	"hash/fnv"
	"github.com/segmentio/ksuid"
)

func GenUuid() string {
	return ksuid.New().String()
}

func GenHash(str string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(str))
	return h.Sum32()
}

func SliceInsertHead[T any](e T, s []T) []T {
	res := append([]T{e}, s...)
	return res
}