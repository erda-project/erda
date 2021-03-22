package kmscrypto

import (
	"fmt"
	"testing"

	"github.com/erda-project/erda/pkg/uuid"

	"github.com/stretchr/testify/assert"
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
