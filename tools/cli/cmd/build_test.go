package cmd

import "testing"

func TestNormalizePipelineYmlName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain root pipeline", in: "pipeline.yml", want: "pipeline.yml"},
		{name: "dot slash root pipeline", in: "./pipeline.yml", want: "pipeline.yml"},
		{name: "nested pipeline", in: ".erda/pipelines/java-demo.yml", want: ".erda/pipelines/java-demo.yml"},
		{name: "dot slash nested pipeline", in: "./.erda/pipelines/java-demo.yml", want: ".erda/pipelines/java-demo.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizePipelineYmlName(tt.in); got != tt.want {
				t.Fatalf("normalizePipelineYmlName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
