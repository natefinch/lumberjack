package lumberjack

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ensure we always implement io.WriteCloser
var _ io.WriteCloser = (*Roller)(nil)

// !!!NOTE!!!
//
// Running these tests in parallel will almost certainly cause sporadic (or even
// regular) failures, because they're all messing with the same global variable
// that controls the logic's mocked time.Now.  So... don't do that.

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
	r, err := NewRoller(logFile(dir), 5, nil)
	isNil(err, t)
	defer r.Close()
	b := []byte("boo!")
	n, err := r.Write(b)
	isNil(err, t)
	equals(len(b), n, t)
	existsWithContent(logFile(dir), b, t)
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
	existsWithContent(filename, data, t)

	r, err := NewRoller(filename, 100, nil)
	isNil(err, t)
	defer r.Close()
	b := []byte("boo!")
	n, err := r.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	// make sure the file got appended
	existsWithContent(filename, append(data, b...), t)

	// make sure no other files were created
	fileCount(dir, 1, t)
}

func TestWriteTooLong(t *testing.T) {
	currentTime = fakeTime
	dir := makeTempDir("TestWriteTooLong", t)
	defer os.RemoveAll(dir)
	r, err := NewRoller(logFile(dir), 5, nil)
	isNil(err, t)
	defer r.Close()
	b := []byte("booooooooooooooo!")
	n, err := r.Write(b)
	notNil(err, t)
	equals(0, n, t)
	if !errors.Is(err, ErrWriteTooLong) {
		t.Fatalf("expected error to be ErrWriteTooLong, but it was %#v", err)
	}
}

func TestMakeLogDir(t *testing.T) {
	currentTime = fakeTime
	dir := time.Now().Format("TestMakeLogDir" + backupTimeFormat)
	dir = filepath.Join(os.TempDir(), dir)
	defer os.RemoveAll(dir)
	r, err := NewRoller(logFile(dir), 5, nil)
	isNil(err, t)
	defer r.Close()
	b := []byte("boo!")
	n, err := r.Write(b)
	isNil(err, t)
	equals(len(b), n, t)
	existsWithContent(logFile(dir), b, t)
	fileCount(dir, 1, t)
}

func TestEmptyFilename(t *testing.T) {
	_, err := NewRoller("", 100, nil)
	notNil(err, t)
}

func TestAutoRotate(t *testing.T) {
	currentTime = fakeTime

	dir := makeTempDir("TestAutoRotate", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)

	r, err := NewRoller(filename, 10, nil)
	isNil(err, t)
	defer r.Close()
	b := []byte("boo!")
	n, err := r.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	existsWithContent(filename, b, t)
	fileCount(dir, 1, t)

	newFakeTime()

	b2 := []byte("foooooo!")
	n, err = r.Write(b2)
	isNil(err, t)
	equals(len(b2), n, t)

	// the old logfile should be moved aside and the main logfile should have
	// only the last write in it.
	existsWithContent(filename, b2, t)

	fileCount(dir, 2, t)

	// the backup file will use the current fake time and have the old contents.
	existsWithContent(backupFile(dir), b, t)

}

func TestFirstWriteRotate(t *testing.T) {
	currentTime = fakeTime

	dir := makeTempDir("TestFirstWriteRotate", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)
	start := []byte("boooooo!")
	err := ioutil.WriteFile(filename, start, 0600)
	isNil(err, t)

	r, err := NewRoller(filename, 10, nil)
	isNil(err, t)

	defer r.Close()

	newFakeTime()

	// this would make us rotate
	b := []byte("fooo!")
	n, err := r.Write(b)
	isNil(err, t)
	equals(len(b), n, t)
	fileCount(dir, 2, t)

	existsWithContent(filename, b, t)
	existsWithContent(backupFile(dir), start, t)

}

