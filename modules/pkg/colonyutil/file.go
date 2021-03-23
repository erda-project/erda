package colonyutil

import (
	"crypto/md5" // #nosec G501
	"encoding/hex"
	"io"
	"os"
	"strings"
)

func CheckMd5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New() // #nosec G401
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func GetFileName(path string) string {
	pathList := strings.Split(path, "/")
	return pathList[len(pathList)-1]
}

func GetFileSize(path string) (int64, error) {
	f, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return f.Size(), nil
}
