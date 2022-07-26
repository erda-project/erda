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
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	MinPatchVersion57 = 38
	MinPatchVersion80 = 29
)

type MysqlVersion struct {
	Major   int  `json:"major,omitempty"`
	Minor   int  `json:"minor,omitempty"`
	NoPatch bool `json:"noPatch,omitempty"`
	Patch   int  `json:"patch,omitempty"`
}

func (v MysqlVersion) LT(major, minor, patch int) bool {
	return v.Major < major || v.Minor < minor || v.Patch < patch
}

func (v MysqlVersion) ValidateMasterHost(s string) bool {
	if v.Major == 5 {
		return Between(len(s), 1, 60)
	}
	if v.Major == 8 {
		return Between(len(s), 1, 255)
	}
	return false
}

func ParseVersion(s string) (v MysqlVersion, err error) {
	errVer := fmt.Errorf("version invalid: %s", s)

	a := strings.Split(strings.TrimPrefix(s, "v"), ".")
	if len(a) == 2 {
		v.NoPatch = true
	} else if len(a) == 3 {
		v.Patch, err = strconv.Atoi(a[2])
	} else {
		err = errVer
	}
	if err == nil {
		v.Major, err = strconv.Atoi(a[0])
		if err == nil {
			v.Minor, err = strconv.Atoi(a[1])
		}
	}
	if err != nil {
		err = errVer
		return
	}

	switch v.Major {
	case 5:
		if v.Minor != 7 {
			err = fmt.Errorf("minor version unsupported: %s", s)
		} else if v.NoPatch {
			v.Patch = MinPatchVersion57
		} else if v.Patch < MinPatchVersion57 {
			err = fmt.Errorf("patch version unsupported: %s", s)
		}
	case 8:
		if v.Minor != 0 {
			err = fmt.Errorf("minor version unsupported: %s", s)
		} else if v.NoPatch {
			v.Patch = MinPatchVersion80
		} else if v.Patch < MinPatchVersion80 {
			err = fmt.Errorf("patch version unsupported: %s", s)
		}
	default:
		err = fmt.Errorf("major version unsupported: %s", s)
	}

	return
}

var r = rand.New(rand.NewSource(time.Now().UnixNano() * int64(os.Getpid())))

func GeneratePassword(n int) string {
	a := [...]string{
		"0123456789",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"abcdefghijklmnopqrstuvwxyz",
	}
	b := make([]byte, n)
	for i := range b {
		j := r.Int()
		s := a[j%len(a)]
		b[i] = s[j%len(s)]
	}
	return string(b)
}

func Between(i, min, max int) bool {
	return min <= i && i <= max
}

func HasEqual(a ...int) bool {
	for i := 1; i < len(a); i++ {
		if a[i] == a[i-1] {
			return true
		}
	}
	return false
}

func HasQuote(a ...string) bool {
	for _, s := range a {
		if strings.ContainsAny(s, "'\"`") {
			return true
		}
	}
	return false
}

func SplitHostPort(addr string) (host string, port string) {
	i, j := -1, -1
	n := 0

	for k := 0; k < len(addr); k++ {
		switch addr[k] {
		case '[':
			if k > 0 {
				return
			}
		case ']':
			if i == -1 && addr[0] == '[' {
				i = k
			} else {
				return
			}
		case ':':
			j = k
			n++
		}
	}

	if i == -1 {
		if j == -1 {
			host = addr
		} else if n == 1 {
			host = addr[:j]
			port = addr[j+1:]
		}
	} else if j == -1 {
		if i+1 == len(addr) {
			host = addr[1:i]
		}
	} else if i+1 == j {
		host = addr[1:i]
		port = addr[j+1:]
	}

	return
}
