package dal

import "os"

func IsDirExist(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !info.Mode().IsDir() {
		return false
	}
	return true
}

func IsRegularFileExist(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !info.Mode().IsRegular() {
		return false
	}
	return true
}

func CreateRegularFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
