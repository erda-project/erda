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
