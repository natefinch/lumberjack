// +build linux

package lumberjack

import (
	"os"
	"syscall"
	"testing"
)

func TestMaintainOwner(t *testing.T) {
	fakeC := fakeChown{}
	os_Chown = fakeC.Set
	os_Stat = fakeStat
	defer func() {
		os_Chown = os.Chown
		os_Stat = os.Stat
	}()
	currentTime = fakeTime
	dir := makeTempDir("TestMaintainOwner", t)
	defer os.RemoveAll(dir)

	filename := logFile(dir)

	l := &Logger{
		Filename:   filename,
		MaxBackups: 1,
		MaxSize:    100, // megabytes
	}
	defer l.Close()
	b := []byte("boo!")
	n, err := l.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	newFakeTime()

	err = l.Rotate()
	isNil(err, t)

	equals(555, fakeC.uid, t)
	equals(666, fakeC.gid, t)
}

type fakeChown struct {
	name string
	uid  int
	gid  int
}

func (f *fakeChown) Set(name string, uid, gid int) error {
	f.name = name
	f.uid = uid
	f.gid = gid
	return nil
}

func fakeStat(name string) (os.FileInfo, error) {
	info, err := os.Stat(name)
	if err != nil {
		return info, err
	}
	stat := info.Sys().(*syscall.Stat_t)
	stat.Uid = 555
	stat.Gid = 666
	return info, nil
}
