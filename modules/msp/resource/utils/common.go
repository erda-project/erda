/*
 * // Copyright (c) 2021 Terminus, Inc.
 * //
 * // This program is free software: you can use, redistribute, and/or modify
 * // it under the terms of the GNU Affero General Public License, version 3
 * // or later ("AGPL"), as published by the Free Software Foundation.
 * //
 * // This program is distributed in the hope that it will be useful, but WITHOUT
 * // ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * // FITNESS FOR A PARTICULAR PURPOSE.
 * //
 * // You should have received a copy of the GNU Affero General Public License
 * // along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package utils

import (
	"encoding/base64"
	"encoding/json"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/kms/kmscrypto"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

// GetRandomId 生成随机33位uuid，并且，（首字母开头 + 32位uuid）构成Id
func GetRandomId() string {
	str := "abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 1; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return strings.Join([]string{string(result), uuid.UUID()}, "")
}

func JsonConvertObjToType(source interface{}, dest interface{}) error {
	destType := reflect.TypeOf(dest).String()
	if str, ok := source.(string); ok && destType != "*string" {
		bytes := []byte(str)
		err := json.Unmarshal(bytes, dest)
		return err
	}

	if bytes, err := json.Marshal(source); err != nil {
		return err
	} else {
		err = json.Unmarshal(bytes, dest)
		return err
	}
}

func JsonConvertObjToString(source interface{}) (string, error) {
	bytes, err := json.Marshal(source)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func AppendMap(base map[string]string, toadd map[string]string) {
	if toadd == nil {
		return
	}
	for k, v := range toadd {
		base[k] = v
	}
}

// AesDecrypt decode base64 data with AES
func AesDecrypt(data string) string {
	// decrypt using key="terminus-dice@20" with mode AES/GCM/NoPadding
	debase64Data, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return data
	}

	key := "terminus-dice@20"
	plainText, err := kmscrypto.AesGcmDecrypt([]byte(key), debase64Data, nil)
	if err != nil {
		return data
	}

	return string(plainText)
}
