package lib

import (
	"os"
)

func GetFile(location string) (*os.File, error) {
	return os.Open(location)
}
