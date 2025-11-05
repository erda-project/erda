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
	renderedBytes, err := findAndReplacePlaceholders(templateName, tpl.Placeholders, cfgBytes, params)
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
// Traverse every value in the config; if it is a slice or map, recurse until reaching leaf nodes.
// For each leaf node, check whether the value matches the `${@template.placeholders.xxx}` format.
// placeholderValues example:
// - api-key: 1234
// - another-api-key: 5678
// The placeholders argument enables fine-grained rendering (e.g., required flags, defaults, etc.).
func findAndReplacePlaceholders(templateName string, placeholderDefines []*pb.Placeholder, cfgJSON []byte, placeholderValues map[string]string) ([]byte, error) {
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

	updated, err := replacePlaceholdersRecursively(templateName, cfg, defByName, placeholderValues)
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
)

// ServiceProviderTypeParamKey is the reserved key that callers can use
// to inject the current service provider type into the rendering
// context. Mapping rules can leverage it to select provider-specific values.
const ServiceProviderTypeParamKey = "__service_provider_type__"

type omissionSentinel struct{}

var omitPlaceholder = &omissionSentinel{}

func replacePlaceholdersRecursively(templateName string, value interface{}, placeholderDefines map[string]*pb.Placeholder, placeholderValues map[string]string) (interface{}, error) {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, v := range typed {
			replaced, err := replacePlaceholdersRecursively(templateName, v, placeholderDefines, placeholderValues)
			if err != nil {
				return nil, err
			}
			if replaced == omitPlaceholder {
				delete(typed, key)
				continue
			}
			typed[key] = replaced
		}
		return typed, nil
	case []interface{}:
		result := make([]interface{}, 0, len(typed))
		for _, v := range typed {
			replaced, err := replacePlaceholdersRecursively(templateName, v, placeholderDefines, placeholderValues)
			if err != nil {
				return nil, err
			}
			if replaced == omitPlaceholder {
				continue
			}
			result = append(result, replaced)
		}
		return result, nil
	case string:
		if typed == templateNamePlaceholder {
			return templateName, nil
		}
		if !strings.HasPrefix(typed, placeholderPrefix) || !strings.HasSuffix(typed, placeholderSuffix) {
			return typed, nil
		}
		if len(typed) <= len(placeholderPrefix)+len(placeholderSuffix) {
			return nil, fmt.Errorf("invalid placeholder format: %q", typed)
		}
		name := typed[len(placeholderPrefix) : len(typed)-len(placeholderSuffix)]
		return resolvePlaceholderValue(templateName, name, placeholderDefines, placeholderValues)
	default:
		return value, nil
	}
}

func resolvePlaceholderValue(templateName, name string, placeholderDefines map[string]*pb.Placeholder, placeholderValues map[string]string) (interface{}, error) {
	definition, defined := placeholderDefines[name]
	if !defined {
		return nil, fmt.Errorf("undefined placeholder %q", name)
	}

	if val, ok := placeholderValues[name]; ok {
		converted, err := coercePlaceholderValue(val, definition)
		if err != nil {
			return nil, err
		}
		return evaluateStringIfNeeded(converted, templateName)
	}

	if mapped, ok, err := resolvePlaceholderMapping(definition, placeholderValues); err != nil {
		return nil, err
	} else if ok {
		converted, err := coercePlaceholderValue(mapped, definition)
		if err != nil {
			return nil, err
		}
		return evaluateStringIfNeeded(converted, templateName)
	}

	if hasPlaceholderDefault(definition) {
		defaultValue, err := coercePlaceholderDefault(definition)
		if err != nil {
			return nil, err
		}
		return evaluateStringIfNeeded(defaultValue, templateName)
	}

	if definition.GetRequired() {
		return nil, fmt.Errorf("missing required placeholder %q", name)
	}

	return omitPlaceholder, nil
}

func hasPlaceholderDefault(definition *pb.Placeholder) bool {
	if definition == nil {
		return false
	}
	return strings.TrimSpace(definition.GetDefault()) != ""
}

func coercePlaceholderDefault(definition *pb.Placeholder) (interface{}, error) {
	if definition == nil {
		return nil, nil
	}
	return coercePlaceholderValue(definition.GetDefault(), definition)
}

func coercePlaceholderValue(raw interface{}, definition *pb.Placeholder) (interface{}, error) {
	if definition == nil {
		return raw, nil
	}

	typ := strings.ToLower(strings.TrimSpace(definition.GetType()))
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

func evaluateStringIfNeeded(value interface{}, templateName string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}
	if str == templateNamePlaceholder {
		return templateName, nil
	}
	return str, nil
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

	if byRaw, ok := mapping["by"]; ok {
		by, ok := byRaw.(map[string]interface{})
		if !ok {
			return nil, false, fmt.Errorf("invalid mapping.by for placeholder %q", definition.GetName())
		}
		if value, found, err := resolveMappingBy(by, placeholderValues); err != nil {
			return nil, false, err
		} else if found {
			return value, true, nil
		}
	}

	if defaultRaw, ok := mapping["default"]; ok {
		return defaultRaw, true, nil
	}

	return nil, false, nil
}

func resolveMappingBy(by map[string]interface{}, placeholderValues map[string]string) (interface{}, bool, error) {
	if providerRaw, ok := by["service_provider_type"]; ok {
		providerMap, ok := providerRaw.(map[string]interface{})
		if !ok {
			return nil, false, fmt.Errorf("invalid mapping.by.service_provider_type format")
		}
		providerType := ""
		if placeholderValues != nil {
			providerType = placeholderValues[ServiceProviderTypeParamKey]
		}
		if providerType != "" {
			if value, found := providerMap[providerType]; found {
				return value, true, nil
			}
		}
		if value, found := providerMap["default"]; found {
			return value, true, nil
		}
		if value, found := providerMap["*"]; found {
			return value, true, nil
		}
	}
	return nil, false, nil
}
