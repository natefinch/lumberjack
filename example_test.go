package lumberjack_test

import (
	"log"

	"github.com/natefinch/lumberjack"
)

// To use lumberjack with the standard library's log package, just pass it into
// the SetOutput function when your application starts.
func Example() {
	log.SetOutput(&lumberjack.Logger{
		Dir:        "/var/log/myapp/",
		NameFormat: "2006-01-02T15-04-05.000.log",
		MaxSize:    lumberjack.Gigabyte,
		MaxBackups: 3,
		MaxAge:     28,
	})
}
