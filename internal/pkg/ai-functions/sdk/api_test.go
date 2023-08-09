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

package sdk_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/erda-project/erda-proto-go/apps/aifunction/pb"
	"github.com/erda-project/erda/internal/pkg/ai-functions/functions"
	_ "github.com/erda-project/erda/internal/pkg/ai-functions/functions/test-case"
	"github.com/erda-project/erda/internal/pkg/ai-functions/sdk"
)

const completion = `{
  "id": "chatcmpl-7lXUuoOKvQRwugUDLJoLmdEnJcrFT",
  "object": "chat.completion",
  "created": 1691564536,
  "model": "gpt-35-turbo-16k",
  "prompt_annotations": [
    {
      "prompt_index": 0,
      "content_filter_results": {
        "hate": {
          "filtered": false,
          "severity": "safe"
        },
        "self_harm": {
          "filtered": false,
          "severity": "safe"
        },
        "sexual": {
          "filtered": false,
          "severity": "safe"
        },
        "violence": {
          "filtered": false,
          "severity": "safe"
        }
      }
    }
  ],
  "choices": [
    {
      "index": 0,
      "finish_reason": "stop",
      "message": {
        "role": "assistant",
        "function_call": {
          "name": "create-test-case",
          "arguments": "{\n  \"name\": \"渠道商品列表含税单价字段保留小数点后两位\",\n  \"preCondition\": \"\",\n  \"stepAndResult\": [\n    {\n      \"step\": \"进入【渠道商品】列表页面\",\n      \"result\": \"成功打开【渠道商品】列表页面\"\n    },\n    {\n      \"step\": \"查看【渠道商品】列表页面的含税单价字段\",\n      \"result\": \"字段显示的小数点后两位为保留后的值\"\n    },\n    {\n      \"step\": \"创建一个含税单价为1.234的商品\",\n      \"result\": \"成功创建商品，并且列表页面的【含税单价】字段显示为1.23\"\n    },\n    {\n      \"step\": \"创建一个含税单价为0.005的商品\",\n      \"result\": \"成功创建商品，并且列表页面的【含税单价】字段显示为0.01\"\n    },\n    {\n      \"step\": \"创建一个含税单价为1.236的商品\",\n      \"result\": \"成功创建商品，并且列表页面的【含税单价】字段显示为1.24\"\n    },\n    {\n      \"step\": \"创建一个含税单价为0的商品\",\n      \"result\": \"成功创建商品，并且列表页面的【含税单价】字段显示为0.00\"\n    }\n  ]\n}"
        }
      },
      "content_filter_results": {}
    }
  ],
  "usage": {
    "completion_tokens": 321,
    "prompt_tokens": 333,
    "total_tokens": 654
  }
}`

func TestChatCompletion_JSONUnmarshal(t *testing.T) {
	var cc sdk.ChatCompletions
	if err := json.Unmarshal([]byte(completion), &cc); err != nil {
		t.Fatalf("failed to json.Unmarshal, err: %v", err)
	}

	if len(cc.Choices) == 0 {
		t.Fatal("failed to json.Unmarshal, cc.Choices is empty")
	}

	factory, ok := functions.Retrieve("create-test-case")
	if !ok {
		t.Fatal("failed to functions.Retrieve")
	}
	f := factory(context.Background(), "hello", &pb.Background{})
	fd := sdk.FunctionDefinition{
		Name:        f.Name(),
		Description: f.Description(),
		Parameters:  f.Schema(),
	}
	arguments := json.RawMessage(cc.Choices[0].Message.FunctionCall.Arguments)
	if err := fd.VerifyArguments(arguments); err != nil {
		t.Fatal(err)
	}
}
