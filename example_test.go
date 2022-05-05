package lumberjack

import (
	"log"
)

// To use lumberjack with the standard library's log package, just pass it into
// the SetOutput function when your application starts.
func Example() {
	log.SetOutput(&Logger{
		Filename:   "/var/log/myapp/foo.log",
		MaxBytes:   500, // bytes
		MaxBackups: 3,
		MaxAge:     28,   // days
		Compress:   true, // disabled by default
	})
}
