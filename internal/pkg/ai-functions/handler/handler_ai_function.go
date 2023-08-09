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

package handler

import (
	"context"
	"encoding/json"
	"github.com/erda-project/erda/internal/pkg/ai-functions/functions"
	"github.com/erda-project/erda/internal/pkg/ai-functions/sdk"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/pkg/errors"
	structpb "google.golang.org/protobuf/types/known/structpb"
	"net/http"
	"net/url"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/apps/aifunction/pb"
)

var (
	_ pb.AiFunctionServer = (*AIFunction)(nil)
)

type AIFunction struct {
	Log       logs.Logger
	OpenaiURL *url.URL
}

func (h *AIFunction) Apply(ctx context.Context, req *pb.ApplyRequest) (*structpb.Value, error) {
	h.Log.Infof("apply the function with request %s", strutil.TryGetJsonStr(req))

	factory, ok := functions.Retrieve(req.GetFunctionName())
	if !ok {
		err := errors.Errorf("AI function %s not found", req.GetFunctionName())
		return nil, HTTPError(err, http.StatusBadRequest)
	}
	var (
		f         = factory(ctx, req.GetPrompt().GetPrompt(), req.GetBackground())
		systemMsg = &sdk.ChatMessage{
			Role:    "system",
			Content: f.SystemMessage(),
			Name:    "system",
		}
		userMsg = &sdk.ChatMessage{
			Role:    "user",
			Content: f.UserMessage(),
			Name:    "erda",
		}
		fd = &sdk.FunctionDefinition{
			Name:        f.Name(),
			Description: f.Description(),
			Parameters:  f.Schema(),
		}
		options = sdk.CreateCompletionOptions{
			Messages:     []*sdk.ChatMessage{systemMsg, userMsg},
			Functions:    []*sdk.FunctionDefinition{fd},
			FunctionCall: sdk.FunctionCall{Name: fd.Name},
			Temperature:  "1",
			Stream:       false,
			Model:        "gpt-35-turbo-16k",
		}
	)
	if valid := json.Valid(fd.Parameters); !valid {
		if err := yaml.Unmarshal(f.Schema(), &fd.Parameters); err != nil {
			return nil, err
		}
	}
	client, err := sdk.NewClient(h.OpenaiURL, http.DefaultClient)
	if err != nil {
		return nil, err
	}
	completion, err := client.CreateCompletion(ctx, &options)
	if err != nil {
		return nil, errors.Wrap(err, "failed to CreateCompletion")
	}
	if len(completion.Choices) == 0 {
		return nil, errors.New("no idea") // todo: do not return error, response friendly
	}
	arguments := json.RawMessage(completion.Choices[0].Message.FunctionCall.Arguments)
	if err = fd.VerifyArguments(arguments); err != nil {
		return nil, errors.Wrap(err, "invalid arguments from FunctionCall")
	}
	result, err := f.Callback(ctx, arguments)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to Callback with arguments: %s", string(arguments))
	}
	return structpb.NewValue(result)
}

// HTTPError todo: duplicate
func HTTPError(err error, code int) error {
	if err == nil {
		err = errors.New(http.StatusText(code))
	}
	return httpError{error: err, code: code}
}

type httpError struct {
	error
	code int
}

func (e httpError) HTTPStatus() int {
	return e.code
}
