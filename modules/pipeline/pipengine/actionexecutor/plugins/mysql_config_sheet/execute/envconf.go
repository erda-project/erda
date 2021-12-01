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

package execute

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/encoding/jsonparse"
)

// Conf js action param collection
type Conf struct {
	DataSource string        `env:"ACTION_DATASOURCE"`
	Host       string        `env:"ACTION_HOST"`
	Port       string        `env:"ACTION_PORT"`
	Username   string        `env:"ACTION_USERNAME"`
	Password   string        `env:"ACTION_PASSWORD"`
	Database   string        `env:"ACTION_DATABASE"`
	Executors  []SqlExecutor `env:"ACTION_EXECUTORS"`

	OrgID  string `env:"DICE_ORG_ID" required:"true"`
	UserID string `env:"DICE_USER_ID" required:"true"`
}

type SqlExecutor struct {
	Sql       string                   `json:"sql"`
	OutParams []apistructs.APIOutParam `json:"out_params"`
	Asserts   []APIAssert              `json:"asserts"`
}

type APIAssert struct {
	Arg      string      `json:"arg"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

func (a APIAssert) convert() apistructs.APIAssert {
	return apistructs.APIAssert{
		Arg:      a.Arg,
		Operator: a.Operator,
		Value:    jsonparse.JsonOneLine(a.Value),
	}
}
