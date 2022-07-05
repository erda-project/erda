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

package log_service

import (
	"testing"

	"github.com/erda-project/erda-proto-go/msp/apm/log-service/pb"
)

func TestParseRegexp(t *testing.T) {
	type parseRes struct {
		groups []*pb.RegexpGroup
	}
	isEqual := func(a, b parseRes) bool {
		if len(a.groups) != len(b.groups) {
			return false
		}
		for i := range a.groups {
			if a.groups[i].GetPattern() != b.groups[i].GetPattern() ||
				a.groups[i].GetName() != b.groups[i].GetName() {
				return false
			}
		}
		return true
	}

	tests := []struct {
		content string
		res     parseRes
	}{
		{
			content: "(?P<group1>.*)",
			res: parseRes{groups: []*pb.RegexpGroup{
				{
					Pattern: "(?P<group1>.*)",
					Name:    "group1",
				},
			}},
		},
		{
			content: "(\\d*)",
			res: parseRes{groups: []*pb.RegexpGroup{
				{
					Pattern: "(\\d*)",
					Name:    "",
				},
			}},
		},
		{
			content: "^(?!xxx)\\w+",
			res:     parseRes{},
		},
		{
			content: "^(?!xxx)(?P<group1>.*)(\\d*)(?P<group2>.*)",
			res: parseRes{groups: []*pb.RegexpGroup{
				{
					Pattern: "(?P<group1>.*)",
					Name:    "group1",
				},
				{
					Pattern: "(\\d*)",
					Name:    "",
				},
				{
					Pattern: "(?P<group2>.*)",
					Name:    "group2",
				},
			}},
		},
	}

	for i, test := range tests {
		groups, err := parseRegexp(test.content)
		if err != nil {
			t.Fatal(err)
		}
		if !isEqual(parseRes{groups}, test.res) {
			t.Errorf("case %d is not expected", i)
		}
	}
}
