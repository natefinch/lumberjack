// +build !linux

package lumberjack

import (
	"os"
	"syscall"
)

func chown(_ string, _ os.FileInfo) error {
	return nil
}
