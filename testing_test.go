package lumberjack

import (
	"reflect"
	"testing"
)

// equals tests that the two values are equal according to reflect.DeepEqual.
func equals(exp, act interface{}, t testing.TB) {
	t.Helper()
	if !reflect.DeepEqual(exp, act) {
		t.Fatalf("exp: %v (%T), got: %v (%T)\n", exp, exp, act, act)
	}
}

// isNil reports a failure if the given value is not nil.  Note that values
// which cannot be nil will always fail this check.
func isNil(obtained interface{}, t testing.TB) {
	t.Helper()
	if !_isNil(obtained, t) {
		t.Fatalf("expected nil, got: %v (%#v)\n", obtained, obtained)
	}
}

// notNil reports a failure if the given value is nil.
func notNil(obtained interface{}, t testing.TB) {
	t.Helper()
	if _isNil(obtained, t) {
		t.Fatalf("expected non-nil, got: %#v\n", obtained)
	}
}

// _isNil is a helper function for isNil and notNil, and should not be used
// directly.
func _isNil(obtained interface{}, t testing.TB) bool {
	t.Helper()
	if obtained == nil {
		return true
	}

	switch v := reflect.ValueOf(obtained); v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}

	return false
}
