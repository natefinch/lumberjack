
# lumberjack  [![GoDoc](https://godoc.org/github.com/natefinch/lumberjack?status.png)](https://godoc.org/github.com/natefinch/lumberjack) [![Build Status](https://travis-ci.org/natefinch/lumberjack.png)](https://travis-ci.org/natefinch/lumberjack)

### Lumberjack is a Go package for writing logs to rolling files.

Lumberjack is intended to be one part of a logging infrastructure.
It is not an all-in-one solution, but instead is a pluggable
component at the bottom of the logging stack that simply controls the files
to which logs are written.

Lumberjack plays well with any logger that can write to an io.Writer,
including the standard library's log package.

For example, to use lumberjack with the std lib's log package, just pass it
into the SetOutput function when your application starts:


	log.SetOutput(&lumberjack.Logger{
	    Dir: "/var/log/myapp/"
	    NameFormat: time.RFC822+".log",
	    MaxSize: lumberjack.Gigabyte,
	    MaxBackups: 3,
	    MaxAge: lumberjack.Week * 4,
	))

Lumberjack assumes that only one process is writing to the output files.
Using the same lumberjack configuration from multiple processes on the same
machine will result in improper behavior.




## Constants
``` go
const (
    Megabyte = 1024 * 1024
    Gigabyte = 1024 * Megabyte

    Day  = 24 * time.Hour
    Week = 7 * Day
)
```



## type Logger
``` go
type Logger struct {
    // Dir determines the directory in which to store log files.
    // It defaults to os.TempDir() if empty.
    Dir string `json:"dir" yaml:"dir"`

    // NameFormat is the time formatting layout used to generate filenames.
    // It defaults to "2006-01-02T15-04-05.000.log".
    NameFormat string `json:"nameformat" yaml:"nameformat"`

    // MaxSize is the maximum size in bytes of the log file before it gets
    // rolled. It defaults to 100 megabytes.
    MaxSize int64 `json:"maxsize" yaml:"maxsize"`

    // MaxAge is the maximum time to retain old log files based on
    // FileInfo.ModTime.  The default is not to remove old log files based on
    // age.
    MaxAge time.Duration `json:"maxage" yaml:"maxage"`

    // MaxBackups is the maximum number of old log files to retain.  The default
    // is to retain all old log files (though MaxAge may still cause them to get
    // deleted.)
    MaxBackups int `json:"maxbackups" yaml:"maxbackups"`

    // LocalTime determines if the time used for formatting the filename is the
    // computer's local time.  The default is to use UTC time.
    LocalTime bool `json:"localtime" yaml:"localtime"`
    // contains filtered or unexported fields
}
```
Logger is an io.WriteCloser that writes to a log file in the given directory
with the given NameFormat.  NameFormat should include a time formatting
layout in it that produces a valid unique filename for the OS.  For more
about time formatting layouts, read a href="http://golang.org/pkg/time/#pkg-constants">http://golang.org/pkg/time/#pkg-constants</a>.

The date encoded in the filename by NameFormat is used to determine which log
files are most recent in several situations.

Logger opens or creates a logfile on first Write.  It looks for files in the
directory that match its name format, and if the one with the most recent
NameFormat date is less than MaxSize, it will open and append to that file.
If no such file exists, or the file is >= MaxSize, a new file is created
using the current time with NameFormat to generate the filename.

Whenever a write would cause the current log file exceed MaxSize, a new file
is created using the current time.

### Cleaning Up Old Log Files
Whenever a new file gets created, old log files may be deleted.  The log file
directory is scanned for files that match NameFormat.  The most recent files
according to their NameFormat date will be retained, up to a number equal to
MaxBackups (or all of them if MaxBackups is 0).  Any files with a last
modified time (based on FileInfo.ModTime) older than MaxAge are deleted,
regardless of MaxBackups.

If MaxBackups and MaxAge are both 0, no old log files will be deleted.











### func (\*Logger) Close
``` go
func (l *Logger) Close() error
```
Close implements io.Closer, and closes the current logfile.



### func (\*Logger) Rotate
``` go
func (l *Logger) Rotate() error
```
Rotate causes Logger to close the existing log file and immediately create a
new one.  This is a helper function for applications that want to initiate
rotations outside of the normal rotation rules, such as in response to
SIGHUP.  After rotating, this initiates a cleanup of old log files according
to the normal rules.


#### Example

Example of how to rotate in response to SIGHUP.

Code:

  l := &lumberjack.Logger{}
  log.SetOutput(l)
  c := make(chan os.Signal, 1)
  signal.Notify(c, syscall.SIGHUP)

  go func() {
      for {
          <-c
          l.Rotate()
      }
  }()



### func (\*Logger) Write
``` go
func (l *Logger) Write(p []byte) (n int, err error)
```
Write implements io.Writer.  If a write would cause the log file to be larger
than MaxSize, a new log file is created using the current time formatted with
PathFormat.  If the length of the write is greater than MaxSize, an error is
returned that satisfies IsWriteTooLong.









- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
