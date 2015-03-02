package tests

import (
	"bytes"
	"runtime"
	"testing"
)

func CheckString(t *testing.T, expected, actual string) {
	if expected != actual {
		failure(t, expected, actual)
	}
}

func CheckBool(t *testing.T, expected, actual bool) {
	if expected != actual {
		failure(t, expected, actual)
	}
}

func CheckBytes(t *testing.T, expected, actual []byte) {
	if !bytes.Equal(expected, actual) {
		failure(t, expected, actual)
	}
}

func CheckInt64(t *testing.T, expected, actual int64) {
	if expected != actual {
		failure(t, expected, actual)
	}
}

func failure(t *testing.T, expected, actual interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		t.Fatalf("Failure at %s:%d! Expected '%v', got '%v'\n", file, line, expected, actual)
	} else {
		t.Fatalf("Failure! Expected '%v', got '%v'\n", expected, actual)
	}
}