func TestMaxBackups(t *testing.T) {
	currentTime = fakeTime

	dir := makeTempDir("TestMaxBackups", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)
	r, err := NewRoller(filename, 10, &Options{MaxBackups: 1})
	isNil(err, t)

	defer r.Close()
	b := []byte("boo!")
	n, err := r.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	existsWithContent(filename, b, t)
	fileCount(dir, 1, t)

	newFakeTime()

	// this will put us over the max
	b2 := []byte("foooooo!")
	n, err = r.Write(b2)
	isNil(err, t)
	equals(len(b2), n, t)

	// this will use the new fake time
	secondFilename := backupFile(dir)
	existsWithContent(secondFilename, b, t)

	// make sure the old file still exists with the same content.
	existsWithContent(filename, b2, t)

	fileCount(dir, 2, t)

	newFakeTime()

	// this will make us rotate again
	b3 := []byte("baaaaaar!")
	n, err = r.Write(b3)
	isNil(err, t)
	equals(len(b3), n, t)

	// this will use the new fake time
	thirdFilename := backupFile(dir)
	existsWithContent(thirdFilename, b2, t)

	existsWithContent(filename, b3, t)

	// we need to wait a little bit since the files get deleted on a different
	// goroutine.
	<-time.After(time.Millisecond * 10)

	// should only have two files in the dir still
	fileCount(dir, 2, t)

	// second file name should still exist
	existsWithContent(thirdFilename, b2, t)

	// should have deleted the first backup
	notExist(secondFilename, t)

	// now test that we don't delete directories or non-logfile files

	newFakeTime()

	// create a file that is close to but different from the logfile name.
	// It shouldn't get caught by our deletion filters.
	notlogfile := logFile(dir) + ".foo"
	err = ioutil.WriteFile(notlogfile, []byte("data"), 0644)
	isNil(err, t)

	// Make a directory that exactly matches our log file filters... it still
	// shouldn't get caught by the deletion filter since it's a directory.
	notlogfiledir := backupFile(dir)
	err = os.Mkdir(notlogfiledir, 0700)
	isNil(err, t)

	newFakeTime()

	// this will use the new fake time
	fourthFilename := backupFile(dir)

	// Create a log file that is/was being compressed - this should
	// not be counted since both the compressed and the uncompressed
	// log files still exist.
	compLogFile := fourthFilename + compressSuffix
	err = ioutil.WriteFile(compLogFile, []byte("compress"), 0644)
	isNil(err, t)

	// this will make us rotate again
	b4 := []byte("baaaaaaz!")
	n, err = r.Write(b4)
	isNil(err, t)
	equals(len(b4), n, t)

	existsWithContent(fourthFilename, b3, t)
	existsWithContent(fourthFilename+compressSuffix, []byte("compress"), t)

	// we need to wait a little bit since the files get deleted on a different
	// goroutine.
	<-time.After(time.Millisecond * 10)

	// We should have four things in the directory now - the 2 log files, the
	// not log file, and the directory
	fileCount(dir, 5, t)

	// third file name should still exist
	existsWithContent(filename, b4, t)

	existsWithContent(fourthFilename, b3, t)

	// should have deleted the first filename
	notExist(thirdFilename, t)

	// the not-a-logfile should still exist
	exists(notlogfile, t)

	// the directory
	exists(notlogfiledir, t)
}

func TestCleanupExistingBackups(t *testing.T) {
	// test that if we start with more backup files than we're supposed to have
	// in total, that extra ones get cleaned up when we rotate.

	currentTime = fakeTime

	dir := makeTempDir("TestCleanupExistingBackups", t)
	defer os.RemoveAll(dir)

	// make 3 backup files

	data := []byte("data")
	backup := backupFile(dir)
	err := ioutil.WriteFile(backup, data, 0644)
	isNil(err, t)

	newFakeTime()

	backup = backupFile(dir)
	err = ioutil.WriteFile(backup+compressSuffix, data, 0644)
	isNil(err, t)

	newFakeTime()

	backup = backupFile(dir)
	err = ioutil.WriteFile(backup, data, 0644)
	isNil(err, t)

	// now create a primary log file with some data
	filename := logFile(dir)
	err = ioutil.WriteFile(filename, data, 0644)
	isNil(err, t)

	l, err := NewRoller(filename, 10, &Options{MaxBackups: 1})
	isNil(err, t)
	defer l.Close()

	newFakeTime()

	b2 := []byte("foooooo!")
	n, err := l.Write(b2)
	isNil(err, t)
	equals(len(b2), n, t)

	// we need to wait a little bit since the files get deleted on a different
	// goroutine.
	<-time.After(time.Millisecond * 10)

	// now we should only have 2 files left - the primary and one backup
	fileCount(dir, 2, t)
}

