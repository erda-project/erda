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

package jsonnet

import (
	"github.com/google/go-jsonnet"
)

type Engine struct {
	JsonnetVM *jsonnet.VM
}

type TLACodeConfig struct {
	Key   string
	Value string
}

func (t *Engine) EvaluateBySnippet(snippet string, configs []TLACodeConfig) (string, error) {
	t.SetTLACodes(configs)
	jsonStr, err := t.JsonnetVM.EvaluateAnonymousSnippet("template.jsonnet", snippet)
	if err != nil {
		return "", err
	}
	return jsonStr, err
}

func (t *Engine) EvaluateByFile(fileName string) (string, error) {
	jsonStr, err := t.JsonnetVM.EvaluateFile(fileName)
	if err != nil {
		return "", err
	}
	return jsonStr, err
}

func (t *Engine) SetTLACodes(configs []TLACodeConfig) {
	for _, config := range configs {
		t.JsonnetVM.TLACode(config.Key, config.Value)
	}
}
