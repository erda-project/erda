// Copyright (c) 2023 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
