package template

import (
	"encoding/json"
	"testing"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/template/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestFindAndReplacePlaceholders(t *testing.T) {
	t.Run("successfully replaces placeholders", func(t *testing.T) {
		cfg := map[string]interface{}{
			"service": map[string]interface{}{
				"apiKey": "${@template.placeholders.api-key}",
				"name":   "static-name",
			},
			"endpoints": []interface{}{
				"${@template.placeholders.endpoint}",
				"constant",
				map[string]interface{}{
					"token": "${@template.placeholders.token}",
				},
			},
		}

		cfgJSON, err := json.Marshal(cfg)
		require.NoError(t, err)

		defs := []*pb.Placeholder{
			{Name: "api-key", Type: "string", Required: true},
			{Name: "endpoint", Type: "string"},
			{Name: "token", Type: "string", Required: true},
		}

		rendered, err := findAndReplacePlaceholders("test-template", defs, cfgJSON, map[string]string{
			"api-key":  "1234",
			"endpoint": "https://example.com",
			"token":    "abcd",
		})
		require.NoError(t, err)

		var actual map[string]interface{}
		require.NoError(t, json.Unmarshal(rendered, &actual))

		expected := map[string]interface{}{
			"service": map[string]interface{}{
				"apiKey": "1234",
				"name":   "static-name",
			},
			"endpoints": []interface{}{
				"https://example.com",
				"constant",
				map[string]interface{}{
					"token": "abcd",
				},
			},
		}

		require.Equal(t, expected, actual)
	})

	t.Run("missing placeholder value returns error", func(t *testing.T) {
		cfg := map[string]interface{}{
			"apiKey": "${@template.placeholders.missing}",
		}
		cfgJSON, err := json.Marshal(cfg)
		require.NoError(t, err)

		defs := []*pb.Placeholder{
			{Name: "missing", Required: true},
		}

		_, err = findAndReplacePlaceholders("test-template", defs, cfgJSON, map[string]string{})
		require.Error(t, err)
		require.Contains(t, err.Error(), `missing required placeholder "missing"`)
	})

	t.Run("invalid placeholder format returns error", func(t *testing.T) {
		cfg := map[string]interface{}{
			"apiKey": "${@template.placeholders.}",
		}
		cfgJSON, err := json.Marshal(cfg)
		require.NoError(t, err)

		_, err = findAndReplacePlaceholders("test-template", nil, cfgJSON, map[string]string{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid placeholder format")
	})

	t.Run("uses default value when provided", func(t *testing.T) {
		cfg := map[string]interface{}{
			"metadata": map[string]interface{}{
				"host": "${@template.placeholders.host}",
			},
		}
		cfgJSON, err := json.Marshal(cfg)
		require.NoError(t, err)

		defs := []*pb.Placeholder{
			{
				Name:    "host",
				Default: "terminus.example.com",
			},
		}

		rendered, err := findAndReplacePlaceholders("test-template", defs, cfgJSON, map[string]string{})
		require.NoError(t, err)

		var actual map[string]interface{}
		require.NoError(t, json.Unmarshal(rendered, &actual))

		require.Equal(t, "terminus.example.com", actual["metadata"].(map[string]interface{})["host"])
	})

	t.Run("omits optional placeholder without default", func(t *testing.T) {
		cfg := map[string]interface{}{
			"secret": map[string]interface{}{
				"anotherApiKey": "${@template.placeholders.another-api-key}",
			},
		}
		cfgJSON, err := json.Marshal(cfg)
		require.NoError(t, err)

		defs := []*pb.Placeholder{
			{Name: "another-api-key", Required: false},
		}

		rendered, err := findAndReplacePlaceholders("test-template", defs, cfgJSON, map[string]string{})
		require.NoError(t, err)

		var actual map[string]interface{}
		require.NoError(t, json.Unmarshal(rendered, &actual))

		secret := actual["secret"].(map[string]interface{})
		_, exists := secret["anotherApiKey"]
		require.False(t, exists)
	})

	t.Run("object placeholder default string converts to map", func(t *testing.T) {
		cfg := map[string]interface{}{
			"metadata": map[string]interface{}{
				"public": map[string]interface{}{
					"context": "${@template.placeholders.context}",
				},
			},
		}
		cfgJSON, err := json.Marshal(cfg)
		require.NoError(t, err)

		defs := []*pb.Placeholder{
			{
				Name:    "context",
				Type:    "object",
				Default: `{"context_length": 131072, "max_completion_tokens": 16384, "max_prompt_tokens": 129024}`,
			},
		}

		rendered, err := findAndReplacePlaceholders("test-template", defs, cfgJSON, map[string]string{})
		require.NoError(t, err)

		var actual map[string]interface{}
		require.NoError(t, json.Unmarshal(rendered, &actual))

		context, ok := actual["metadata"].(map[string]interface{})["public"].(map[string]interface{})["context"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, float64(131072), context["context_length"])
		require.Equal(t, float64(16384), context["max_completion_tokens"])
		require.Equal(t, float64(129024), context["max_prompt_tokens"])
	})

	t.Run("object placeholder provided string converts to map", func(t *testing.T) {
		cfg := map[string]interface{}{
			"context": "${@template.placeholders.context}",
		}
		cfgJSON, err := json.Marshal(cfg)
		require.NoError(t, err)

		defs := []*pb.Placeholder{
			{
				Name: "context",
				Type: "object",
			},
		}

		value := `{"context_length": 10, "max_completion_tokens": 20, "max_prompt_tokens": 30}`
		rendered, err := findAndReplacePlaceholders("test-template", defs, cfgJSON, map[string]string{
			"context": value,
		})
		require.NoError(t, err)

		var actual map[string]interface{}
		require.NoError(t, json.Unmarshal(rendered, &actual))
		context := actual["context"].(map[string]interface{})
		require.Equal(t, float64(10), context["context_length"])
		require.Equal(t, float64(20), context["max_completion_tokens"])
		require.Equal(t, float64(30), context["max_prompt_tokens"])
	})

	t.Run("object placeholder invalid json returns error", func(t *testing.T) {
		cfg := map[string]interface{}{
			"context": "${@template.placeholders.context}",
		}
		cfgJSON, err := json.Marshal(cfg)
		require.NoError(t, err)

		defs := []*pb.Placeholder{
			{
				Name: "context",
				Type: "object",
			},
		}

		_, err = findAndReplacePlaceholders("test-template", defs, cfgJSON, map[string]string{
			"context": "not-json",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid object value for placeholder \"context\"")
	})

	t.Run("placeholder default referencing template name resolves", func(t *testing.T) {
		cfg := map[string]interface{}{
			"model_name": "${@template.placeholders.target_model_name}",
		}
		cfgJSON, err := json.Marshal(cfg)
		require.NoError(t, err)

		defs := []*pb.Placeholder{
			{
				Name:    "target_model_name",
				Type:    "string",
				Default: "${@template.name}",
			},
		}

		rendered, err := findAndReplacePlaceholders("gpt-4o", defs, cfgJSON, map[string]string{})
		require.NoError(t, err)

		var actual map[string]interface{}
		require.NoError(t, json.Unmarshal(rendered, &actual))
		require.Equal(t, "gpt-4o", actual["model_name"])
	})

	t.Run("resolves template name placeholder", func(t *testing.T) {
		cfg := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "${@template.name}",
			},
		}
		cfgJSON, err := json.Marshal(cfg)
		require.NoError(t, err)

		rendered, err := findAndReplacePlaceholders("template-foo", nil, cfgJSON, map[string]string{})
		require.NoError(t, err)

		var actual map[string]interface{}
		require.NoError(t, json.Unmarshal(rendered, &actual))
		require.Equal(t, "template-foo", actual["metadata"].(map[string]interface{})["name"])
	})

	t.Run("mapping applies service provider specific value", func(t *testing.T) {
		cfg := map[string]interface{}{
			"model_name": "${@template.placeholders.target_model_name}",
		}
		cfgJSON, err := json.Marshal(cfg)
		require.NoError(t, err)

		mapping := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"default": structpb.NewStringValue("${@template.name}"),
				"by": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"service_provider_type": structpb.NewStructValue(&structpb.Struct{
							Fields: map[string]*structpb.Value{
								"aws-bedrock": structpb.NewStringValue("us.anthropic.claude-opus-4-1-20250805-v1:0"),
							},
						}),
					},
				}),
			},
		}

		defs := []*pb.Placeholder{
			{
				Name:     "target_model_name",
				Type:     "string",
				Default:  "should-not-use",
				Mapping:  mapping,
				Required: false,
			},
		}

		// provider specific mapping
		rendered, err := findAndReplacePlaceholders("claude-opus-4.1", defs, cfgJSON, map[string]string{
			ServiceProviderTemplateNameParamKey: "aws-bedrock",
		})
		require.NoError(t, err)

		var actual map[string]interface{}
		require.NoError(t, json.Unmarshal(rendered, &actual))
		require.Equal(t, "us.anthropic.claude-opus-4-1-20250805-v1:0", actual["model_name"])

		// fallback to mapping default when provider does not match
		rendered, err = findAndReplacePlaceholders("claude-opus-4.1", defs, cfgJSON, map[string]string{
			ServiceProviderTemplateNameParamKey: "volcengine-ark",
		})
		require.NoError(t, err)
		actual = map[string]interface{}{}
		require.NoError(t, json.Unmarshal(rendered, &actual))
		require.Equal(t, "claude-opus-4.1", actual["model_name"])

		// fallback to mapping default when provider context missing
		rendered, err = findAndReplacePlaceholders("claude-opus-4.1", defs, cfgJSON, map[string]string{})
		require.NoError(t, err)
		actual = map[string]interface{}{}
		require.NoError(t, json.Unmarshal(rendered, &actual))
		require.Equal(t, "claude-opus-4.1", actual["model_name"])
	})

	t.Run("mapping default overrides placeholder default", func(t *testing.T) {
		cfg := map[string]interface{}{
			"model_name": "${@template.placeholders.target_model_name}",
		}
		cfgJSON, err := json.Marshal(cfg)
		require.NoError(t, err)

		mapping := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"default": structpb.NewStringValue("mapping-default"),
			},
		}

		defs := []*pb.Placeholder{
			{
				Name:    "target_model_name",
				Type:    "string",
				Default: "placeholder-default",
				Mapping: mapping,
			},
		}

		rendered, err := findAndReplacePlaceholders("sample-template", defs, cfgJSON, map[string]string{})
		require.NoError(t, err)

		var actual map[string]interface{}
		require.NoError(t, json.Unmarshal(rendered, &actual))
		require.Equal(t, "mapping-default", actual["model_name"])
	})
}
