// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"math/rand"
	"unsafe"

	uuid "github.com/satori/go.uuid"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	APPLICATION_CONFIG_PATH = "conf/application.yml"
	TEMPLATE_JSON_PATH      = "conf/template.json"
	IPDB_PATH               = "conf/ipdata.dat"
	TA_JS_PATH              = "static/ta.js"
	TRANSLATE_PATH          = "conf/translate.yml"
)

func UUID() (string, error) {
	u := uuid.NewV4()
	return u.String(), nil
}

func InString(s string, ss []string) bool {
	for _, item := range ss {
		if s == item {
			return true
		}
	}
	return false
}

func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
