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

package bundle_test

import (
	"testing"

	"github.com/erda-project/erda/bundle"
)

func TestNewGittarFileTree(t *testing.T) {
	if _, err := bundle.NewGittarFileTree(""); err != nil {
		t.Error(err)
	}
	if _, err := bundle.NewGittarFileTree("some-error-string"); err == nil {
		t.Error("can not NewGittarFileTree on error string")
	}
	var inode = "QT01ODgwJlA9Mzg3JmE9ZXJkYSZiPW1hc3RlciZwPWVyZGEtcHJvamVjdA=="
	ft, err := bundle.NewGittarFileTree(inode)
	if err != nil {
		t.Fatal(err)
	}
	for k := range ft.Values {
		t.Log(k, ft.Values[k])
	}
}
