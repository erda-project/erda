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
	"errors"
	"os"
	"path"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/modules/dicehub/extension/db"
)

func Test_provider(t *testing.T) {
	pv := &provider{Cfg: &config{ExtensionMenu: "{}"}}
	pv.newExtensionService()
	cl := &db.ExtensionConfigDB{}
	p := &extensionService{
		db: cl,
	}
	fc := monkey.PatchInstanceMethod(reflect.TypeOf(cl), "QueryAllExtensions", func(_ *db.ExtensionConfigDB) ([]db.ExtensionVersion, error) {
		return nil, nil
	})
	defer fc.Unpatch()

	dir := path.Join(os.TempDir(), "extension")
	defer func() {
		err := os.RemoveAll(dir)
		assert.NoError(t, err)
	}()
	dir1 := path.Join(dir, "test1", "actions", "one")
	err := os.MkdirAll(dir1, os.ModePerm)
	assert.NoError(t, err)
	f, err := os.Create(path.Join(dir1, "spec.yaml"))
	assert.NoError(t, err)
	specYml := `name: api-publish
	version: "1.0"
	type: action`
	_, err = f.Write([]byte(specYml))
	assert.NoError(t, err)

	fc2 := monkey.PatchInstanceMethod(reflect.TypeOf(p), "RunExtensionsPush", func(_ *extensionService, dir string, extensionVersionMap, extensionTypeMap map[string][]string) (string, string, error) {
		if dir != dir1 {
			return "", "", errors.New("path wrong")
		}
		return "", "", nil
	})
	defer fc2.Unpatch()
	err = p.InitExtension(FilePath)
	assert.NoError(t, err)
}
