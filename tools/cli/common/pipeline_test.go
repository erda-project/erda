package common

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func TestLatestPipelineID(t *testing.T) {
	tests := []struct {
		name      string
		pipelines []apistructs.PagePipeline
		wantID    uint64
		wantOK    bool
	}{
		{
			name:      "empty pipelines",
			pipelines: nil,
			wantID:    0,
			wantOK:    false,
		},
		{
			name: "pick max id",
			pipelines: []apistructs.PagePipeline{
				{ID: 1002},
				{ID: 998},
				{ID: 1200},
			},
			wantID: 1200,
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := latestPipelineID(tt.pipelines)
			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Fatalf("latestPipelineID() = (%d, %v), want (%d, %v)", gotID, gotOK, tt.wantID, tt.wantOK)
			}
		})
	}
}
