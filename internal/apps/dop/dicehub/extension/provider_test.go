// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package extension

import (
	"os"
	"path"
	"testing"

	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/pkg/mock"
)

func Test_provider(t *testing.T) {
	pv := &provider{Cfg: &config{ExtensionMenu: nil}, ExtensionSvc: &mock.MockExtensionInterface{}}

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
	err = pv.ExtensionSvc.InitSources()
	assert.NoError(t, err)
}
