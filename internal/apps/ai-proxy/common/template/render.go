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

package template

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/template/pb"
)

func RenderTemplate(templateName string, tpl *pb.Template, params map[string]string) error {
	// convert to json
	cfgBytes, err := json.Marshal(tpl.Config)
	if err != nil {
		return err
	}
	// render json bytes
	renderedBytes, err := findAndReplacePlaceholders(templateName, tpl.Desc, tpl.Placeholders, cfgBytes, params)
	if err != nil {
		return err
	}
	// convert to tpl.Config
	if err := json.Unmarshal(renderedBytes, &tpl.Config); err != nil {
		return fmt.Errorf("failed to unmarshal rendered bytes: %w", err)
	}
	return nil
}

// findAndReplacePlaceholders
// traverse every value in the config; if it is a slice or map, recurse until reaching leaf nodes.
// for each leaf node, check whether the value matches the `${@template.placeholders.xxx}` format.
// placeholderValues example:
// - api-key: 1234
// - another-api-key: 5678
// the placeholders argument enables fine-grained rendering (e.g., required flags, defaults, etc.).
func findAndReplacePlaceholders(templateName, templateDesc string, placeholderDefines []*pb.Placeholder, cfgJSON []byte, placeholderValues map[string]string) ([]byte, error) {
	var cfg interface{}
	if err := json.Unmarshal(cfgJSON, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse template config: %w", err)
	}

	defByName := make(map[string]*pb.Placeholder, len(placeholderDefines))
	for _, def := range placeholderDefines {
		if def == nil {
			continue
		}
		if def.GetName() == "" {
			continue
		}
		defByName[def.GetName()] = def
	}

	updated, err := replacePlaceholdersRecursively(templateName, templateDesc, cfg, defByName, placeholderValues)
	if err != nil {
		return nil, err
	}

	rendered, err := json.Marshal(updated)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rendered config: %w", err)
	}
	return rendered, nil
}

const (
	placeholderPrefix       = "${@template.placeholders."
	placeholderSuffix       = "}"
	templateNamePlaceholder = "${@template.name}"
	TemplateDescPlaceholder = "${@template.desc}"
)

// ServiceProviderTypeParamKey is the reserved key that callers can use
// to inject the current service provider type into the rendering
// context. Mapping rules can leverage it to select provider-specific values.
const ServiceProviderTypeParamKey = "__service_provider_type__"

// PathMatcherParamKey is the reserved key that callers can use
// to inject the current path matcher pattern into the rendering
// context. Mapping rules can leverage it to select path-specific values.
const PathMatcherParamKey = "__path_matcher__"

func replacePlaceholdersRecursively(templateName, templateDesc string, value interface{}, placeholderDefines map[string]*pb.Placeholder, placeholderValues map[string]string) (interface{}, error) {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, v := range typed {
			replaced, err := replacePlaceholdersRecursively(templateName, templateDesc, v, placeholderDefines, placeholderValues)
			if err != nil {
				return nil, err
			}
			typed[key] = replaced
		}
		return typed, nil
	case []interface{}:
		result := make([]interface{}, 0, len(typed))
		for _, v := range typed {
			replaced, err := replacePlaceholdersRecursively(templateName, templateDesc, v, placeholderDefines, placeholderValues)
			if err != nil {
				return nil, err
			}
			result = append(result, replaced)
		}
		return result, nil
	case string:
		// exact match replacements preserve non-string types
		if typed == templateNamePlaceholder {
			return templateName, nil
		}
		if typed == TemplateDescPlaceholder {
			return templateDesc, nil
		}
		if strings.HasPrefix(typed, placeholderPrefix) && strings.HasSuffix(typed, placeholderSuffix) {
			if len(typed) <= len(placeholderPrefix)+len(placeholderSuffix) {
				return nil, fmt.Errorf("invalid placeholder format: %q", typed)
			}
			// exactly one placeholder and nothing else (starts with prefix and ends with suffix)
			name := typed[len(placeholderPrefix) : len(typed)-len(placeholderSuffix)]
			return resolvePlaceholderValue(templateName, templateDesc, name, placeholderDefines, placeholderValues)
		}
		// otherwise, support embedded placeholders inside strings
		replaced, err := replaceEmbeddedTemplatePlaceholders(typed, templateName, templateDesc, placeholderDefines, placeholderValues)
		if err != nil {
			return nil, err
		}
		return replaced, nil
	default:
		return value, nil
	}
}

var embeddedTokenRe = regexp.MustCompile(`\$\{([^{}]+)\}`)

