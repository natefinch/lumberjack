// +build !linux

package woodpecker

import (
	"os"
)

func chown(_ string, _ os.FileInfo) error {
	return nil
}
