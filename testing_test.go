package lumberjack

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func assert(condition bool, t testing.TB, msg string, v ...interface{}) {
	assertUp(condition, t, 1, msg, v...)
}

func assertUp(condition bool, t testing.TB, caller int, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(caller + 1)
		v = append([]interface{}{filepath.Base(file), line}, v...)
		fmt.Printf("%s:%d: "+msg+"\n", v...)
		t.FailNow()
	}
}

func equals(exp, act interface{}, t testing.TB) {
	equalsUp(exp, act, t, 1)
}

func equalsUp(exp, act interface{}, t testing.TB, caller int) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(caller + 1)
		fmt.Printf("%s:%d: exp: %v (%T), got: %v (%T)\n",
			filepath.Base(file), line, exp, exp, act, act)
		t.FailNow()
	}
}

func isNil(obtained interface{}, t testing.TB) {
	isNilUp(obtained, t, 1)
}

func isNilUp(obtained interface{}, t testing.TB, caller int) {
	if !_isNil(obtained) {
		_, file, line, _ := runtime.Caller(caller + 1)
		fmt.Printf("%s:%d: expected nil, got: %v\n", filepath.Base(file), line, obtained)
		t.FailNow()
	}
}

func notNil(obtained interface{}, t testing.TB) {
	notNilUp(obtained, t, 1)
}

func notNilUp(obtained interface{}, t testing.TB, caller int) {
	if _isNil(obtained) {
		_, file, line, _ := runtime.Caller(caller + 1)
		fmt.Printf("%s:%d: expected non-nil, got: %v\n", filepath.Base(file), line, obtained)
		t.FailNow()
	}
}

func _isNil(obtained interface{}) bool {
	if obtained == nil {
		return true
	}

	switch v := reflect.ValueOf(obtained); v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}

	return false
}