// replaceEmbeddedTemplatePlaceholders scans a string for ${...} tokens and replaces
// only template-related placeholders (e.g., ${@template.placeholders.name}, ${@template.name}, ${@template.desc}).
// for embedded usage, values are coerced to strings. Unknown tokens are left as-is.
func replaceEmbeddedTemplatePlaceholders(input, templateName, templateDesc string, placeholderDefines map[string]*pb.Placeholder, placeholderValues map[string]string) (string, error) {
	var firstErr error
	out := embeddedTokenRe.ReplaceAllStringFunc(input, func(tok string) string {
		if firstErr != nil {
			return tok
		}
		inner := tok[2 : len(tok)-1]
		switch inner {
		case "@template.name":
			return templateName
		case "@template.desc":
			return templateDesc
		default:
			const pfx = "@template.placeholders."
			if strings.HasPrefix(inner, pfx) {
				name := inner[len(pfx):]
				if name == "" {
					firstErr = fmt.Errorf("invalid placeholder format: %q", tok)
					return tok
				}
				val, err := resolvePlaceholderValue(templateName, templateDesc, name, placeholderDefines, placeholderValues)
				if err != nil {
					firstErr = err
					return tok
				}
				// coerce to string for embedding
				switch v := val.(type) {
				case string:
					return v
				case []byte:
					return string(v)
				default:
					return fmt.Sprintf("%v", v)
				}
			}
			return tok
		}
	})
	if firstErr != nil {
		return "", firstErr
	}
	return out, nil
}

func resolvePlaceholderValue(templateName, templateDesc, name string, placeholderDefines map[string]*pb.Placeholder, placeholderValues map[string]string) (interface{}, error) {
	definition, defined := placeholderDefines[name]
	if !defined {
		return nil, fmt.Errorf("undefined placeholder %q", name)
	}

	if val, ok := placeholderValues[name]; ok {
		converted, err := coercePlaceholderValue(val, definition)
		if err != nil {
			return nil, err
		}
		return evaluateStringIfNeeded(converted, templateName, templateDesc, placeholderDefines, placeholderValues)
	}

	if mapped, ok, err := resolvePlaceholderMapping(definition, placeholderValues); err != nil {
		return nil, err
	} else if ok {
		converted, err := coercePlaceholderValue(mapped, definition)
		if err != nil {
			return nil, err
		}
		return evaluateStringIfNeeded(converted, templateName, templateDesc, placeholderDefines, placeholderValues)
	}

	if hasPlaceholderDefault(definition) {
		defaultValue, err := coercePlaceholderDefault(definition)
		if err != nil {
			return nil, err
		}
		return evaluateStringIfNeeded(defaultValue, templateName, templateDesc, placeholderDefines, placeholderValues)
	}

	// strict mode: any declared placeholder must render to a concrete value.
	// if neither provided, nor mapped, nor defaulted -> error.
	return nil, fmt.Errorf("missing value for placeholder %q", name)
}

func hasPlaceholderDefault(definition *pb.Placeholder) bool {
	if definition == nil {
		return false
	}
	return definition.Default != nil
}

func coercePlaceholderDefault(definition *pb.Placeholder) (interface{}, error) {
	if definition == nil {
		return nil, nil
	}
	if definition.Default == nil {
		return nil, nil
	}
	return coercePlaceholderValue(*definition.Default, definition)
}

func coercePlaceholderValue(raw interface{}, definition *pb.Placeholder) (interface{}, error) {
	if definition == nil {
		return raw, nil
	}

	typ := strings.ToLower(strings.TrimSpace(definition.GetType()))
	if s, ok := raw.(string); ok && strings.TrimSpace(s) == "" {
		switch typ {
		case "", "string":
			// allow empty strings for string targets
		default:
			return nil, fmt.Errorf("empty string is not a valid %q value for placeholder %q", typ, definition.GetName())
		}
	}

	switch typ {
	case "", "string":
		switch v := raw.(type) {
		case string:
			return v, nil
		case []byte:
			return string(v), nil
		default:
			return fmt.Sprintf("%v", v), nil
		}
	case "bool", "boolean":
		switch v := raw.(type) {
		case bool:
			return v, nil
		case string:
			parsed, err := strconv.ParseBool(strings.TrimSpace(v))
			if err != nil {
				return nil, fmt.Errorf("invalid boolean value %q for placeholder %q", v, definition.GetName())
			}
			return parsed, nil
		default:
			return nil, fmt.Errorf("invalid boolean value for placeholder %q", definition.GetName())
		}
	case "int", "integer":
		switch v := raw.(type) {
		case int:
			return v, nil
		case int64:
			return v, nil
		case float64:
			return int64(v), nil
		case string:
			parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid integer value %q for placeholder %q", v, definition.GetName())
			}
			return parsed, nil
		default:
			return nil, fmt.Errorf("invalid integer value for placeholder %q", definition.GetName())
		}
	case "float", "double", "number":
		switch v := raw.(type) {
		case float32:
			return float64(v), nil
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case string:
			parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid float value %q for placeholder %q", v, definition.GetName())
			}
			return parsed, nil
		default:
			return nil, fmt.Errorf("invalid float value for placeholder %q", definition.GetName())
		}
	case "json", "object", "map":
		switch v := raw.(type) {
		case map[string]interface{}:
			return v, nil
		case []interface{}:
			return v, nil
		case string:
			trimmed := strings.TrimSpace(v)
			if trimmed == "" {
				return map[string]interface{}{}, nil
			}
			var result interface{}
			if err := json.Unmarshal([]byte(trimmed), &result); err != nil {
				return nil, fmt.Errorf("invalid object value for placeholder %q: %w", definition.GetName(), err)
			}
			return result, nil
		default:
			return v, nil
		}
	default:
		return raw, nil
	}
}

