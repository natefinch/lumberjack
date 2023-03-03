package lumberjack

import (
	"log"
	"time"

	"github.com/natefinch/lumberjack/v3"
)

// To use lumberjack with the standard library's log package, just pass it into
// the SetOutput function when your application starts.
func Example() {
	l, _ := lumberjack.NewRoller(
		"/var/log/myapp/foo.log",
		500*1024*1024, // 500 megabytes
		&lumberjack.Options{
			MaxBackups: 3,
			MaxAge:     28 * time.Hour * 24, // 28 days
			Compress:   true,
		})
	log.SetOutput(l)
}