func TestMaxAge(t *testing.T) {
	currentTime = fakeTime

	dir := makeTempDir("TestMaxAge", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)
	l, err := NewRoller(filename, 10, &Options{MaxAge: 24 * time.Hour})
	isNil(err, t)
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	existsWithContent(filename, b, t)
	fileCount(dir, 1, t)

	// two days later
	newFakeTime()

	b2 := []byte("foooooo!")
	n, err = l.Write(b2)
	isNil(err, t)
	equals(len(b2), n, t)
	existsWithContent(backupFile(dir), b, t)

	// we need to wait a little bit since the files get deleted on a different
	// goroutine.
	<-time.After(10 * time.Millisecond)

	// We should still have 2 log files, since the most recent backup was just
	// created.
	fileCount(dir, 2, t)

	existsWithContent(filename, b2, t)
	existsWithContent(backupFile(dir), b, t)

	// two days later
	newFakeTime()

	b3 := []byte("baaar!")
	n, err = l.Write(b3)
	isNil(err, t)
	equals(len(b3), n, t)
	existsWithContent(backupFile(dir), b2, t)

	// we need to wait a little bit since the files get deleted on a different
	// goroutine.
	<-time.After(10 * time.Millisecond)

	// We should have 2 log files - the main log file, and the most recent
	// backup.  The earlier backup is past the cutoff and should be gone.
	fileCount(dir, 2, t)

	existsWithContent(filename, b3, t)

	// we should have deleted the old file due to being too old
	existsWithContent(backupFile(dir), b2, t)
}

func TestOldLogFiles(t *testing.T) {
	currentTime = fakeTime

	dir := makeTempDir("TestOldLogFiles", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)
	data := []byte("data")
	err := ioutil.WriteFile(filename, data, 07)
	isNil(err, t)

	// This gives us a time with the same precision as the time we get from the
	// timestamp in the name.
	t1, err := time.Parse(backupTimeFormat, fakeTime().UTC().Format(backupTimeFormat))
	isNil(err, t)

	backup := backupFile(dir)
	err = ioutil.WriteFile(backup, data, 07)
	isNil(err, t)

	newFakeTime()

	t2, err := time.Parse(backupTimeFormat, fakeTime().UTC().Format(backupTimeFormat))
	isNil(err, t)

	backup2 := backupFile(dir)
	err = ioutil.WriteFile(backup2, data, 07)
	isNil(err, t)

	l, err := NewRoller(filename, 100, nil)
	isNil(err, t)
	files, err := l.oldLogFiles()
	isNil(err, t)
	equals(2, len(files), t)

	// should be sorted by newest file first, which would be t2
	equals(t2, files[0].timestamp, t)
	equals(t1, files[1].timestamp, t)
}

func TestTimeFromName(t *testing.T) {
	l := &Roller{filename: "/var/log/myfoo/foo.log"}
	prefix, ext := l.prefixAndExt()

	tests := []struct {
		filename string
		want     time.Time
		wantErr  bool
	}{
		{"foo-2014-05-04T14-44-33.555.log", time.Date(2014, 5, 4, 14, 44, 33, 555000000, time.UTC), false},
		{"foo-2014-05-04T14-44-33.555", time.Time{}, true},
		{"2014-05-04T14-44-33.555.log", time.Time{}, true},
		{"foo.log", time.Time{}, true},
	}

	for _, test := range tests {
		got, err := l.timeFromName(test.filename, prefix, ext)
		equals(got, test.want, t)
		equals(err != nil, test.wantErr, t)
	}
}

func TestLocalTime(t *testing.T) {
	currentTime = fakeTime

	dir := makeTempDir("TestLocalTime", t)
	defer os.RemoveAll(dir)

	l, err := NewRoller(logFile(dir), 10, &Options{LocalTime: true})
	isNil(err, t)
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	b2 := []byte("fooooooo!")
	n2, err := l.Write(b2)
	isNil(err, t)
	equals(len(b2), n2, t)

	existsWithContent(logFile(dir), b2, t)
	existsWithContent(backupFileLocal(dir), b, t)
}

func TestRotate(t *testing.T) {
	currentTime = fakeTime
	dir := makeTempDir("TestRotate", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)

	l, err := NewRoller(logFile(dir), 100, &Options{MaxBackups: 1})
	isNil(err, t)
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	existsWithContent(filename, b, t)
	fileCount(dir, 1, t)

	newFakeTime()

	err = l.Rotate()
	isNil(err, t)

	// we need to wait a little bit since the files get deleted on a different
	// goroutine.
	<-time.After(10 * time.Millisecond)

	filename2 := backupFile(dir)
	existsWithContent(filename2, b, t)
	existsWithContent(filename, []byte{}, t)
	fileCount(dir, 2, t)
	newFakeTime()

	err = l.Rotate()
	isNil(err, t)

	// we need to wait a little bit since the files get deleted on a different
	// goroutine.
	<-time.After(10 * time.Millisecond)

	filename3 := backupFile(dir)
	existsWithContent(filename3, []byte{}, t)
	existsWithContent(filename, []byte{}, t)
	fileCount(dir, 2, t)

	b2 := []byte("foooooo!")
	n, err = l.Write(b2)
	isNil(err, t)
	equals(len(b2), n, t)

	// this will use the new fake time
	existsWithContent(filename, b2, t)
}

