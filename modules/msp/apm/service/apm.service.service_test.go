package service

import (
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	testing "testing"
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
