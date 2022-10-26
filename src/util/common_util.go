package util

import (
	"os"
	"hash/fnv"
	"path/filepath"
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

func GetExecName() string {
	path, _ := os.Executable()
	_, exec := filepath.Split(path)
	return exec
}

func GetExecPath() string {
	path, _ := os.Executable()
	return path
}

func SliceInsertHead[T any](e T, s []T) []T {
	res := append([]T{e}, s...)
	return res
}