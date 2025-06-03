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

package api_style

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ohler55/ojg/jp"

	"github.com/erda-project/erda/pkg/strutil"
)

// JSONPathParser parse JSONPath-like template strings.
// e.g.,
// - api-version=${@model.metadata.public.api_version||@provider.metadata.public.api_version||2025-03-01-preview}
//   - context(obj): model.metadata.public.api_version = ""
//     provider.metadata.public.api_version = "2024-01-01"
//     -> result: api-version=2024-01-01
//
// More examples, see: #TestJSONPathParser_SearchAndReplace
type JSONPathParser struct {
	RegexpPattern string
	re            *regexp.Regexp

	// splitter = `||` : a||b||c -> [ "a", "b", "c" ]
	MultiChoiceSplitter string
}

const DefaultRegexpPattern = `\$\{([^{}]+)\}`
const DefaultMultiChoiceSplitter = `||`

var defaultRegexpPatternRe = regexp.MustCompile(DefaultRegexpPattern)

func NewJSONPathParser(regexpPattern string, multiChoiceSplitter string) (*JSONPathParser, error) {
	var re *regexp.Regexp
	var err error
	if regexpPattern == DefaultRegexpPattern {
		re = defaultRegexpPatternRe
	} else {
		re, err = regexp.Compile(regexpPattern)
		if err != nil {
			return nil, fmt.Errorf("failed to parse regexp pattern %q: %w", regexpPattern, err)
		}
	}
	return &JSONPathParser{RegexpPattern: regexpPattern, re: re, MultiChoiceSplitter: multiChoiceSplitter}, nil
}

func MustNewJSONPathParser(regexpPattern string, multiChoiceSplitter string) *JSONPathParser {
	parser, err := NewJSONPathParser(regexpPattern, multiChoiceSplitter)
	if err != nil {
		panic(fmt.Sprintf("failed to create JSONPathParser: %v", err))
	}
	return parser
}

func (p *JSONPathParser) NeedDoReplace(s string) bool {
	return len(p.Search(s)) > 0
}

// Search return all matched keys. e.g.,
//   - api-version=${@model.metadata.public.api_version||@provider.metadata.public.api_version||2025-03-01-preview}
//     -> [ "@model.metadata.public.api_version||@provider.metadata.public.api_version||2025-03-01-preview ]
//   - text="${a||b||c} ${d}"
//     -> [ "a||b||c", "d" ]
func (p *JSONPathParser) Search(s string) []string {
	matches := p.re.FindAllStringSubmatch(s, -1)
	var keys []string
	for _, match := range matches {
		if len(match) > 1 {
			keys = append(keys, match[1])
		}
	}
	return keys
}

// SearchAndReplace .
// - api-version=${@model.metadata.public.api_version||@provider.metadata.public.api_version||2025-03-01-preview}
//   - context : model.metadata.public.api_version = ""
//     provider.metadata.public.api_version = "2024-01-01"
//     -> result: api-version=2024-01-01
func (p *JSONPathParser) SearchAndReplace(s string, availableValues map[string]any) string {
	return p.re.ReplaceAllStringFunc(s, func(match string) string {
		groups := p.re.FindStringSubmatch(match)
		if len(groups) < 2 {
			return "" // no match found, return empty
		}
		key := groups[1] // the first group is the key
		values := strings.Split(key, p.MultiChoiceSplitter)
		for _, v := range values {
			if !strings.HasPrefix(v, "@") {
				return v
			}
			value, err := p.getByJSONPath(v, availableValues)
			if err != nil {
				// log the error, but continue to try the next value
				fmt.Printf("error getting value by JSON path %q: %v\n", v, err)
				continue
			}
			if value != "" {
				return value // return the first non-empty value found
			}
		}
		return ""
	})
}

func (p *JSONPathParser) getByJSONPath(expr string, availableValues map[string]any) (string, error) {
	// remove prefix `@` to `.`
	jsonPath := "." + expr[1:]
	x, err := jp.ParseString(jsonPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse json path %q: %w", jsonPath, err)
	}
	out := x.Get(availableValues)
	if len(out) > 0 {
		return strutil.String(out[0]), nil
	}
	return "", nil
}
