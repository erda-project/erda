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
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

//go:embed prompt.yaml
var PromptYaml embed.FS

var createTestCase = openai.FunctionDefinition{
	Name:        "create-test-case",
	Description: "create test case",
	Parameters: &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"name": {
				Type:        jsonschema.String,
				Description: "the name of test case",
			},
			"preset": {
				Type:        jsonschema.String,
				Description: "前置条件",
				Items: &jsonschema.Definition{
					Type: jsonschema.String,
				},
			},
			"steps": {
				Type:        jsonschema.Array,
				Description: "步骤及结果",
				Items: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"step_name": {
							Type:        jsonschema.String,
							Description: "步骤名称",
						},
						"step_result": {
							Type:        jsonschema.String,
							Description: "期望结果",
						},
					},
				},
			},
		},
		Required: []string{"name", "preset", "steps"},
	},
}

type Prompt struct {
	SystemMessage string `yaml:"system,omitempty"`
	UserMessage   string `yaml:"user,omitempty"`
}

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
	//cfg.APIVersion = "2023-07-01-preview"
	client := openai.NewClientWithConfig(cfg)
	stream, err := client.CreateChatCompletionStream(
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
			Temperature: 0.8,
			Stream:      true,
			Functions:   []openai.FunctionDefinition{createTestCase},
			//FunctionCall: openai.FunctionCall{Name: createTestCase.Name},
			FunctionCall: "auto",
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()
	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("EOF")
				break
			}
			log.Fatal(err)
		}
		if len(resp.Choices) == 0 {
			b, _ := json.MarshalIndent(resp, "", "  ")
			log.Fatal("no choices in response: ", string(b))
		}
		fmt.Print(resp.Choices[0].Delta.Content)
	}
}
