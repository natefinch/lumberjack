// +build linux

package lumberjack

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

// Example of how to rotate in response to SIGHUP.
func TestExampleLogger_Rotate(t *testing.T) {

	t.FailNow()
	l := &Logger{
		Filename: "./t.log"
		BackupDir: "./bk"
		MaxSize: 200
	}
	log.SetOutput(l)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for {
			<-c
			l.Rotate()
		}
	}()
}
