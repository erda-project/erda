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

package provider

import (
	"context"
	"encoding/json"
	"github.com/erda-project/erda/internal/pkg/ai-functions/functions"
	"github.com/erda-project/erda/internal/pkg/ai-functions/sdk"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
)

type provider struct {
	Config Config
}

func (p *provider) ApplyFunction(ctx context.Context, name string, background *functions.Background) (any, error) {
	factory, ok := functions.Retrieve(name)
	if !ok {
		return nil, errors.Errorf("AI function %s not found", name)
	}
	var (
		f         = factory(ctx, background)
		systemMsg = &sdk.ChatMessage{
			Role:    "system",
			Content: f.SystemMessage(),
		}
		userMsg = &sdk.ChatMessage{
			Role:    "user",
			Content: f.UserMessage(),
		}
		fd = &sdk.FunctionDefinition{
			Name:        f.Name(),
			Description: f.Description(),
			Parameters:  f.Schema(),
		}
		req = sdk.CreateCompletionOptions[string]{
			Messages:     []*sdk.ChatMessage{systemMsg, userMsg},
			Functions:    []*sdk.FunctionDefinition{fd},
			FunctionCall: "auto",
		}
	)
	client := sdk.NewClient(p.Config.openAIOpenapiURL, http.DefaultClient)
	completion, err := sdk.CreateCompletion(ctx, client, &req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to CreateCompletion")
	}
	if len(completion.Choices) == 0 {
		return nil, errors.New("no idea")
	}
	arguments := json.RawMessage(completion.Choices[0].Message.FunctionCall.Arguments)
	if err = fd.VerifyArguments(arguments); err != nil {
		return nil, errors.Wrap(err, "invalid arguments from FunctionCall")
	}
	result, err := f.Callback(ctx, arguments)
	if err != nil {
		return nil, errors.Wrapf(err, "field to Callback with arguments: %s", string(arguments))
	}
	return result, nil
}

type Config struct {
	// OpenAIOpenapi is the API address which implemented the OpenAI API.
	// It like https://api.openai.com, https://ai-proxy.erda.cloud
	OpenAIOpenapi string `json:"openAIOpenapi" yaml:"openAIOpenapi"`

	openAIOpenapiURL *url.URL
}
