package lumberjack

import (
	"os"
)

// log roation under linux by process running as root gets chown not permitted errors so turn it off
func chown(_ string, _ os.FileInfo) error {
	return nil
}
