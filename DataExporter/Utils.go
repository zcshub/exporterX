package dataExporter

import "os"

// returns whether the given file or directory exists
func pathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

// make sure path exists
func makePathExists(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	return err
}
