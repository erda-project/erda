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
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
)

func TestJSONPathParser_search(t *testing.T) {
	parser, err := NewJSONPathParser(DefaultRegexpPattern, DefaultMultiChoiceSplitter)
	assert.NoError(t, err)
	s := `api-version=${@model.metadata.public.api_version||@provider.metadata.public.api_version||2025-03-01-preview}`
	result := parser.Search(s)
	fmt.Println(result)

	s = `text="${a||b||c}-${d}"`
	result = parser.Search(s)
	assert.True(t, len(result) == 2)
	assert.Equal(t, "a||b||c", result[0])
	assert.Equal(t, "d", result[1])
}

func TestJSONPathParser_SearchAndReplace(t *testing.T) {
	type fields struct {
		RegexpPattern       string
		MultiChoiceSplitter string
	}
	type args struct {
		s               string
		availableValues map[string]any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "no context",
			fields: fields{
				RegexpPattern:       DefaultRegexpPattern,
				MultiChoiceSplitter: DefaultMultiChoiceSplitter,
			},
			args: args{
				s:               `api-version=${@model.metadata.public.api_version||@provider.metadata.public.api_version||2025-03-01-preview}`,
				availableValues: nil,
			},
			want: "api-version=2025-03-01-preview",
		},
		{
			name: "context: multi-choice, only first has value",
			fields: fields{
				RegexpPattern:       DefaultRegexpPattern,
				MultiChoiceSplitter: DefaultMultiChoiceSplitter,
			},
			args: args{
				s: `api-version=${@model.metadata.public.api_version||@provider.metadata.public.api_version||2025-03-01-preview}`,
				availableValues: map[string]any{
					"model": map[string]any{"metadata": map[string]any{"public": map[string]any{"api_version": "2025-05-28"}}},
				},
			},
			want: "api-version=2025-05-28",
		},
		{
			name: "context: multi-choice, only second has value",
			fields: fields{
				RegexpPattern:       DefaultRegexpPattern,
				MultiChoiceSplitter: DefaultMultiChoiceSplitter,
			},
			args: args{
				s: `api-version=${@model.metadata.public.api_version||@provider.metadata.public.api_version||2025-03-01-preview}`,
				availableValues: map[string]any{
					"provider": map[string]any{"metadata": map[string]any{"public": map[string]any{"api_version": "2025-05-27"}}},
				},
			},
			want: "api-version=2025-05-27",
		},
		{
			name: "context: multi-choice, all has value",
			fields: fields{
				RegexpPattern:       DefaultRegexpPattern,
				MultiChoiceSplitter: DefaultMultiChoiceSplitter,
			},
			args: args{
				s: `api-version=${@model.metadata.public.api_version||@provider.metadata.public.api_version||2025-03-01-preview}`,
				availableValues: map[string]any{
					"model":    map[string]any{"metadata": map[string]any{"public": map[string]any{"api_version": "2025-05-28"}}},
					"provider": map[string]any{"metadata": map[string]any{"public": map[string]any{"api_version": "2025-05-27"}}},
				},
			},
			want: "api-version=2025-05-28",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewJSONPathParser(tt.fields.RegexpPattern, tt.fields.MultiChoiceSplitter)
			assert.NoError(t, err)
			got, err := p.SearchAndReplace(tt.args.s, tt.args.availableValues)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, got, "SearchAndReplace(%v, %v)", tt.args.s, tt.args.availableValues)
		})
	}
}

func TestGetByJSONPath(t *testing.T) {
	parser, err := NewJSONPathParser(DefaultRegexpPattern, DefaultMultiChoiceSplitter)
	assert.NoError(t, err)

	// Test with a valid JSON path
	availableValues := map[string]any{
		"model": modelpb.Model{
			Metadata: &metadatapb.Metadata{
				Public: map[string]*structpb.Value{
					"api_version": structpb.NewStringValue("2025-05-28"),
				},
			},
		},
	}
	var m map[string]any
	cputil.MustObjJSONTransfer(&availableValues, &m)
	fmt.Printf("m: %+v\n", m)
	value, err := parser.getByJSONPath("@model.metadata.public.api_version", m)
	assert.NoError(t, err)
	assert.Equal(t, "2025-05-28", value)
}

func TestDotToBracketJSONPath(t *testing.T) {
	// dot -> bracket, keep prefix
	assert.Equal(t, `@["a"]["b"]["c"]`, DotToBracketJSONPath(`@.a.b.c`))
	assert.Equal(t, `$["a"]["b"]`, DotToBracketJSONPath(`$.a.b`))
	// already bracket
	assert.Equal(t, `@["a"]["b"]`, DotToBracketJSONPath(`@["a"]["b"]`))
	// mix bracket and dot
	assert.Equal(t, `@["a"]["b"]`, DotToBracketJSONPath(`@["a"].b`))
	// hyphen key
	assert.Equal(t, `@["provider"]["metadata"]["public"]["gcp-project-id"]`, DotToBracketJSONPath(`@provider.metadata.public.gcp-project-id`))
}

func TestGetByJSONPath_WithHyphenKey(t *testing.T) {
	parser, err := NewJSONPathParser(DefaultRegexpPattern, DefaultMultiChoiceSplitter)
	assert.NoError(t, err)

	// available values with hyphenated key
	availableValues := map[string]any{
		"provider": map[string]any{
			"metadata": map[string]any{
				"public": map[string]any{
					"gcp-project-id": "proj-123",
					"scheme":         "https",
					"host":           "example.com",
				},
				"secret": map[string]any{
					"gcp-sa-access-token": "token-xyz",
				},
			},
		},
		"template": map[string]any{
			"placeholders": map[string]any{
				"location": "us-central1",
			},
		},
	}

	// Convert to generic map to mimic runtime behavior
	var m map[string]any
	cputil.MustObjJSONTransfer(&availableValues, &m)

	// hyphenated key via dot notation (converted under the hood)
	v, err := parser.getByJSONPath("@provider.metadata.public.gcp-project-id", m)
	assert.NoError(t, err)
	assert.Equal(t, "proj-123", v)

	// same key via bracket notation
	v, err = parser.getByJSONPath(`@["provider"]["metadata"]["public"]["gcp-project-id"]`, m)
	assert.NoError(t, err)
	assert.Equal(t, "proj-123", v)

	// other fields without hyphens
	v, err = parser.getByJSONPath("@provider.metadata.public.scheme", m)
	assert.NoError(t, err)
	assert.Equal(t, "https", v)

	v, err = parser.getByJSONPath("@provider.metadata.public.host", m)
	assert.NoError(t, err)
	assert.Equal(t, "example.com", v)

	v, err = parser.getByJSONPath("@template.placeholders.location", m)
	assert.NoError(t, err)
	assert.Equal(t, "us-central1", v)

	v, err = parser.getByJSONPath("@provider.metadata.secret.gcp-sa-access-token", m)
	assert.NoError(t, err)
	assert.Equal(t, "token-xyz", v)
}
