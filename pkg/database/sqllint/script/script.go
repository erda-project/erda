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

package script

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
)

// Script represents a script file
type Script struct {
	// data is script file content
	data []byte
	// name is script file name
	name string
}

// New returns a *Script
func New(name string, data []byte) Script {
	return Script{
		data: data,
		name: name,
	}
}

// Name returns script file name
func (s *Script) Name() string {
	return s.name
}

// Data returns script file content
func (s *Script) Data() []byte {
	return s.data
}

// Checksum returns the script checksum with sha-256
func (s *Script) Checksum() string {
	data := bytes.TrimLeftFunc(s.Data(), func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t' || r == '\r'
	})
	hash := sha256.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}
