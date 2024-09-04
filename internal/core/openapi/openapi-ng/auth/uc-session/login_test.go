package ucoauth

import (
	"testing"
)

func TestGetUCRedirectHost(t *testing.T) {
	tests := []struct {
		name       string
		referer    string
		host       string
		config     config
		wantResult string
	}{
		{
			name:    "Matching referer with UCRedirectAddrs",
			referer: "https://erda.cloud/login",
			host:    "erda.cloud",
			config: config{
				UCRedirectAddrs: []string{"openapi.erda.cloud"},
			},
			wantResult: "openapi.erda.cloud",
		},
		{
			name:    "Non-matching referer with UCRedirectAddrs",
			referer: "https://erda.cloud/login",
			host:    "erda.cloud",
			config: config{
				UCRedirectAddrs: []string{"fake.erda.cloud"},
			},
			wantResult: "fake.erda.cloud",
		},
		{
			name:    "Host with port number",
			referer: "https://erda.cloud:8080/login",
			host:    "erda.cloud:8080",
			config: config{
				UCRedirectAddrs: []string{"openapi.erda.cloud:8080"},
			},
			wantResult: "openapi.erda.cloud:8080",
		},
		{
			name:    "Empty host",
			referer: "https://erda.cloud:8080/login",
			host:    "",
			config: config{
				UCRedirectAddrs: []string{"openapi.erda.cloud:8080"},
			},
			wantResult: "openapi.erda.cloud:8080",
		},
		{
			name:    "Referer and host have different domains",
			referer: "https://erda.cloud/login",
			host:    "another.com",
			config: config{
				UCRedirectAddrs: []string{"openapi.erda.cloud"},
			},
			wantResult: "openapi.erda.cloud",
		},
		{
			name:    "Empty host and Referer with diff port",
			referer: "https://erda.cloud:8080/login",
			host:    "",
			config: config{
				UCRedirectAddrs: []string{"openapi.erda.cloud:9090"},
			},
			wantResult: "openapi.erda.cloud:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				Cfg: &tt.config,
			}
			got := p.getUCRedirectHost(tt.referer, tt.host)
			if got != tt.wantResult {
				t.Errorf("getUCRedirectHost() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}
