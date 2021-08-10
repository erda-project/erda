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

package kmscrypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/crypto/uuid"
)

func TestAES(t *testing.T) {
	key, err := GenerateAes256Key()
	assert.NoError(t, err)
	plaintext := []byte("hello world")

	cmk := uuid.UUID()

	// encrypt
	ciphertext, err := AesGcmEncrypt(key, plaintext, []byte(cmk))
	assert.NoError(t, err)
	fmt.Printf("ciphertext: %x\n", ciphertext)

	// decrypt
	decrypted, err := AesGcmDecrypt(key, ciphertext, []byte(cmk))
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestWrapWithLenPrefix(t *testing.T) {
	b := []byte("hello world")
	bb, err := PrefixAppend000Length(b)
	assert.NoError(t, err)
	assert.Equal(t, "011"+string(b), string(bb))

	under, remains, err := PrefixUnAppend000Length(bb)
	assert.NoError(t, err)
	assert.Equal(t, string(b), string(under))
	assert.Equal(t, 0, len(remains))
}
