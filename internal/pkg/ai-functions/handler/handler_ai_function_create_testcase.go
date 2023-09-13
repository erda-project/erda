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
	"net/url"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/apps/aifunction/pb"
	"github.com/erda-project/erda/internal/pkg/ai-functions/functions"
	aitestcase "github.com/erda-project/erda/internal/pkg/ai-functions/functions/test-case"
	aiHandlerUtils "github.com/erda-project/erda/internal/pkg/ai-functions/handler/utils"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

func (h *AIFunction) createTestCaseForRequirementIDAndTestID(ctx context.Context, factory functions.FunctionFactory, req *pb.ApplyRequest, openaiURL *url.URL) (any, error) {
	results := make([]any, 0)
	var wg sync.WaitGroup
	var functionParams aitestcase.FunctionParams

	FunctionParamsBytes, err := req.GetFunctionParams().MarshalJSON()
	if err != nil {
		return nil, errors.Wrapf(err, "MarshalJSON for req.FunctionParams failed.")
	}
	if err = json.Unmarshal(FunctionParamsBytes, &functionParams); err != nil {
		return nil, errors.Wrapf(err, "Unmarshal req.FunctionParams to struct FunctionParams failed.")
	}
	logrus.Debugf("parse createTestCase functionParams=%+v", functionParams)

	if err := validateParamsForCreateTestcase(functionParams); err != nil {
		return nil, errors.Wrapf(err, "validateParamsForCreateTestcase faild")
	}

	for _, tp := range functionParams.Requirements {
		wg.Add(1)
		var err error
		if err = processSingleTestCase(ctx, factory, req, openaiURL, &wg, tp, functionParams.TestSetID, &results); err != nil {
			return nil, errors.Wrapf(err, "process single testCase create faild")
		}
	}
	wg.Wait()

	content := httpserver.Resp{
		Success: true,
		Data:    results,
	}
	return json.Marshal(content)
}

func processSingleTestCase(ctx context.Context, factory functions.FunctionFactory, req *pb.ApplyRequest, openaiURL *url.URL, wg *sync.WaitGroup, tp aitestcase.TestCaseParam, testSetId uint64, results *[]any) error {
	defer wg.Done()

	callbackInput := aitestcase.TestCaseFunctionInput{
		TestSetID: testSetId,
		IssueID:   tp.IssueID,
		Prompt:    tp.Prompt,
	}

	result, err := aiHandlerUtils.GetChatMessageFunctionCallArguments(ctx, factory, req, openaiURL, tp.Prompt, callbackInput)
	if err != nil {
		return err
	}

	*results = append(*results, result)
	return nil
}

// validateParamsForCreateTestcase 校验创建测试用例对应的参数配置
func validateParamsForCreateTestcase(req aitestcase.FunctionParams) error {
	if req.TestSetID <= 0 {
		return errors.Errorf("AI function functionParams testSetID for %s invalid", aitestcase.Name)
	}

	if len(req.Requirements) == 0 {
		return errors.Errorf("AI function functionParams requirements for %s invalid", aitestcase.Name)
	}

	for idx, tp := range req.Requirements {
		if tp.IssueID <= 0 {
			return errors.Errorf("AI function functionParams requirements[%d].issueID for %s invalid", idx, aitestcase.Name)
		}
		if tp.Prompt == "" {
			return errors.Errorf("AI function functionParams requirements[%d].prompt for %s invalid", idx, aitestcase.Name)
		}
	}

	return nil
}
