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

package v1

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVersion(t *testing.T) {
	v, err := ParseVersion("v5.7.38")
	assert.Equal(t, err, nil)
	assert.Equal(t, v.Major, 5)
	assert.Equal(t, v.Minor, 7)
	assert.Equal(t, v.NoPatch, false)
	assert.Equal(t, v.Patch, 38)

	v, err = ParseVersion("8.0")
	assert.Equal(t, err, nil)
	assert.Equal(t, v.Major, 8)
	assert.Equal(t, v.Minor, 0)
	assert.Equal(t, v.NoPatch, true)
	assert.Equal(t, v.Patch, 29)
}

func TestHasQuote(t *testing.T) {
	ok := HasQuote("123", "abc")
	assert.Equal(t, ok, false)

	ok = HasQuote("123'", "a`bc")
	assert.Equal(t, ok, true)
}

func TestHasEqual(t *testing.T) {
	ok := HasEqual(123, 456)
	assert.Equal(t, ok, false)

	ok = HasEqual(123, 123)
	assert.Equal(t, ok, true)
}
func TestSplitHostPort(t *testing.T) {
	_, _ = SplitHostPort("[abc]:123")
}

func TestMysqlGroupReplicationLocalAddress(t *testing.T) {
	mysql := new(Mysql)
	mysql.Default()
	mysql.Validate()
	mysql.GroupReplicationGroupSeeds()
	s, err := mysql.NormalizeSolo(0)
	assert.Equal(t, err, nil)
	s.GroupReplicationLocalAddress()
	mysql.DeepCopy()
	mysql.NamespacedName()
	mysql.NewLabels()

	mysql = new(Mysql)
	mysql.Spec.PrimaryMode = ModeMulti
	mysql.Spec.EnableExporter = true
	mysql.Default()
	mysql.Validate()

	out := new(Mysql)
	mysql.DeepCopyInto(out)
}

func TestWrite(t *testing.T) {
	b := new(bytes.Buffer)
	WriteString(b, "123")
	WriteByte(b, 'b')
	WriteInt(b, 123)
}
