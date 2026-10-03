package lib

import (
	"os"
)

// GetFile opens the file at the specified location and returns a pointer to the os.File object and an error if any occurred during the opening process.
func GetFile(location string) (*os.File, error) {
	return os.Open(location)
}
