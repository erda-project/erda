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

package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
)

func (a *AIProxyProviders) ResetAesKey() {
	key := make([]byte, 8)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	a.AesKey = hex.EncodeToString(key)
}

func (a *AIProxyProviders) SetAPIKeyWithEncrypt(key string) {
	if len(a.AesKey) != 16 {
		a.ResetAesKey()
	}
	block, err := aes.NewCipher([]byte(a.AesKey))
	if err != nil {
		panic(err)
	}
	ciphertext := make([]byte, aes.BlockSize+len(key))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(key))
	a.APIKey = base64.URLEncoding.EncodeToString(ciphertext)
}

func (a *AIProxyProviders) GetAPIKeyWithDecrypt() string {
	ciphertext, err := base64.URLEncoding.DecodeString(a.APIKey)
	if err != nil {
		panic(err)
	}
	block, err := aes.NewCipher([]byte(a.AesKey))
	if err != nil {
		panic(err)
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext)
}