func evaluateStringIfNeeded(value interface{}, templateName, templateDesc string, placeholderDefines map[string]*pb.Placeholder, placeholderValues map[string]string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}
	if str == templateNamePlaceholder {
		return templateName, nil
	}
	if str == TemplateDescPlaceholder {
		return templateDesc, nil
	}
	return replaceEmbeddedTemplatePlaceholders(str, templateName, templateDesc, placeholderDefines, placeholderValues)
}

func resolvePlaceholderMapping(definition *pb.Placeholder, placeholderValues map[string]string) (interface{}, bool, error) {
	mappingStruct := definition.GetMapping()
	if mappingStruct == nil {
		return nil, false, nil
	}

	mapping := mappingStruct.AsMap()
	if len(mapping) == 0 {
		return nil, false, nil
	}

	return resolveMappingRecursive(definition.GetName(), mapping, placeholderValues, make(map[string]bool))
}

func resolveMappingRecursive(placeholderName string, mapping map[string]interface{}, placeholderValues map[string]string, usedByTypes map[string]bool) (interface{}, bool, error) {
	var matchedValue interface{}
	found := false

	// define priority for 'by' types
	byTypesPriority := []string{"path_matcher", "service_provider_type"}

	if byRaw, ok := mapping["by"]; ok {
		by, ok := byRaw.(map[string]interface{})
		if !ok {
			return nil, false, fmt.Errorf("invalid mapping.by for placeholder %q", placeholderName)
		}

		for _, byType := range byTypesPriority {
			if condRaw, ok := by[byType]; ok {
				if usedByTypes[byType] {
					return nil, false, fmt.Errorf("duplicate mapping by type %q for placeholder %q in recursion", byType, placeholderName)
				}
				condMap, ok := condRaw.(map[string]interface{})
				if !ok {
					return nil, false, fmt.Errorf("invalid mapping.by.%s format for placeholder %q", byType, placeholderName)
				}

				currentVal := ""
				switch byType {
				case "path_matcher":
					currentVal = placeholderValues[PathMatcherParamKey]
				case "service_provider_type":
					currentVal = placeholderValues[ServiceProviderTypeParamKey]
				}

				if currentVal != "" {
					if val, ok := condMap[currentVal]; ok {
						matchedValue = val
						found = true
					}
				}

				if !found {
					if val, ok := condMap["default"]; ok {
						matchedValue = val
						found = true
					} else if val, ok := condMap["*"]; ok {
						matchedValue = val
						found = true
					}
				}

				if found {
					// mark this byType as used for the next recursion levels
					newUsedByTypes := make(map[string]bool, len(usedByTypes)+1)
					for k, v := range usedByTypes {
						newUsedByTypes[k] = v
					}
					newUsedByTypes[byType] = true

					// check if matchedValue is a nested mapping node
					if nextMapping, isMappingNode := isMappingNode(matchedValue); isMappingNode {
						return resolveMappingRecursive(placeholderName, nextMapping, placeholderValues, newUsedByTypes)
					}
					return matchedValue, true, nil
				}
			}
		}
	}

	// fallback to top-level default if no branch in 'by' matched or 'by' is missing
	if defaultRaw, ok := mapping["default"]; ok {
		if nextMapping, isMappingNode := isMappingNode(defaultRaw); isMappingNode {
			return resolveMappingRecursive(placeholderName, nextMapping, placeholderValues, usedByTypes)
		}
		return defaultRaw, true, nil
	}

	return nil, false, nil
}

func isMappingNode(val interface{}) (map[string]interface{}, bool) {
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, false
	}
	// a map is considered a mapping node if it contains "by" or "default" keys
	_, hasBy := m["by"]
	_, hasDefault := m["default"]
	if hasBy || hasDefault {
		return m, true
	}
	return nil, false
}
