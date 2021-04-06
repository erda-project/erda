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

package license

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

var testAesKey = "0123456789abcdef"
var testAesEncryptText = "AL5riQzgwmMhXZX+5MtU0A=="
var testOriginText = "123456"

func TestAESEncrypt(t *testing.T) {
	s, err := AesEncrypt(testOriginText, testAesKey)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, s, testAesEncryptText)
}

func TestAESDecrypt(t *testing.T) {
	s, err := AesDecrypt(testAesEncryptText, testAesKey)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, s, testOriginText)
}
