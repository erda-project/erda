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

package apistructs

import (
	"strconv"
)

type ProfileRenderRequest struct {
	Query             string `json:"query"`
	From              string `json:"from"`
	Until             string `json:"until"`
	MaxNodes          int    `json:"maxNodes"`
	FormatFlamebearer bool   `json:"formatFlamebearer"`
}

func (p *ProfileRenderRequest) URLQueryString() map[string][]string {
	query := make(map[string][]string)
	if p.Query != "" {
		query["query"] = append(query["query"], p.Query)
	}
	if p.From != "" {
		query["from"] = append(query["from"], p.From)
	}
	if p.Until != "" {
		query["until"] = append(query["until"], p.Until)
	}
	if p.FormatFlamebearer {
		query["formatFlamebearer"] = []string{"true"}
	}
	query["maxNodes"] = []string{"8192"}
	if p.MaxNodes != 0 {
		query["maxNodes"] = []string{strconv.Itoa(p.MaxNodes)}
	}

	return query
}
