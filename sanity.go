package lumberjack

import (
	"fmt"
	"path/filepath"
	"time"
	"os"
	"io/ioutil"
	"sort"
	"strings"
	"errors"
)

const (
	CompressSuffix = ".gz"
)

var (
	DEFAULT_ROTATE_FILE_PATTERN = &rotateFilePattern{}
)

type RotateFilePattern interface {
	BackupName(name string, localTime bool) string
	OldLogFiles(name, dir string) ([]*LogInfo, error)
}

type rotateFilePattern struct{}

// backupName creates a new filename from the given name, inserting a timestamp
// between the filename and the extension, using the local time if requested
// (otherwise UTC).
func (p *rotateFilePattern) BackupName(name string, local bool) string {
	dir := filepath.Dir(name)
	filename := filepath.Base(name)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]
	t := currentTime()
	if !local {
		t = t.UTC()
	}

	timestamp := t.Format(backupTimeFormat)
	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", prefix, timestamp, ext))
}

// oldLogFiles returns the list of backup log files stored in the same
// directory as the current log file, sorted by ModTime
func (p *rotateFilePattern) OldLogFiles(name, dir string) ([]*LogInfo, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("can't read log file directory: %s", err)
	}
	logFiles := []*LogInfo{}

	prefix, ext := prefixAndExt(name)

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if t, err := timeFromName(f.Name(), prefix, ext); err == nil {
			logFiles = append(logFiles, &LogInfo{t, f})
			continue
		}
		if t, err := timeFromName(f.Name(), prefix, ext+CompressSuffix); err == nil {
			logFiles = append(logFiles, &LogInfo{t, f})
			continue
		}
		// error parsing means that the suffix at the end was not generated
		// by lumberjack, and therefore it's not a backup file.
	}

	sort.Sort(ByFormatTime(logFiles))

	return logFiles, nil
}

// timeFromName extracts the formatted time from the filename by stripping off
// the filename's prefix and extension. This prevents someone's filename from
// confusing time.parse.
func timeFromName(filename, prefix, ext string) (time.Time, error) {
	if !strings.HasPrefix(filename, prefix) {
		return time.Time{}, errors.New("mismatched prefix")
	}
	if !strings.HasSuffix(filename, ext) {
		return time.Time{}, errors.New("mismatched extension")
	}
	ts := filename[len(prefix): len(filename)-len(ext)]
	return time.Parse(backupTimeFormat, ts)
}

// prefixAndExt returns the filename part and extension part from the Logger's
// filename.
func prefixAndExt(name string) (prefix, ext string) {
	filename := filepath.Base(name)
	ext = filepath.Ext(filename)
	prefix = filename[:len(filename)-len(ext)] + "-"
	return prefix, ext
}

// logInfo is a convenience struct to return the filename and its embedded
// timestamp.
type LogInfo struct {
	Timestamp time.Time
	os.FileInfo
}

// byFormatTime sorts by newest time formatted in the name.
type ByFormatTime []*LogInfo

func (b ByFormatTime) Less(i, j int) bool {
	return b[i].Timestamp.After(b[j].Timestamp)
}

func (b ByFormatTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b ByFormatTime) Len() int {
	return len(b)
}
