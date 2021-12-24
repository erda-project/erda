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

package tar_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/erda-project/erda/pkg/archive/tar"
)

func TestNew(t *testing.T) {
	var buf = bytes.NewBuffer(nil)
	tape := tar.New(buf)
	defer tape.Close()
	if _, err := tape.Write("a.sql", 777, []byte("create t1 (id bigint);")); err != nil {
		t.Fatal(err)
	}
	if _, err := tape.Write("sqls/b.sql", 777, []byte("alter table t1 add column col1 varchar(64);")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("the-tar.tar", buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}
