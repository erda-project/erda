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
	"strings"

	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/pkg/excel"
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
	// read .xlsx file
	excelFile, err := os.Open("/Users/sfwn/Downloads/待评估项目需求.xlsx")
	if err != nil {
		log.Fatalf("failed to open excel file: %v", err)
	}
	sheets, err := excel.Decode(excelFile)
	if err != nil {
		log.Fatalf("failed to decode excel file: %v", err)
	}
	if len(sheets) != 3 {
		log.Fatalf("invalid excel file: %v", err)
	}
	requirementList := getRequirementListFromExcel(sheets[1])

	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("AZURE_OPENAI_BASE_URL")
	cfg := openai.DefaultAzureConfig(apiKey, baseURL)
	cfg.APIVersion = "2023-07-01-preview"
	client := openai.NewClientWithConfig(cfg)

	// handle requirement one by one
	var tableLines []string
	for _, requirement := range requirementList {
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
						Content: "这是一个'采购系统'，需要明确哪些是外部系统。",
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "当前已知的功能点，请勿重复计算 ILF 和 ELF。\n" + strings.Join(tableLines, "\n"),
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: fmt.Sprintf("业务模块: %s\n需求名称: %s\n需求描述: %s", requirement.BusinessModel, requirement.Name, requirement.Desc),
					},
				},
				Temperature:  0,
				Stream:       false,
				Functions:    []openai.FunctionDefinition{calculateFunctionPoint},
				FunctionCall: openai.FunctionCall{Name: calculateFunctionPoint.Name},
			},
		)
		if err != nil {
			log.Fatalf("failed to create chat completion: %v", err)
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
		var fpList struct {
			List []FunctionPoint `json:"list"`
		}
		if err := json.Unmarshal([]byte(fc.Arguments), &fpList); err != nil {
			log.Fatal("failed to unmarshal arguments", err)
		}
		// set defaults
		for i := range fpList.List {
			fpList.List[i].ReuseLevel = "低"
			fpList.List[i].AdjustLevel = "新增"
			fpList.List[i].RepeatCount = "否"
			if fpList.List[i].FirstLevelModule == "" {
				fpList.List[i].FirstLevelModule = requirement.BusinessModel
			}
			if fpList.List[i].SecondLevelModule == "" {
				fpList.List[i].SecondLevelModule = fpList.List[i].FirstLevelModule
			}
			tableLines = append(tableLines, fpList.List[i].ToTableLine())
		}
	}

	// write to file
	resultFile, err := os.Create("./internal/tools/ai/nesma/result.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer resultFile.Close()
	for _, line := range tableLines {
		if _, err := resultFile.WriteString(line + "\n"); err != nil {
			log.Fatal(err)
		}
	}
}

func init() {
	handleEnvFile()
}

type Requirement struct {
	BusinessModel string
	Name          string
	Desc          string
}

func getRequirementListFromExcel(sheet [][]string) (requirementList []Requirement) {
	var lastModuleName string
	for _, row := range sheet[2:] {
		if row[0] == "" {
			break
		}
		requirement := Requirement{
			BusinessModel: row[1],
			Name:          row[2],
			Desc:          row[3],
		}
		if requirement.BusinessModel == "" {
			requirement.BusinessModel = lastModuleName
		}
		lastModuleName = requirement.BusinessModel
		requirementList = append(requirementList, requirement)
	}
	return
}
