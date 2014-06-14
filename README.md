# lumberjack

Lumberjack is still a work in progress, use at your own risk.


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
	    Backups: 3,
	    MaxAge: lumberjack.Week * 4,
	))

Note that lumberjack assumes whatever is writing to it will use locks to
prevent concurrent writes. Lumberjack does not implement its own lock.




## Constants
``` go
const (
    // Some helper constants to make your declarations easier to read.
    Megabyte = 1024 * 1024
    Gigabyte = 1024 * Megabyte

    // note that lumberjack days and weeks may not exactly conform to calendar
    // days and weeks due to daylight savings, leap seconds, etc.
    Day  = 24 * time.Hour
    Week = 7 * Day
)
```


## func IsWriteTooLong
``` go
func IsWriteTooLong(err error) bool
```
IsWriteTooLong returns whether the given error reports a write to Logger that
exceeds the Logger's MaxSize.



## type Logger
``` go
type Logger struct {
    // Dir determines the directory in which to store log files.
    // It defaults to os.TempDir() if empty.
    Dir string

    // NameFormat is the time formatting layout used to generate filenames.
    // It defaults to "2006-01-02T15-04-05.000.log".
    NameFormat string

    // MaxSize is the maximum size in bytes of the log file before it gets
    // rolled. It defaults to 100 megabytes.
    MaxSize int64

    // MaxAge is the maximum time to retain old log files.  The default is not
    // to remove old log files based on age.
    MaxAge time.Duration

    // Backups is the maximum number of old log files to retain.  The default is
    // to retain all old log files (though MaxAge may still cause them to get
    // deleted.)
    Backups int

    // LocalTime determines if the time used for formatting the filename is the
    // computer's local time.  The default is to use UTC time.
    LocalTime bool
    // contains filtered or unexported fields
}
```
Logger is an io.WriteCloser that writes to a log file in the given directory
with the given NameFormat.  NameFormat should include a time formatting
layout in it that produces a valid filename for the OS.  For more about time
formatting layouts, read a href="http://golang.org/pkg/time/#pkg-constants">http://golang.org/pkg/time/#pkg-constants</a>.

Logger opens or creates the logfile on first Write.  If the most recently
modified file in the log file directory that matches the NameFormat is less
than MaxSize, that file will be appended to.  If no file such exists, a new
file is created using the current time to generate the filename.

Whenever a write would cause the current log file exceed MaxSize, a new file
is created using the current time.

### Cleaning Up Old Log Files
Whenever a new file gets created, old log files may be deleted.  The log file
directory is scanned for files that match NameFormat.  The most recently
modified files which are newer than MaxAge (up to a number of files equal to
Backups) are retained, all other log files are deleted.

### Defaults
If Dir is empty, the files will be created in os.TempDir().

If NameFormat is empty,  will be used as the
name format.

If MaxSize is 0, 100 megabytes will be used as the max size.

if MaxAge is 0, last modification time will not be used to delete old log
files.

If Backups is 0, there's no limit to the number of old log files that will be
retained, as long as they're newer than MaxAge.

If MaxAge and Backups are both 0, no old log files will be deteled.

Thus, an default lumberjack.Logger struct will log to os.TempDir() with a 100
megabyte max size and never delete old log files.











### func (\*Logger) Close
``` go
func (l *Logger) Close() error
```
Close implements io.Closer, and closes the current logfile.



### func (\*Logger) Write
``` go
func (l *Logger) Write(p []byte) (n int, err error)
```
Write implements io.Writer.  If a write would cause the log file to be larger
than MaxSize, a new log file is created using the current time formatted with
PathFormat.  If the length of the write is greater than MaxSize, an error is
returned that satisfies IsWriteTooLong.



# License
MIT License
