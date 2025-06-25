package mcp_server

import (
	"testing"

	"gotest.tools/assert"
)

func TestBuildConstraint(t *testing.T) {
	tests := []struct {
		target string
		want   string
	}{
		{
			target: "1",
			want:   ">=1.0.0 <2.0.0",
		},
		{
			target: "1.2",
			want:   ">=1.2.0 <1.3.0",
		},
		{
			target: "1.2.3",
			want:   "=1.2.3",
		},
	}

	for _, tt := range tests {
		got, err := buildConstraint(tt.target)
		if err != nil {
			t.Errorf("buildConstraint(%s) error: %v", tt.target, err)
			continue
		}
		assert.Equal(t, tt.want, got.String())
	}
}
