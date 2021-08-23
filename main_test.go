package woodpecker

import (
	// External

	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFile(t *testing.T) {
	dir := makeTempDir("TestNewFile", t)
	defer os.RemoveAll(dir)

	l := New(Config{Filename: logFile(dir)})
	defer l.Close()

	b := []byte("boo!")
	n, err := l.Write(b)
	assert.NoError(t, err)
	assert.Equal(t, len(b), n)

	fileCount(t, dir, 1)
}

func TestOpenExisting(t *testing.T) {
	dir := makeTempDir("TestOpenExisting", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)
	data := []byte("foo!")
	err := ioutil.WriteFile(filename, data, 0644)
	assert.NoError(t, err)
	fileCount(t, dir, 1)

	l := New(Config{Filename: logFile(dir)})
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	assert.NoError(t, err)
	assert.Equal(t, len(b), n)

	// make sure the file got appended
	existsWithContent(t, filename, append(data, b...))

	// make sure no other files were created
	fileCount(t, dir, 1)
}

func TestWriteTooLong(t *testing.T) {
	megabyteSize = 1

	dir := makeTempDir("TestWriteTooLong", t)
	defer os.RemoveAll(dir)

	cfg := Config{
		Filename:    logFile(dir),
		MaxSizeInMb: 1,
	}
	l := New(cfg)
	defer l.Close()
	b := []byte("booooooooooo!")
	n, err := l.Write(b)
	assert.Error(t, err)
	assert.Equal(t, n, 0)

	_, err = os.Stat(logFile(dir))
	assert.True(t, os.IsNotExist(err))
}

func TestMakeLogDir(t *testing.T) {
	dir := time.Now().Format("TestMakeLogDir" + BackupTimeFormat)
	dir = filepath.Join(os.TempDir(), dir)
	defer os.RemoveAll(dir)

	l := New(Config{Filename: logFile(dir)})
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	assert.NoError(t, err)
	assert.Equal(t, len(b), n)
	existsWithContent(t, logFile(dir), b)
	fileCount(t, dir, 1)
}

func TestAutoRotate(t *testing.T) {
	megabyteSize = 1

	dir := makeTempDir("TestAutoRotate", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)
	cfg := Config{
		Filename:    filename,
		MaxSizeInMb: 10,
	}
	l := New(cfg)
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	assert.NoError(t, err)
	assert.Equal(t, len(b), n)

	existsWithContent(t, filename, b)
	fileCount(t, dir, 1)

	// New timestamp
	currentTime = currentTime().AddDate(0, 0, 1).UTC

	b2 := []byte("foooooo!")
	n, err = l.Write(b2)
	assert.NoError(t, err)
	assert.Equal(t, len(b2), n)

	// the old logfile should be moved aside and the main logfile should have only the last write in it.
	existsWithContent(t, filename, b2)

	fileCount(t, dir, 2)
}

func TestEverydayRotate(t *testing.T) {
	dir := makeTempDir("TestEverydayRotate", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)
	cfg := Config{
		Filename:       filename,
		RotateEveryday: true,
		MaxSizeInMb:    10,
	}
	l := New(cfg)
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	assert.NoError(t, err)
	assert.Equal(t, len(b), n)

	existsWithContent(t, filename, b)
	fileCount(t, dir, 1)

	// New timestamp
	currentTime = currentTime().AddDate(0, 0, 1).UTC

	b2 := []byte("foooooo!")
	n, err = l.Write(b2)
	assert.NoError(t, err)
	assert.Equal(t, len(b2), n)

	// the old logfile should be moved aside and the main logfile should have only the last write in it.
	existsWithContent(t, filename, b2)

	fileCount(t, dir, 2)
}

func TestFirstWriteRotate(t *testing.T) {
	megabyteSize = 1

	dir := makeTempDir("TestFirstWriteRotate", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)
	cfg := Config{
		Filename:    filename,
		MaxSizeInMb: 10,
	}
	l := New(cfg)
	defer l.Close()

	start := []byte("boooooo!")
	err := ioutil.WriteFile(filename, start, 0600)
	assert.NoError(t, err)

	// New timestamp
	currentTime = currentTime().AddDate(0, 0, 2).UTC

	// this would make us rotate
	b := []byte("fooo!")
	n, err := l.Write(b)
	assert.NoError(t, err)
	assert.Equal(t, len(b), n)

	existsWithContent(t, filename, b)

	fileCount(t, dir, 2)
}

// makeTempDir creates a file with a semi-unique name in the OS temp directory.
// It should be based on the name of the test, to keep parallel tests from
// colliding, and must be cleaned up after the test is finished.
func makeTempDir(name string, t testing.TB) string {
	dir := time.Now().Format(name + BackupTimeFormat)
	dir = filepath.Join(os.TempDir(), dir)
	err := os.Mkdir(dir, 0700)
	assert.NoError(t, err)
	return dir
}

// existsWithContent checks that the given file exists and has the correct content.
func existsWithContent(t *testing.T, path string, content []byte) {
	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), info.Size())

	b, err := ioutil.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, content, b)
}

// logFile returns the log file name in the given directory for the current fake time.
func logFile(dir string) string {
	return filepath.Join(dir, "test.log")
}

// fileCount checks that the number of files in the directory is exp.
func fileCount(t *testing.T, dir string, exp int) {
	files, err := ioutil.ReadDir(dir)
	assert.NoError(t, err)
	// Make sure no other files were created.
	assert.Equal(t, len(files), exp)
}
