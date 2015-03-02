package base

import (
	"errors"
	"testing"

	"github.com/urandom/readeef/tests"
)

func TestErr(t *testing.T) {
	err := Error{}

	tests.CheckBool(t, false, err.HasErr())
	tests.CheckBool(t, true, err.Err() == nil)

	err1 := errors.New("test1")
	tests.CheckBool(t, true, err.Err(err1) == nil)

	tests.CheckBool(t, true, err.HasErr())
	tests.CheckBool(t, true, err.Err() == err1)

	tests.CheckBool(t, false, err.HasErr())
	tests.CheckBool(t, true, err.Err() == nil)
}
