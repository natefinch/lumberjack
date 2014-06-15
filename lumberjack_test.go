package lumberjack

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// !!!NOTE!!!
//
// Running these tests in parallel will almost certainly cause sporadic (or even
// regular) failures, because they're all messing with the same global variable
// that controls the logic's mocked time.Now.  So... don't do that.

// make sure we set the format to something safe for windows, too.
const format = "2006-01-02T15-04-05.000"

// Since all the tests uses the time to determine filenames etc, we need to
// control the wall clock as much as possible, which means having a wall clock
// that doesn't change unless we want it to.
var fakeCurrentTime = time.Now()

func fakeTime() time.Time {
	return fakeCurrentTime
}

func TestNewFile(t *testing.T) {
	currentTime = fakeTime

	dir := makeTempDir("TestNewFile", t)
	defer os.RemoveAll(dir)
	l := &Logger{
		Dir:        dir,
		NameFormat: format,
		MaxSize:    Megabyte,
	}
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	isNil(err, t)
	equals(len(b), n, t)
	existsWithLen(logFile(dir), n, t)
	fileCount(dir, 1, t)
}

func TestOpenExisting(t *testing.T) {
	currentTime = fakeTime
	dir := makeTempDir("TestOpenExisting", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)
	data := []byte("foo!")
	err := ioutil.WriteFile(filename, data, 0644)
	isNil(err, t)
	existsWithLen(filename, len(data), t)

	l := &Logger{
		Dir:        dir,
		NameFormat: format,
		MaxSize:    Megabyte,
	}
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	// make sure the file got appended
	existsWithLen(filename, len(data)+n, t)

	// make sure no other files were created
	fileCount(dir, 1, t)
}

func TestWriteTooLong(t *testing.T) {
	currentTime = fakeTime
	dir := makeTempDir("TestWriteTooLong", t)
	defer os.RemoveAll(dir)
	l := &Logger{
		Dir:        dir,
		NameFormat: format,
		MaxSize:    5,
	}
	defer l.Close()
	b := []byte("booooooooooooooo!")
	n, err := l.Write(b)
	assert(IsWriteTooLong(err), t,
		"Should have gotten write too long error, instead got %s (%T)", err, err)
	equals(0, n, t)
	_, err = os.Stat(logFile(dir))
	assert(os.IsNotExist(err), t, "File exists, but should not have been created")
}

func TestRotate(t *testing.T) {
	currentTime = fakeTime
	dir := makeTempDir("TestRotate", t)
	defer os.RemoveAll(dir)

	l := &Logger{
		Dir:        dir,
		NameFormat: format,
		MaxSize:    10,
	}
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	filename := logFile(dir)
	existsWithLen(filename, n, t)
	fileCount(dir, 1, t)

	// set the current time one day later
	defer newFakeTime(Day)()

	b2 := []byte("foooooo!")
	n, err = l.Write(b2)
	isNil(err, t)
	equals(len(b2), n, t)

	// this will use the new fake time
	newFilename := logFile(dir)
	existsWithLen(newFilename, n, t)

	// make sure the old file still exists with the same size.
	existsWithLen(filename, len(b), t)

	fileCount(dir, 2, t)
}

func TestMaxBackups(t *testing.T) {
	currentTime = fakeTime
	dir := makeTempDir("TestMaxBackups", t)
	defer os.RemoveAll(dir)

	l := &Logger{
		Dir:        dir,
		NameFormat: format,
		MaxSize:    10,
		MaxBackups: 1,
	}
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	firstFilename := logFile(dir)
	existsWithLen(firstFilename, n, t)
	fileCount(dir, 1, t)

	// set the current time one day later
	defer newFakeTime(Day)()

	// this will put us over the max
	b2 := []byte("foooooo!")
	n, err = l.Write(b2)
	isNil(err, t)
	equals(len(b2), n, t)

	// this will use the new fake time
	secondFilename := logFile(dir)
	existsWithLen(secondFilename, n, t)

	// make sure the old file still exists with the same size.
	existsWithLen(firstFilename, len(b), t)

	fileCount(dir, 2, t)

	// set the current time one day later
	defer newFakeTime(Day)()

	// this will make us rotate again
	n, err = l.Write(b2)
	isNil(err, t)
	equals(len(b2), n, t)

	// this will use the new fake time
	thirdFilename := logFile(dir)
	existsWithLen(thirdFilename, n, t)

	// we need to wait a little bit since the files get deleted on a different
	// goroutine.
	<-time.After(time.Millisecond * 10)

	// should only have two files in the dir still
	fileCount(dir, 2, t)

	// second file name should still exist
	existsWithLen(secondFilename, n, t)

	// should have deleted the first filename
	notExist(firstFilename, t)
}

