package utils

import (
	"fmt"
	"os"
)

func GetLocalStoragePath(suffix string) string {
	currentDir, _ := os.Getwd()
	return fmt.Sprintf("%s/%s/%s", currentDir, LocalStorageDir, suffix)
}
