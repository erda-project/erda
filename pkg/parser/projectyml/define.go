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

package projectyml

import "github.com/erda-project/erda/apistructs"

type Spec struct {
	Version      string         `yaml:"version"`
	Applications []*Application `yaml:"applications"`
}

type Application struct {
	Name           string                    `yaml:"name"`
	DisplayName    string                    `yaml:"display_name"`
	Logo           string                    `yaml:"logo"`
	Desc           string                    `yaml:"desc"`
	Mode           string                    `yaml:"mode"`
	Config         map[string]interface{}    `yaml:"config"`
	IsExternalRepo bool                      `yaml:"isExternalRepo"`
	RepoConfig     *apistructs.GitRepoConfig `yaml:"repo_config"`
}
