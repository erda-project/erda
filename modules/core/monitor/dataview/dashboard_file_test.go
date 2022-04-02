package dataview

import (
	"strings"
	"testing"
)

func Test_dashboardFileName(t *testing.T) {
	type args struct {
		scope   string
		scopeId string
		viewIds []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"case1", args{
				scope:   "org",
				scopeId: "1",
				viewIds: []string{"1"},
			}, "b3JnLTEtMj",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dashboardFilename(tt.args.scope, tt.args.scopeId); !strings.HasPrefix(got, tt.want) {
				t.Errorf("dashboardFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
