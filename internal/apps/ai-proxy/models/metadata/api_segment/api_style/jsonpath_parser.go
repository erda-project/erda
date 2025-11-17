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
func (p *JSONPathParser) SearchAndReplace(s string, availableValues map[string]any) (string, error) {
	var firstErr error
	out := p.re.ReplaceAllStringFunc(s, func(match string) string {
		groups := p.re.FindStringSubmatch(match)
		if len(groups) < 2 {
			if firstErr == nil {
				firstErr = fmt.Errorf("invalid placeholder: %q", match)
			}
			return match
		}
		key := groups[1] // the first group is the key
		values := strings.Split(key, p.MultiChoiceSplitter)
		for _, v := range values {
			if !strings.HasPrefix(v, "@") {
				return v
			}
			value, err := p.getByJSONPath(v, availableValues)
			if err != nil {
				if firstErr == nil {
					firstErr = fmt.Errorf("failed to resolve %q: %w", v, err)
				}
				continue
			}
			if value != "" {
				return value // return the first non-empty value found
			}
		}
		if firstErr == nil {
			firstErr = fmt.Errorf("no value resolved for %q", key)
		}
		return match
	})
	if firstErr != nil {
		return out, firstErr
	}
	return out, nil
}

// SearchAndReplaceStrict 会在无法解析占位符时返回错误。
// 规则：
// - 支持多路选择（a||b||c），按顺序尝试；
// - 以 '@' 开头按 JSONPath 解析，否则视作字面量直接返回；
// - 任一路获得非空字符串即成功；若全部失败，则返回错误。
// Deprecated: SearchAndReplaceStrict is now the same as SearchAndReplace. Use SearchAndReplace instead.
func (p *JSONPathParser) SearchAndReplaceStrict(s string, availableValues map[string]any) (string, error) {
	return p.SearchAndReplace(s, availableValues)
}

func (p *JSONPathParser) getByJSONPath(expr string, availableValues map[string]any) (string, error) {
	// Normalize to a JSONPath starting with '$' and using bracket notation
	// Keep root instead of converting to a leading dot to avoid jp parser errors
	jsonPath := expr
	if strings.HasPrefix(jsonPath, "@") {
		jsonPath = "$" + jsonPath[1:]
	}
	jsonPath = DotToBracketJSONPath(jsonPath)
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

// DotToBracketJSONPath 把类似 @.a.b.c 这种 dot 形式，转换成 @["a"]["b"]["c"] 形式。
// 如果本来就包含 ["..."] 这种 bracket 形式，会原样保留。
func DotToBracketJSONPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return path
	}

	// 支持 @ 或 $ 开头
	prefix := ""
	rest := path
	if strings.HasPrefix(rest, "@") || strings.HasPrefix(rest, "$") {
		prefix = rest[:1]
		rest = rest[1:]
	}

	// 去掉紧随其后的可选点号：@.a.b.c / $.a.b.c / @a.b.c
	if strings.HasPrefix(rest, ".") {
		rest = rest[1:]
	}

	if rest == "" {
		if prefix != "" {
			return prefix
		}
		return path
	}

	// 将路径转换为完全的 bracket 形式，同时保留原有的 bracket 片段
	var b strings.Builder
	b.WriteString(prefix)

	i := 0
	for i < len(rest) {
		switch rest[i] {
		case '.':
			// 跳过点号
			i++
		case '[':
			// 直接保留现有的 bracket 片段，直到对应的 ']'
			end := i + 1
			depth := 1
			for end < len(rest) && depth > 0 {
				if rest[end] == '[' {
					depth++
				} else if rest[end] == ']' {
					depth--
				}
				end++
			}
			b.WriteString(rest[i:end])
			// 跳过紧随其后的点号
			i = end
			if i < len(rest) && rest[i] == '.' {
				i++
			}
		default:
			// 解析一个标识符直到下一个 '.' 或 '['
			start := i
			for i < len(rest) && rest[i] != '.' && rest[i] != '[' {
				i++
			}
			if start < i {
				b.WriteString(`["`)
				b.WriteString(rest[start:i])
				b.WriteString(`"]`)
			}
			// 点号在下一轮被跳过
		}
	}

	return b.String()
}
