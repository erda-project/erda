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
)

func TestPropertyInstanceForShow_String(t *testing.T) {
	tests := []struct {
		name string
		p    PropertyInstanceForShow
		want string
	}{
		{
			name: "",
			p: PropertyInstanceForShow{
				ArbitraryValue: structpb.NewNumberValue(1.1),
			},
			want: "1.1",
		},
		{
			name: "",
			p: PropertyInstanceForShow{
				ArbitraryValue: structpb.NewStringValue("hello"),
			},
			want: "hello",
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
