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

package extension

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/alecthomas/assert"
)

func Test_action_isThereSpecFile(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "*")
	defer os.RemoveAll(dir)
	assert.NoError(t, err)

	_, v := isThereSpecFile(dir)
	assert.Equal(t, v, false)

	os.Create(path.Join(dir, "spec.yaml"))
	assert.NoError(t, err)
	_, v = isThereSpecFile(dir)
	assert.Equal(t, v, true)
}

func Test_action_LoadExtensions(t *testing.T) {
	dir := path.Join(os.TempDir(), "extension")
	dir1 := path.Join(dir, "test1", "actions", "one")
	dir2 := path.Join(dir, "test2", "actions", "two")

	err := os.MkdirAll(dir1, os.ModePerm)
	assert.NoError(t, err)
	err = os.MkdirAll(dir2, os.ModePerm)
	assert.NoError(t, err)
	defer func() {
		err := os.RemoveAll(dir)
		assert.NoError(t, err)
	}()

	f, err := os.Create(path.Join(dir1, "spec.yaml"))
	assert.NoError(t, err)
	specYml := `name: api-publish
version: "1.0"
type: action`
	_, err = f.Write([]byte(specYml))
	assert.NoError(t, err)

	tests := &Repo{
		addr:     dir,
		versions: []string{dir1},
	}

	r := LoadExtensions(dir)
	assert.Equal(t, tests, r)
}
