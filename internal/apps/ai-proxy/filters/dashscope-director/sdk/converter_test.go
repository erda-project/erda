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

package sdk

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertDsResponseToOpenAIFormat_MultiPart(t *testing.T) {
	dsRespJson := `
{
    "output": {
        "choices": [
            {
                "finish_reason": "stop",
                "message": {
                    "role": "assistant",
                    "content": [ { "text": "hello" } ]
                }
            }
        ]
    },
    "usage": { "output_tokens": 75, "input_tokens": 572, "image_tokens": 391 },
    "request_id": "b90e5d5e-e60d-9f5c-880e-f4fe28627673"
}
`
	var dsResp DsResponse
	err := json.Unmarshal([]byte(dsRespJson), &dsResp)
	assert.NoError(t, err)
	oaiResp, err := ConvertDsResponseToOpenAIFormat(dsResp, "qwen-vl-max")
	assert.NoError(t, err)
	assert.Equal(t, "qwen-vl-max", oaiResp.Model)
	assert.Equal(t, "stop", string(oaiResp.Choices[0].FinishReason))
	assert.Equal(t, "hello", oaiResp.Choices[0].Message.Content)
}

func TestDsResponseOutputContent_Content(t *testing.T) {
	dsRespJson := `
{
    "output": {
        "choices": [
            {
                "finish_reason": "stop",
                "message": {
                    "role": "assistant",
                    "content": "hello"
                }
            }
        ]
    }
}
`
	var dsResp DsResponse
	err := json.Unmarshal([]byte(dsRespJson), &dsResp)
	assert.NoError(t, err)
	oaiResp, err := ConvertDsResponseToOpenAIFormat(dsResp, "test")
	assert.NoError(t, err)
	assert.Equal(t, "hello", oaiResp.Choices[0].Message.Content)
}
