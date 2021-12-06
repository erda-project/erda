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

package service

import (
	testing "testing"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
)

func Test_parseLanguage(t *testing.T) {
	type args struct {
		platform string
	}
	tests := []struct {
		name         string
		args         args
		wantLanguage commonpb.Language
	}{
		{"case1", args{platform: "unknown 1.2.3"}, commonpb.Language_unknown},
		{"case2", args{platform: "JDK 1.2.3"}, commonpb.Language_java},
		{"case3", args{platform: "NODEJS 1.2.3"}, commonpb.Language_nodejs},
		{"case4", args{platform: "PYTHON 1.2.3"}, commonpb.Language_python},
		{"case5", args{platform: "c 1.2.3"}, commonpb.Language_c},
		{"case6", args{platform: "c++ 1.2.3"}, commonpb.Language_cpp},
		{"case7", args{platform: "c# 1.2.3"}, commonpb.Language_csharp},
		{"case8", args{platform: "go 1.2.3"}, commonpb.Language_golang},
		{"case9", args{platform: "php 1.2.3"}, commonpb.Language_php},
		{"case10", args{platform: ".net 1.2.3"}, commonpb.Language_dotnet},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotLanguage := parseLanguage(tt.args.platform); gotLanguage != tt.wantLanguage {
				t.Errorf("parseLanguage() = %v, want %v", gotLanguage, tt.wantLanguage)
			}
		})
	}
}
