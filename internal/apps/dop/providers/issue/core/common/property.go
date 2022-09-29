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

package common

import (
	"strconv"
	"strings"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
)

type PropertyInstanceForShow pb.IssuePropertyExtraProperty

func (p *PropertyInstanceForShow) String() string {
	if p.ArbitraryValue != nil {
		switch p.PropertyType {
		case pb.PropertyTypeEnum_Text:
			return p.ArbitraryValue.GetStringValue()
		case pb.PropertyTypeEnum_Number:
			return strconv.FormatFloat(p.ArbitraryValue.GetNumberValue(), 'f', 10, 64)
		default:
			return p.ArbitraryValue.GetStructValue().String()
		}
	}
	enumValueMap := make(map[int64]string, len(p.EnumeratedValues))
	for _, enum := range p.EnumeratedValues {
		enumValueMap[enum.Id] = enum.Name
	}
	var valueStrings []string
	for _, value := range p.Values {
		valueStrings = append(valueStrings, enumValueMap[value])
	}
	return strings.Join(valueStrings, ", ")
}
