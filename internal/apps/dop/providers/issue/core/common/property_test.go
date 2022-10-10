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
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
)

func TestPropertyInstanceForShow_String(t *testing.T) {
	tests := []struct {
		name string
		p    PropertyInstanceForShow
		want string
	}{
		{
			name: "is option: select",
			p: PropertyInstanceForShow{
				PropertyType: pb.PropertyTypeEnum_Select,
				EnumeratedValues: []*pb.Enumerate{
					{
						Id:    1024,
						Name:  "1024",
						Index: 0,
					},
				},
				Values: []int64{1024},
			},
			want: "1024",
		},
		{
			name: "is option: multi-select",
			p: PropertyInstanceForShow{
				PropertyType: pb.PropertyTypeEnum_MultiSelect,
				EnumeratedValues: []*pb.Enumerate{
					{
						Id:    1024,
						Name:  "1024",
						Index: 0,
					},
					{
						Id:    1025,
						Name:  "1025",
						Index: 1,
					},
				},
				Values: []int64{1024, 1025},
			},
			want: "1024, 1025",
		},
		{
			name: "is option: checkbox",
			p: PropertyInstanceForShow{
				PropertyType: pb.PropertyTypeEnum_CheckBox,
				EnumeratedValues: []*pb.Enumerate{
					{
						Id:    1024,
						Name:  "1024",
						Index: 0,
					},
					{
						Id:    1025,
						Name:  "1025",
						Index: 1,
					},
				},
				Values: []int64{1025},
			},
			want: "1025",
		},
		{
			name: "not option: no arbitrary value",
			p: PropertyInstanceForShow{
				PropertyType:   pb.PropertyTypeEnum_Text,
				ArbitraryValue: nil,
			},
			want: "",
		},
		{
			name: "not option: text",
			p: PropertyInstanceForShow{
				PropertyType:   pb.PropertyTypeEnum_Text,
				ArbitraryValue: structpb.NewStringValue("abc"),
			},
			want: "abc",
		},
		{
			name: "not option: number",
			p: PropertyInstanceForShow{
				PropertyType:   pb.PropertyTypeEnum_Number,
				ArbitraryValue: structpb.NewStringValue("123.45"),
			},
			want: "123.45",
		},
		{
			name: "not option: date",
			p: PropertyInstanceForShow{
				PropertyType:   pb.PropertyTypeEnum_Date,
				ArbitraryValue: structpb.NewStringValue("2022-10-09T00:00:00+08:00"),
			},
			want: "2022-10-09",
		},
		{
			name: "not option: invalid date",
			p: PropertyInstanceForShow{
				PropertyType:   pb.PropertyTypeEnum_Date,
				ArbitraryValue: structpb.NewStringValue("invalid date"),
			},
			want: "invalid date",
		},
		{
			name: "not option: person",
			p: PropertyInstanceForShow{
				PropertyType:   pb.PropertyTypeEnum_Person,
				ArbitraryValue: structpb.NewStringValue("10001"),
			},
			want: "10001",
		},
		{
			name: "not option: url",
			p: PropertyInstanceForShow{
				PropertyType:   pb.PropertyTypeEnum_URL,
				ArbitraryValue: structpb.NewStringValue("https://erda.cloud"),
			},
			want: "https://erda.cloud",
		},
		{
			name: "not option: email",
			p: PropertyInstanceForShow{
				PropertyType:   pb.PropertyTypeEnum_Email,
				ArbitraryValue: structpb.NewStringValue("a@erda.cloud"),
			},
			want: "a@erda.cloud",
		},
		{
			name: "not option: phone",
			p: PropertyInstanceForShow{
				PropertyType:   pb.PropertyTypeEnum_Phone,
				ArbitraryValue: structpb.NewStringValue("18012345678"),
			},
			want: "18012345678",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPropertyInstanceForShow_TryGetUserID(t *testing.T) {
	tests := []struct {
		name string
		p    PropertyInstanceForShow
		want string
	}{
		{
			name: "not person type",
			p: PropertyInstanceForShow{
				PropertyType: pb.PropertyTypeEnum_Email,
			},
			want: "",
		},
		{
			name: "person type",
			p: PropertyInstanceForShow{
				PropertyType:   pb.PropertyTypeEnum_Person,
				ArbitraryValue: structpb.NewStringValue("bob"),
			},
			want: "bob",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.TryGetUserID(); got != tt.want {
				t.Errorf("TryGetUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}
