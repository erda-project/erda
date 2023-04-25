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

package audit

import "encoding/json"

func ExtractCompletionFromCreateCompletionResp(s string) (string, error) {
	var m = make(map[string]json.RawMessage)
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return "", err
	}
	message, ok := m["choices"]
	if !ok {
		return "", nil
	}
	var choices []*CreateCompletionChoice
	if err := json.Unmarshal(message, &choices); err != nil {
		return "", err
	}
	if len(choices) == 0 {
		return "", nil
	}
	return choices[0].Text, nil
}
