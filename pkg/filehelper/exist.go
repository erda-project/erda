// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
