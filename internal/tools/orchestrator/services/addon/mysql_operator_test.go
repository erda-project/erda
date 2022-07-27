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

package addon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getMyletHost(t *testing.T) {
	s := getMyletHost("abc")
	assert.Equal(t, "abc:33080", s)
}

func Test_getToken(t *testing.T) {
	s := getToken("abc-write", "123")
	assert.Equal(t, "abc-myctl:0@726f6f743a31323340616263e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", s)
}

func Test_createUserDB(t *testing.T) {
	err := createUserDB("user", "pass", "db", "erda.cloud", "key")
	assert.NotEqual(t, err, nil)
}

func Test_runSQL(t *testing.T) {
	err := runSQL("user", "pass", "db", "run.sql.gz", "erda.cloud", "key")
	assert.NotEqual(t, err, nil)
}
