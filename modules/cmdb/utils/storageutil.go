package utils

import (
	"strings"
)

// IsOSS 是否为对象存储
func IsOSS(storageURL string) bool {
	return strings.HasPrefix(storageURL, "oss")
}
