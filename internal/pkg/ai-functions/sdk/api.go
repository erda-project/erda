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

package sdk

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"

	"github.com/erda-project/erda/pkg/strutil"
)

// VerifyArguments verifies that the given JSON conforms to the JSON Schema FunctionDefinition.Parameters
func VerifyArguments(parameters, data json.RawMessage) (err error) {
	// fd.Parameters and data may be either JSON or Yaml structured, convert to JSON structured uniformly.
	if parameters, err = strutil.YamlOrJsonToJson(parameters); err != nil {
		return errors.Wrap(err, "failed to unmarshal Parameters to JSON")
	}
	if data, err = strutil.YamlOrJsonToJson(data); err != nil {
		return errors.Wrap(err, "failed to unmarshal to JSON")
	}
	if valid := json.Valid(data); !valid {
		return errors.New("data is invalid JSON")
	}
	ls := gojsonschema.NewBytesLoader(parameters)
	ld := gojsonschema.NewBytesLoader(data)
	result, err := gojsonschema.Validate(ls, ld)
	if err != nil {
		return errors.Wrap(err, "failed to Validate")
	}
	if result.Valid() {
		return nil
	}
	var ss []string
	for _, item := range result.Errors() {
		ss = append(ss, item.String())
	}
	return errors.New(strings.Join(ss, "; "))
}
