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

package test_case

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/pkg/ai-functions/functions"
)

const Name = "create-test-case"

//go:embed schema.yaml
var schema json.RawMessage

//go:embed system-message.txt
var systemMessage string

var (
	_ functions.Function = (*Function)(nil)
)

func init() {
	functions.Register(Name, New)
}

type Function struct {
	background *functions.Background
}

func New(ctx context.Context, background *functions.Background) functions.Function {
	return &Function{background: background}
}

func (f *Function) Name() string {
	return Name
}

func (f *Function) Description() string {
	return "create test case"
}

func (f *Function) SystemMessage() string {
	return systemMessage
}

func (f *Function) UserMessage() string {
	return f.background.Prompt
}

func (f *Function) Schema() json.RawMessage {
	return schema
}

func (f *Function) Callback(ctx context.Context, arguments json.RawMessage) (any, error) {
	var req apistructs.TestCaseCreateRequest
	if err := json.Unmarshal(arguments, &req); err != nil {
		return nil, err
	}

	// todo: 背景信息如何获取
	req.Name = f.UserMessage()
	req.ProjectID = f.background.ProjectID
	req.Desc = fmt.Sprintf("Powered by AI.\n\n对应需求:\n%s", f.UserMessage())
	testSetID, _ := f.background.Resources["TestSet"].Int64()
	req.TestSetID = uint64(testSetID)
	req.Priority = apistructs.TestCasePriorityP3
	req.UserID = f.background.UserID

	// create on daily
	resp, err := bundle.New(bundle.WithErdaServer()).CreateTestCase(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
