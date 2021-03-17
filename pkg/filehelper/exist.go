package filehelper

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// CheckExist please check error is nil or not
func CheckExist(path string, needDir bool) error {
	f, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Wrap(err, fmt.Sprintf("%s not exist", path))
		}
		return errors.Wrap(err,
			fmt.Sprintf("%s does exist, but throw other error when checking", path))
	}
	if needDir {
		if f.IsDir() {
			return nil
		} else {
			return errors.Errorf("%s exist, but is not a directory", path)
		}
	}
	return nil
}
