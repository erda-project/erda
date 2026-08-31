package application

import "testing"

func TestGetRepoConfigStrNil(t *testing.T) {
	if got := getRepoConfigStr(nil); got != "" {
		t.Fatalf("getRepoConfigStr(nil) = %q, want empty string", got)
	}
}
