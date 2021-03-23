package filehelper

import (
	"fmt"
	"path/filepath"
	"strings"
)

func Abs2Rel(path string) string {
	path = filepath.Clean(path)
	if strings.HasPrefix(path, "/") {
		path = fmt.Sprintf(".%s", path)
	}
	return filepath.Clean(path)
}
