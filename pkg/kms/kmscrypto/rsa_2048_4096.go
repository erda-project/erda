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

package kmscrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

type RSACrypt struct {
	secretInfo RSASecret
}

type RSASecret struct {
	PublicKey          string
	PublicKeyDataType  Encode
	PrivateKey         string
	PrivateKeyDataType Encode
	PrivateKeyType     Secret
}

// NewRSACrypt init with the RSA secret info
func NewRSACrypt(secretInfo RSASecret) *RSACrypt {
	return &RSACrypt{secretInfo: secretInfo}
}

// Encrypt encrypts the given message with public key
// src the original data
// outputDataType the encoded type of encrypted data ,such as Base64,HEX
func (rc *RSACrypt) Encrypt(src string, outputDataType Encode) (dst string, err error) {
	secretInfo := rc.secretInfo
	if secretInfo.PublicKey == "" {
		return "", fmt.Errorf("secretInfo PublicKey can't be empty")
	}
	pubKeyDecoded, err := DecodeString(secretInfo.PublicKey, secretInfo.PublicKeyDataType)
	if err != nil {
		return
	}
	block, _ := pem.Decode(pubKeyDecoded) // decode
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return
	}
	var dataEncrypted []byte
	dataEncrypted, err = rsa.EncryptPKCS1v15(rand.Reader, pubKey.(*rsa.PublicKey), []byte(src))
	if err != nil {
		return
	}

	return EncodeToString(dataEncrypted, outputDataType)
}

// ParsePrivateKey parses private key bytes to rsa privateKey
func ParsePrivateKey(privateKeyDecoded []byte, keyType Secret) (*rsa.PrivateKey, error) {
	switch keyType {
	case PKCS1:
		return x509.ParsePKCS1PrivateKey(privateKeyDecoded)
	case PKCS8:
		keyParsed, err := x509.ParsePKCS8PrivateKey(privateKeyDecoded)
		return keyParsed.(*rsa.PrivateKey), err
	default:
		return &rsa.PrivateKey{}, fmt.Errorf("secretInfo PrivateKeyDataType unsupport")
	}
}

// Decrypt decrypts a plaintext using private key
// src the encrypted data with public key
// srcType the encoded type of encrypted data ,such as Base64,HEX
func (rc *RSACrypt) Decrypt(src string, srcType Encode) (dst string, err error) {
	secretInfo := rc.secretInfo
	if secretInfo.PrivateKey == "" {
		return "", fmt.Errorf("secretInfo PrivateKey can't be empty")
	}
	privateKeyDecoded, err := DecodeString(secretInfo.PrivateKey, secretInfo.PrivateKeyDataType)
	if err != nil {
		return
	}
	block, _ := pem.Decode(privateKeyDecoded) // decode
	prvKey, err := ParsePrivateKey(block.Bytes, secretInfo.PrivateKeyType)
	if err != nil {
		return
	}
	decodeData, err := DecodeString(src, srcType)
	if err != nil {
		return
	}
	var dataDecrypted []byte
	dataDecrypted, err = rsa.DecryptPKCS1v15(rand.Reader, prvKey, decodeData)
	if err != nil {
		return
	}
	return string(dataDecrypted), nil
}

// GenRsaKey return publicKey, privateKey, error
func GenRsaKey(bits int) ([]byte, []byte, error) {
	// generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	// serialize the obtained RAS private key through the X509 standard as: ASN.1 DER encoded string
	privateStream := x509.MarshalPKCS1PrivateKey(privateKey)
	// set private key string into pem format block
	block1 := pem.Block{
		Type:  "private key",
		Bytes: privateStream,
	}
	// encode the set data through pem
	privateKeyBytes := pem.EncodeToMemory(&block1)

	// public key
	publicKey := privateKey.PublicKey
	publicStream, err := x509.MarshalPKIXPublicKey(&publicKey)
	block2 := pem.Block{
		Type:  "public key",
		Bytes: publicStream,
	}
	publicKeyBytes := pem.EncodeToMemory(&block2)

	return publicKeyBytes, privateKeyBytes, nil
}
