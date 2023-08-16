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
	"strconv"

	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/openapi/legacy/component-protocol/pkg/autotest/step"
	"github.com/erda-project/erda/pkg/discover"
)

var (
	bdl         *bundle.Bundle
	orgID       uint64
	userID      string
	spaceID     uint64
	sceneSetID  uint64
	sceneID     uint64
	sceneStepID uint64
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

	// get one api
	apiSpecDetail := mustGetOneAPI()

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt.SystemMessage,
		},
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("The swagger documentation content of the API selected by the user: %s", jsonOutput(apiSpecDetail)),
		},
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("In this test case generation, you can also use these context variables: %s", generateContextPrompt()),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("%s", prompt.UserMessage),
		},
	}
	//fmt.Println(jsonOutput(messages))

	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("AZURE_OPENAI_BASE_URL")
	cfg := openai.DefaultAzureConfig(apiKey, baseURL)
	cfg.APIVersion = "2023-07-01-preview"
	client := openai.NewClientWithConfig(cfg)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:        openai.GPT3Dot5Turbo16K,
			Messages:     messages,
			Temperature:  1,
			Stream:       false,
			Functions:    []openai.FunctionDefinition{updateAutotestAPIStep},
			FunctionCall: openai.FunctionCall{Name: updateAutotestAPIStep.Name},
		},
	)
	if err != nil {
		log.Fatalf("failed to create chat completion, err: %v", err)
	}
	if len(resp.Choices) == 0 {
		log.Fatal("no choices in response")
	}
	//fmt.Println(resp.Choices[0].Message.Content)
	fc := resp.Choices[0].Message.FunctionCall
	if fc == nil {
		log.Fatal("no function call in response")
	}
	fmt.Println(fc.Name)
	fmt.Println(fc.Arguments)
	var apiInfo apistructs.APIInfoV2
	if err := json.Unmarshal([]byte(fc.Arguments), &apiInfo); err != nil {
		log.Fatal("failed to unmarshal arguments", err)
	}
	apiInfo.Name = apiSpecDetail.Description
	apiInfo.Method = apiSpecDetail.Method
	apiInfo.URL = apiSpecDetail.Path
	if apiInfo.Body.Type == apistructs.APIBodyTypeApplicationJSON {
		apiInfo.Body.Content = prettyJsonOutput(apiInfo.Body.Content.(string))
	}

	// update api step
	apiStep := step.APISpec{
		APIInfo: apiInfo,
		Loop:    nil,
	}
	updateReq := apistructs.AutotestSceneRequest{
		AutoTestSceneParams: apistructs.AutoTestSceneParams{
			ID:      sceneStepID,
			SpaceID: spaceID,
		},
		SceneID: sceneID,
		Value:   jsonOutput(apiStep),
		IdentityInfo: apistructs.IdentityInfo{
			UserID: userID,
		},
		APISpecID: apiSpecDetail.ID,
		Name:      apiStep.APIInfo.Name,
	}
	stepID, err := bdl.UpdateAutoTestSceneStep(updateReq)
	if err != nil {
		log.Fatalf("failed to update api step, err: %v", err)
	}
	fmt.Printf("update api step success, step id: %d\n", stepID)
}

func init() {
	handleEnvFile()
	// parse org id
	_orgID, err := strconv.ParseUint(os.Getenv("ORG_ID"), 10, 64)
	if err != nil {
		log.Fatalf("failed to parse org id, err: %v", err)
	}
	orgID = _orgID
	// parse user id
	userID = os.Getenv("USER_ID")
	// parse space id
	_spaceID, err := strconv.ParseUint(os.Getenv("SPACE_ID"), 10, 64)
	if err != nil {
		log.Fatalf("failed to parse space id, err: %v", err)
	}
	spaceID = _spaceID
	// parse scene set id
	_sceneSetID, err := strconv.ParseUint(os.Getenv("SCENE_SET_ID"), 10, 64)
	if err != nil {
		log.Fatalf("failed to parse scene set id, err: %v", err)
	}
	sceneSetID = _sceneSetID
	// parse scene id
	_sceneID, err := strconv.ParseUint(os.Getenv("SCENE_ID"), 10, 64)
	if err != nil {
		log.Fatalf("failed to parse scene id, err: %v", err)
	}
	sceneID = _sceneID
	// parse scene step id
	_sceneStepID, err := strconv.ParseUint(os.Getenv("SCENE_STEP_ID"), 10, 64)
	if err != nil {
		log.Fatalf("failed to parse scene step id, err: %v", err)
	}
	sceneStepID = _sceneStepID
	// init bundle
	if err := os.Setenv(discover.EnvDOP, os.Getenv(discover.EnvErdaServer)); err != nil {
		log.Fatalf("failed to set env %s, err: %v", discover.EnvDOP, err)
	}
	bdl = bundle.New(bundle.WithAllAvailableClients())
}

func printJSON(o interface{}) {
	b, _ := json.MarshalIndent(o, "", "  ")
	fmt.Println(string(b))
}
