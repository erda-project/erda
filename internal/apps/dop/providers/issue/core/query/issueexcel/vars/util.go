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

package vars

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/pkg/strutil"
)

func ParseStringSliceByComma(s string) []string {
	results := strutil.Splits(s, []string{",", "，"}, true)
	// trim space
	for i, v := range results {
		results[i] = strings.TrimSpace(v)
	}
	return results
}

func ChangePointerTimeToTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func MustGetJsonManHour(estimateTime string) string {
	manHour, err := NewManhour(estimateTime)
	if err != nil {
		panic(fmt.Errorf("failed to get man hour from estimate time, err: %v", err))
	}
	b, _ := json.Marshal(&manHour)
	return string(b)
}

var estimateRegexp, _ = regexp.Compile(`(\d+)([wdhm]?)`)

func NewManhour(manhour string) (pb.IssueManHour, error) {
	if manhour == "" {
		return pb.IssueManHour{}, nil
	}
	if !estimateRegexp.MatchString(manhour) {
		return pb.IssueManHour{}, fmt.Errorf("invalid estimate time: %s", manhour)
	}
	matches := estimateRegexp.FindAllStringSubmatch(manhour, -1)
	var totalMinutes int64
	for _, match := range matches {
		timeVal, err := strconv.ParseUint(match[1], 10, 64)
		if err != nil {
			return pb.IssueManHour{}, fmt.Errorf("invalid man hour: %s, err: %v", manhour, err)
		}
		timeType := match[2]
		switch timeType {
		case "m":
			totalMinutes += int64(timeVal)
		case "h":
			totalMinutes += int64(timeVal) * 60
		case "d":
			totalMinutes += int64(timeVal) * 60 * 8
		case "w":
			totalMinutes += int64(timeVal) * 60 * 8 * 5
		default:
			return pb.IssueManHour{}, fmt.Errorf("invalid man hour: %s", manhour)
		}
	}
	return pb.IssueManHour{EstimateTime: totalMinutes, RemainingTime: totalMinutes}, nil
}

func GetIssueStage(model IssueSheetModel) string {
	if model.Common.IssueType == pb.IssueTypeEnum_TASK {
		return model.TaskOnly.TaskType
	}
	if model.Common.IssueType == pb.IssueTypeEnum_BUG {
		return model.BugOnly.Source
	}
	return ""
}