func TestCompressOnRotate(t *testing.T) {
	currentTime = fakeTime

	dir := makeTempDir("TestCompressOnRotate", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)

	l, err := NewRoller(filename, 10, &Options{Compress: true})
	isNil(err, t)
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	existsWithContent(filename, b, t)
	fileCount(dir, 1, t)

	newFakeTime()

	err = l.Rotate()
	isNil(err, t)

	// the old logfile should be moved aside and the main logfile should have
	// nothing in it.
	existsWithContent(filename, []byte{}, t)

	// we need to wait a little bit since the files get compressed on a different
	// goroutine.
	<-time.After(300 * time.Millisecond)

	// a compressed version of the log file should now exist and the original
	// should have been removed.
	bc := new(bytes.Buffer)
	gz := gzip.NewWriter(bc)
	_, err = gz.Write(b)
	isNil(err, t)
	err = gz.Close()
	isNil(err, t)
	existsWithContent(backupFile(dir)+compressSuffix, bc.Bytes(), t)
	notExist(backupFile(dir), t)

	fileCount(dir, 2, t)
}

func TestCompressOnResume(t *testing.T) {
	currentTime = fakeTime

	dir := makeTempDir("TestCompressOnResume", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)

	// Create a backup file and empty "compressed" file to represent a file that
	// we started to compress but didn't finish.
	filename2 := backupFile(dir)
	b := []byte("foo!")
	err := ioutil.WriteFile(filename2, b, 0644)
	isNil(err, t)
	err = ioutil.WriteFile(filename2+compressSuffix, []byte{}, 0644)
	isNil(err, t)

	l, err := NewRoller(filename, 10, &Options{Compress: true})
	isNil(err, t)
	defer l.Close()

	newFakeTime()

	b2 := []byte("boo!")
	n, err := l.Write(b2)
	isNil(err, t)
	equals(len(b2), n, t)
	existsWithContent(filename, b2, t)

	// we need to wait a little bit since the files get compressed on a different
	// goroutine.
	<-time.After(300 * time.Millisecond)

	// The write should have started the compression - a compressed version of
	// the log file should now exist and the original should have been removed.
	bc := new(bytes.Buffer)
	gz := gzip.NewWriter(bc)
	_, err = gz.Write(b)
	isNil(err, t)
	err = gz.Close()
	isNil(err, t)
	existsWithContent(filename2+compressSuffix, bc.Bytes(), t)
	notExist(filename2, t)

	fileCount(dir, 2, t)
}

// makeTempDir creates a file with a semi-unique name in the OS temp directory.
// It should be based on the name of the test, to keep parallel tests from
// colliding, and must be cleaned up after the test is finished.
func makeTempDir(name string, t testing.TB) string {
	t.Helper()
	dir, err := os.MkdirTemp("", name)
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// existsWithContent checks that the given file exists and has the correct content.
func existsWithContent(path string, content []byte, t testing.TB) {
	t.Helper()
	info, err := os.Stat(path)
	isNil(err, t)
	equals(int64(len(content)), info.Size(), t)

	b, err := ioutil.ReadFile(path)
	isNil(err, t)
	equals(content, b, t)
}

// logFile returns the log file name in the given directory.
func logFile(dir string) string {
	return filepath.Join(dir, "foobar.log")
}

func backupFile(dir string) string {
	return filepath.Join(dir, "foobar-"+fakeTime().UTC().Format(backupTimeFormat)+".log")
}

func backupFileLocal(dir string) string {
	return filepath.Join(dir, "foobar-"+fakeTime().Format(backupTimeFormat)+".log")
}

// fileCount checks that the number of files in the directory is exp.
func fileCount(dir string, exp int, t testing.TB) {
	t.Helper()
	files, err := ioutil.ReadDir(dir)
	isNil(err, t)
	// Make sure no other files were created.
	equals(exp, len(files), t)
}

// newFakeTime sets the fake "current time" to two days later.
func newFakeTime() {
	fakeCurrentTime = fakeCurrentTime.Add(time.Hour * 24 * 2)
}

func notExist(path string, t testing.TB) {
	t.Helper()
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		t.Fatalf("expected to get os.IsNotExist, but instead got %v", err)
	}
}

func exists(path string, t testing.TB) {
	t.Helper()
	_, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file to exist, but got error from os.Stat: %v", err)
	}
}
