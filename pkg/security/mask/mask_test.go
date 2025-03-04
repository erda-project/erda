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
	"math"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSensitiveEnv_Options(t *testing.T) {
	tests := []struct {
		name     string
		options  []SensitiveEnvOption
		expected *SensitiveEnv
	}{
		{
			name:    "default options",
			options: nil,
			expected: &SensitiveEnv{
				keywords: defaultSensitiveEnvKeywords,
			},
		},
		{
			name: "with keep first char",
			options: []SensitiveEnvOption{
				WithKeepFirstChar(),
			},
			expected: &SensitiveEnv{
				keywords:      defaultSensitiveEnvKeywords,
				keepFirstChar: true,
			},
		},
		{
			name: "with strip ANSI escapes",
			options: []SensitiveEnvOption{
				WithStripANSIEscapes(),
			},
			expected: &SensitiveEnv{
				keywords:         defaultSensitiveEnvKeywords,
				stripANSIEscapes: true,
			},
		},
		{
			name: "with custom keywords",
			options: []SensitiveEnvOption{
				WithKeywords([]string{"password", "api_key"}),
			},
			expected: &SensitiveEnv{
				keywords: []string{"password", "api_key"},
			},
		},
		{
			name: "with all options",
			options: []SensitiveEnvOption{
				WithKeepFirstChar(),
				WithStripANSIEscapes(),
				WithKeywords([]string{"password", "api_key"}),
			},
			expected: &SensitiveEnv{
				keywords:         []string{"password", "api_key"},
				keepFirstChar:    true,
				stripANSIEscapes: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSensitiveEnv(tt.options...)
			assert.Equal(t, tt.expected.keywords, m.keywords)
			assert.Equal(t, tt.expected.keepFirstChar, m.keepFirstChar)
			assert.Equal(t, tt.expected.stripANSIEscapes, m.stripANSIEscapes)
			assert.NotNil(t, m.passRegexp)
			assert.NotNil(t, m.bufferPool)
		})
	}
}

func TestSensitiveEnv_Process(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		options  []SensitiveEnvOption
		expected []byte
	}{
		{
			name:     "empty input",
			input:    []byte(""),
			options:  nil,
			expected: []byte(""),
		},
		{
			name:     "no sensitive data",
			input:    []byte("hello world"),
			options:  nil,
			expected: []byte("hello world"),
		},
		{
			name:     "password masking",
			input:    []byte("password=secret123"),
			options:  nil,
			expected: []byte("password=*****"),
		},
		{
			name:     "password masking, with empty value",
			input:    []byte("password="),
			options:  nil,
			expected: []byte("password="),
		},
		{
			name:     "password masking, with only 1 value",
			input:    []byte("password=1"),
			options:  nil,
			expected: []byte("password=*****"),
		},
		{
			name:     "token masking",
			input:    []byte("api_token=abc123xyz"),
			options:  nil,
			expected: []byte("api_token=*****"),
		},
		{
			name:     "keep first char",
			input:    []byte("password=secret123"),
			options:  []SensitiveEnvOption{WithKeepFirstChar()},
			expected: []byte("password=s****"),
		},
		{
			name:     "keep first char when value's length only 1",
			input:    []byte("password=s"),
			options:  []SensitiveEnvOption{WithKeepFirstChar()},
			expected: []byte("password=s****"),
		},
		{
			name:     "custom keywords",
			input:    []byte("custom_key=value123"),
			options:  []SensitiveEnvOption{WithKeywords([]string{"custom"})},
			expected: []byte("custom_key=*****"),
		},
		{
			name:     "multiple sensitive values",
			input:    []byte("password=secret123 token=abc123"),
			options:  nil,
			expected: []byte("password=***** token=*****"),
		},
		{
			name:     "case insensitive",
			input:    []byte("PASSWORD=secret123 Token=abc123"),
			options:  nil,
			expected: []byte("PASSWORD=***** Token=*****"),
		},
		{
			name:     "with ANSI escapes",
			input:    []byte("MYSQL_\u001B[31mPA\u001B[0mSSWORD=hello123"),
			options:  []SensitiveEnvOption{WithStripANSIEscapes()},
			expected: []byte("MYSQL_PASSWORD=*****"),
		},
		{
			name:     "start with ANSI escapes",
			input:    []byte("\u001B[31mPA\u001B[0mSSWORD=hello123"),
			options:  []SensitiveEnvOption{WithStripANSIEscapes()},
			expected: []byte("PASSWORD=*****"),
		},
		{
			name:     "large input",
			input:    []byte("password=" + strings.Repeat("x", math.MaxUint16)),
			options:  nil,
			expected: []byte("password=*****"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSensitiveEnv(tt.options...)
			result := m.Process(tt.input)
			assert.Equal(t, string(tt.expected), string(result))
		})
	}
}

func TestSensitiveEnv_Concurrent(t *testing.T) {
	m := NewSensitiveEnv()
	wg := sync.WaitGroup{}
	concurrency := 100

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := m.Process([]byte("password=secret123"))
			assert.Equal(t, "password=*****", string(result))
		}()
	}

	wg.Wait()
}
