package woodpecker

import (
	"log"
)

// To use woodpecker with the standard library's log package, just pass it into
// the SetOutput function when your application starts.
func Example() {
	log.SetOutput(&Logger{
		cfg: Config{
			Filename:       "/var/log/myapp/foo.log",
			RotateEveryday: true,
			MaxSizeInMb:    500, // megabytes
			MaxBackups:     3,
			MaxAgeInDays:   28,   // days
			Compress:       true, // disabled by default
		},
	})
}
