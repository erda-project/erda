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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func main() {
	f, err := PromptYaml.Open("prompt.yaml")
	if err != nil {
		log.Fatal(err)
	}
	var prompt Prompt
	if err := yaml.NewDecoder(f).Decode(&prompt); err != nil {
		log.Fatal(err)
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("AZURE_OPENAI_BASE_URL")
	cfg := openai.DefaultAzureConfig(apiKey, baseURL)
	cfg.APIVersion = "2023-07-01-preview"
	client := openai.NewClientWithConfig(cfg)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo16K,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompt.SystemMessage,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt.UserMessage,
				},
			},
			Temperature:  1,
			Stream:       false,
			Functions:    []openai.FunctionDefinition{createTestCase},
			FunctionCall: openai.FunctionCall{Name: createTestCase.Name},
		},
	)
	if err != nil {
		log.Fatal("xxx", err)
	}
	if len(resp.Choices) == 0 {
		log.Fatal("no choices in response")
	}
	fc := resp.Choices[0].Message.FunctionCall
	if fc == nil {
		log.Fatal("no function call in response")
	}
	fmt.Println(fc.Name)
	fmt.Println(fc.Arguments)
	var testCaseCreateInfo apistructs.TestCaseCreateRequest
	if err := json.Unmarshal([]byte(fc.Arguments), &testCaseCreateInfo); err != nil {
		log.Fatal("failed to unmarshal arguments", err)
	}

	// fulfill other infos
	testCaseCreateInfo.Name = prompt.UserMessage
	testCaseCreateInfo.ProjectID = 1904
	testCaseCreateInfo.Desc = fmt.Sprintf("Powered by AI.\n\n对应需求:\n%s", prompt.UserMessage)
	testCaseCreateInfo.TestSetID = 24186
	testCaseCreateInfo.Priority = apistructs.TestCasePriorityP3
	testCaseCreateInfo.UserID = "1005834"

	// create in daily
	bdl := bundle.New(bundle.WithErdaServer())
	createResp, err := bdl.CreateTestCase(testCaseCreateInfo)
	if err != nil {
		log.Fatal("failed to create test case", err)
	}
	fmt.Println(createResp)
}

func init() {
	handleEnvFile()
}
