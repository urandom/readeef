package tests

import (
	"bytes"
	"runtime"
	"testing"
)

func CheckString(t *testing.T, expected, actual string, msg ...interface{}) {
	if expected != actual {
		failure(t, expected, actual, msg...)
	}
}

func CheckBool(t *testing.T, expected, actual bool, msg ...interface{}) {
	if expected != actual {
		failure(t, expected, actual, msg...)
	}
}

func CheckBytes(t *testing.T, expected, actual []byte, msg ...interface{}) {
	if !bytes.Equal(expected, actual) {
		failure(t, expected, actual, msg...)
	}
}

func CheckInt64(t *testing.T, expected, actual int64, msg ...interface{}) {
	if expected != actual {
		failure(t, expected, actual, msg...)
	}
}

func failure(t *testing.T, expected, actual interface{}, msg ...interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if len(msg) > 0 {
		t.Error(msg[0])
	}
	if ok {
		t.Fatalf("Failure at %s:%d! Expected '%v', got '%v'\n", file, line, expected, actual)
	} else {
		t.Fatalf("Failure! Expected '%v', got '%v'\n", expected, actual)
	}
}