func TestMaxAge(t *testing.T) {
	currentTime = fakeTime
	// This test uses ModTime on files, and so we need to make sure we're using
	// the most current time possible.
	fakeCurrentTime = time.Now()
	dir := makeTempDir("TestMaxAge", t)
	defer os.RemoveAll(dir)

	l := &Logger{
		Dir:        dir,
		NameFormat: format,
		MaxSize:    10,
		MaxAge:     10 * time.Millisecond,
	}
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	filename := logFile(dir)
	existsWithLen(filename, n, t)
	fileCount(dir, 1, t)

	// We need to wait for wall clock time since MaxAge uses file ModTime, which
	// can't be mocked.
	<-time.After(50 * time.Millisecond)
	fakeCurrentTime = time.Now()

	b2 := []byte("foooooo!")
	n, err = l.Write(b2)
	isNil(err, t)
	equals(len(b2), n, t)

	// we need to wait a little bit since the files get deleted on a different
	// goroutine.
	<-time.After(10 * time.Millisecond)

	// We should have just one log file
	fileCount(dir, 1, t)

	// this will use the new fake time
	newFilename := logFile(dir)
	existsWithLen(newFilename, n, t)

	// we should have deleted the old file due to being too old
	notExist(filename, t)
}

func TestLocalTime(t *testing.T) {
	currentTime = fakeTime

	dir := makeTempDir("TestMaxAge", t)
	defer os.RemoveAll(dir)

	l := &Logger{
		Dir:        dir,
		NameFormat: format,
		LocalTime:  true,
	}
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	filename := logFileLocal(dir)
	existsWithLen(filename, n, t)
}

// makeTempDir creates a file with a semi-unique name in the OS temp directory.
// It should be based on the name of the test, to keep parallel tests from
// colliding, and must be cleaned up after the test is finished.
func makeTempDir(name string, t testing.TB) string {
	dir := time.Now().Format(name + format)
	dir = filepath.Join(os.TempDir(), dir)
	isNilUp(os.Mkdir(dir, 0777), t, 1)
	return dir
}

// existsWithLen checks that the given file exists and has the correct length.
func existsWithLen(path string, length int, t testing.TB) {
	info, err := os.Stat(path)
	isNilUp(err, t, 1)
	equalsUp(int64(length), info.Size(), t, 1)
}

// logFile returns the log file name in the given directory for the current fake
// time.
func logFile(dir string) string {
	return filepath.Join(dir, fakeTime().UTC().Format(format))
}

// logFileLocal returns the log file name in the given directory for the current
// fake time using the local timezone.
func logFileLocal(dir string) string {
	return filepath.Join(dir, fakeTime().Format(format))
}

// fileCount checks that the number of files in the directory is exp.
func fileCount(dir string, exp int, t testing.TB) {
	files, err := ioutil.ReadDir(dir)
	isNilUp(err, t, 1)
	// Make sure no other files were created.
	equalsUp(exp, len(files), t, 1)
}

// newFakeTime sets the fake "current time" to one day later.
func newFakeTime(later time.Duration) func() {
	old := fakeCurrentTime
	fakeCurrentTime = fakeCurrentTime.Add(later)
	return func() {
		fakeCurrentTime = old
	}
}

func notExist(path string, t testing.TB) {
	_, err := os.Stat(path)
	assertUp(os.IsNotExist(err), t, 1, "expected to get os.IsNotExist, but instead got %v", err)
}
