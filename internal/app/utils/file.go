package utils

import (
	"os"
	"path/filepath"
)

func WriteFileWithDir(path, filename string, data []byte, perm os.FileMode) error {
	// Ensure the directory exists
	fullPath := filepath.Join(path, filename)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, perm); err != nil {
		return err
	}

	// Open the file with the intention to create it if it doesn't exist,
	// truncate it if it does, and ensure it's closed after the function executes.
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer file.Close() // Ensures file is closed even if the next write fails

	// Write data to file
	_, err = file.Write(data)
	return err
}
