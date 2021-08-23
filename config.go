package woodpecker

type Config struct {
	// Filename is the file to write logs to.
	// Backup log files will be retained in the same directory.
	Filename string `default:"log.log"`

	// RotateEveryday is flag which told woodpecker to rotate file every day at 00:00.
	RotateEveryday bool `default:"false"`

	// MaxSize is the maximum size in megabytes of the log file before it gets rotated.
	MaxSizeInMb int `default:"10"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc.
	MaxAgeInDays int `default:"30"`

	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int `default:"50"`

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.
	LocalTime bool `default:"false"`

	// Compress determines if the rotated log files should be compressed using gzip.
	Compress bool `default:"false"`
}
