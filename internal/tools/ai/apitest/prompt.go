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
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed prompt.yaml
var PromptYaml embed.FS

type Prompt struct {
	SystemMessage string `yaml:"system,omitempty"`
	UserMessage   string `yaml:"user,omitempty"`
}

func generateContextPrompt() string {
	exprValues := getStepExpressionValues()
	//fmt.Println("expression values:")
	//printJSON(exprValues)

	// 构造 prompt
	contextPrompt := fmt.Sprintf(`
场景入参：%s
前置接口出参：%s
前置配置单出参：%s
全局变量入参：%s
mock：%s
`,
		jsonOutput(exprValues.SceneInputs),
		jsonOutput(exprValues.PreSceneStepsOutputs),
		jsonOutput(exprValues.PreConfigSheetsOutputs),
		jsonOutput(exprValues.GlobalConfigOutputs),
		jsonOutput(exprValues.MockInputs),
	)

	return contextPrompt
}

func jsonOutput(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func prettyJsonOutput(s string) string {
	m := make(map[string]interface{})
	json.Unmarshal([]byte(s), &m)
	b, _ := json.MarshalIndent(m, "", "  ")
	return string(b)
}
