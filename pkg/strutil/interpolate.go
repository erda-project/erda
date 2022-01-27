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

package strutil

import (
	"strings"

	"github.com/pkg/errors"
)

func FirstPlaceholder(s string) (string, int, int, error) {
	for i := 0; i <= len(s)-2; i++ {
		if s[i] == '$' && s[i+1] == '{' {
			for j := i; j < len(s); j++ {
				if s[j] == '\n' || s[j] == '\r' {
					return "", 0, 0, errors.New("invalid literal: '\\n' or '\\r' in string")
				}
				if s[j] == '}' {
					return s[i+2 : j], i, j + 1, nil
				}
			}
		}
	}
	return "", 0, 0, nil
}

// Interpolate 对 s 中的 ${PLACEHOLDER} 或 ${PLACEHOLDER:DEFAULT} 占位符进行插值
func Interpolate(s string, values map[string]string, defaultPrecedence bool) (string, error) {
	if values == nil {
		values = make(map[string]string)
	}
	var valuesCopy = make(map[string]string)
	for k, v := range values {
		valuesCopy[k] = v
	}

	if err := InterpolationDereference(valuesCopy); err != nil {
		return s, err
	}

	for {
		placeholder, indexStart, indexEnd, err := FirstPlaceholder(s)
		if err != nil {
			return s, err
		}
		if indexStart == indexEnd {
			return s, nil
		}
		kv := strings.Split(placeholder, ":")
		placeholder = strings.TrimSpace(kv[0])
		value, ok := valuesCopy[placeholder]
		if len(kv) > 1 && (!ok || defaultPrecedence) {
			value = strings.TrimSpace(kv[1])
		}
		s = s[:indexStart] + value + s[indexEnd:]
	}
}

func InterpolationDereference(values map[string]string) error {
	for k, v := range values {
		placeholder, indexStart, indexEnd, err := FirstPlaceholder(k)
		if err != nil {
			return err
		}
		if indexStart != indexEnd {
			return errors.Errorf("placeholder %s in the values' key %s", placeholder, k)
		}
		placeholder, indexStart, indexEnd, err = FirstPlaceholder(v)
		if placeholder == k {
			return errors.Errorf("loop reference in key %s", k)
		}
		values[k] = v[:indexStart] + values[placeholder] + v[indexEnd:]
	}
	for k := range values {
		_, indexStart, indexEnd, err := FirstPlaceholder(values[k])
		if err != nil {
			return err
		}
		if indexStart != indexEnd {
			return InterpolationDereference(values)
		}
	}
	return nil
}
