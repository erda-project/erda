package vars

import "testing"

func TestTrimBearer(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "BearerPrefix",
			input: "Bearer token",
			want:  "token",
		},
		{
			name:  "bearerPrefix",
			input: "bearer token",
			want:  "token",
		},
		{
			name:  "BearerOnly",
			input: "Bearer ",
			want:  "",
		},
		{
			name:  "bearerOnly",
			input: "bearer ",
			want:  "",
		},
		{
			name:  "NoPrefix",
			input: "token",
			want:  "token",
		},
		{
			name:  "Empty",
			input: "",
			want:  "",
		},
		{
			name:  "LeadingSpace",
			input: " Bearer token",
			want:  " Bearer token",
		},
		{
			name:  "NotExactPrefix",
			input: "BearerBearer token",
			want:  "BearerBearer token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrimBearer(tt.input)
			if got != tt.want {
				t.Fatalf("TrimBearer(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
