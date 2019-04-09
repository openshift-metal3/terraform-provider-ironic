package testhelper

import (
	"strings"
	"testing"
)

func AssertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func AssertError(t *testing.T, err error, expected string) {
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain '%s', but was '%s'", expected, err)
	}
}