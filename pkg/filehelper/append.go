package filehelper

import (
	"fmt"
	"os"
)

// Append append content to file. Create file if not exists.
func Append(filename, content string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for append, filename: %s, err: %v", filename, err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("failed to append to file, filename: %s, err: %v", filename, err)
	}
	return nil
}
