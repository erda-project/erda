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

package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/apps/aifunction/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/ai-functions/functions"
	"github.com/erda-project/erda/internal/pkg/ai-functions/sdk"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	FunctionNameTestCase = "create-test-case"
)

// getChatMessageFunctionCallArguments return result for AIFunction Server Call OpenAI
func GetChatMessageFunctionCallArguments(ctx context.Context, factory functions.FunctionFactory, req *pb.ApplyRequest, openaiURL *url.URL, messages []openai.ChatCompletionMessage, callbackInput interface{}) (any, error) {
	var (
		err error
		f   = factory(ctx, "", req.GetBackground())
	)

	schema, err := f.Schema()
	if err != nil {
		return nil, err
	}

	fd := openai.FunctionDefinition{
		Name:        f.Name(),
		Description: f.Description(),
		Parameters:  schema,
	}
	logrus.Debugf("openai.FunctionDefinition fd.Parameters string: %s\n", fd.Parameters)

	options := &openai.ChatCompletionRequest{
		Messages:     messages, // todo: history messages
		Functions:    []openai.FunctionDefinition{fd},
		FunctionCall: openai.FunctionCall{Name: fd.Name},
		Temperature:  1, // default 1, can be modified by f.CompletionOptions()
		Stream:       false,
		Model:        "gpt-35-turbo-16k", // default the newest model, can be modified by f.CompletionOptions()
	}

	cos := f.CompletionOptions()
	for _, o := range cos {
		o(options)
	}

	userInfo, err := sdk.GetUserInfo(ctx)
	if err != nil {
		return nil, err
	}
	userName := userInfo.Nick
	if userName == "" {
		userName = userInfo.Name
	}

	// 在 request option 中添加认证信息: 以某组织下某用户身份调用 ai-proxy,
	// ai-proxy 中的 filter erda-auth 会回调 erda.cloud 的 openai, 检查该企业和用户是否有权使用 AI 能力
	source := ""
	switch req.GetFunctionName() {
	case FunctionNameTestCase:
		source = "testcase___"
	default:
		source = "ai_function___"
	}
	ros := append(f.RequestOptions(), func(r *http.Request) {
		r.Header.Set("X-Ai-Proxy-Source", base64.StdEncoding.EncodeToString([]byte(source+os.Getenv(string(apistructs.DICE_CLUSTER_NAME)))))
		r.Header.Set("X-Ai-Proxy-Org-Id", base64.StdEncoding.EncodeToString([]byte(apis.GetOrgID(ctx))))
		r.Header.Set("X-Ai-Proxy-User-Id", base64.StdEncoding.EncodeToString([]byte(apis.GetUserID(ctx))))
		r.Header.Set("X-Ai-Proxy-Email", base64.StdEncoding.EncodeToString([]byte(userInfo.Email)))
		r.Header.Set("X-Ai-Proxy-Name", base64.StdEncoding.EncodeToString([]byte(userName)))
		r.Header.Set("X-Ai-Proxy-Phone", base64.StdEncoding.EncodeToString([]byte(userInfo.Phone)))
	})
	client, err := sdk.NewClient(openaiURL, http.DefaultClient, ros...)
	if err != nil {
		return nil, err
	}
	completion, err := client.CreateCompletion(ctx, options)
	if err != nil {
		return nil, errors.Wrap(err, "failed to CreateCompletion")
	}
	if len(completion.Choices) == 0 || completion.Choices[0].Message.FunctionCall == nil {
		return nil, errors.New("no idea") // todo: do not return error, response friendly
	}
	// todo: check index out of range and invalid memory reference
	arguments, err := strutil.YamlOrJsonToJson([]byte(completion.Choices[0].Message.FunctionCall.Arguments))
	if err != nil {
		arguments = json.RawMessage(completion.Choices[0].Message.FunctionCall.Arguments)
	}

	if err = sdk.VerifyArguments(fd.Parameters.(json.RawMessage), arguments); err != nil {
		return nil, errors.Wrap(err, "invalid arguments from FunctionCall")
	}

	result, err := f.Callback(ctx, arguments, callbackInput)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to Callback with arguments: %s", string(arguments))
	}

	return result, nil
}
