package dataview

import (
	"testing"
)

func Test_provider_ExportTaskExecutor(t *testing.T) {
	type fields struct {
		ExportChannel chan string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"case1", fields{ExportChannel: make(chan string, 1)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				ExportChannel: tt.fields.ExportChannel,
			}
			p.ExportTaskExecutor()
		})
	}
}
