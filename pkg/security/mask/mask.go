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

package mask

import (
	"bytes"
	"regexp"
	"strings"
	"sync"

	"github.com/erda-project/erda/pkg/strutil"
)

const (
	thresholdSize = 1024
)

var (
	defaultSensitiveEnvKeywords = []string{"pass", "key", "secret", "token", "pwd"}
	placeholder                 = "*"
	placeholderLength           = 5
)

type SensitiveEnv struct {
	passRegexp       *regexp.Regexp
	bufferPool       *sync.Pool
	keywords         []string
	keepFirstChar    bool
	stripANSIEscapes bool
}

type SensitiveEnvOption func(*SensitiveEnv)

func WithKeepFirstChar() SensitiveEnvOption {
	return func(m *SensitiveEnv) {
		m.keepFirstChar = true
	}
}

func WithStripANSIEscapes() SensitiveEnvOption {
	return func(m *SensitiveEnv) {
		m.stripANSIEscapes = true
	}
}

func WithKeywords(keywords []string) SensitiveEnvOption {
	return func(m *SensitiveEnv) {
		m.keywords = keywords
	}
}

func NewSensitiveEnv(options ...SensitiveEnvOption) *SensitiveEnv {
	m := &SensitiveEnv{
		bufferPool: &sync.Pool{
			New: func() any {
				b := make([]byte, 0, 1024)
				return &b
			},
		},
	}

	for _, option := range options {
		option(m)
	}

	if len(m.keywords) == 0 {
		m.keywords = defaultSensitiveEnvKeywords
	}

	pattern := `(?i)(?:` + strings.Join(m.keywords, "|") + `)[^\s\0]*=([^\s\0]+)`
	m.passRegexp = regexp.MustCompile(pattern)
	return m
}

func (m *SensitiveEnv) Process(input []byte) []byte {
	if m.stripANSIEscapes {
		input = strutil.StripANSIEscapes(input)
	}

	if len(input) <= thresholdSize {
		return m.passRegexp.ReplaceAllFunc(input, m.handler)
	}

	b := m.bufferPool.Get().(*[]byte)
	defer m.bufferPool.Put(b)
	*b = (*b)[:0]

	result := m.passRegexp.ReplaceAllFunc(input, m.handler)
	*b = append(*b, result...)
	return *b
}

func (m *SensitiveEnv) handler(a []byte) []byte {
	if i := bytes.IndexByte(a, '='); i != -1 {
		offset := 0
		if m.keepFirstChar {
			offset = 1
		}
		maskStartIndex := i + offset + 1
		if len(a) >= maskStartIndex {
			a = append(a[:maskStartIndex], strings.Repeat(placeholder, placeholderLength-offset)...)
		}
	}
	return a
}
