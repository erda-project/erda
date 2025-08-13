package mcp_server_request

import (
	"testing"
)

func TestParseMcpPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantName    string
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "valid path",
			path:        "/proxy/connect/mcp-fetch/1.0.0",
			wantName:    "mcp-fetch",
			wantVersion: "1.0.0",
			wantErr:     false,
		},
		{
			name:        "valid path with underscore",
			path:        "/proxy/connect/mcp_fetch/2.1.0",
			wantName:    "mcp_fetch",
			wantVersion: "2.1.0",
			wantErr:     false,
		},
		{
			name:        "valid path with hyphen",
			path:        "/proxy/connect/mcp-fetch-server/v1.0.0",
			wantName:    "mcp-fetch-server",
			wantVersion: "v1.0.0",
			wantErr:     false,
		},
		{
			name:        "invalid path - missing version",
			path:        "/proxy/connect/mcp-fetch",
			wantName:    "",
			wantVersion: "",
			wantErr:     true,
		},
		{
			name:        "invalid path - wrong prefix",
			path:        "/api/connect/mcp-fetch/1.0.0",
			wantName:    "",
			wantVersion: "",
			wantErr:     true,
		},
		{
			name:        "invalid path - extra segments",
			path:        "/proxy/connect/mcp-fetch/1.0.0/extra",
			wantName:    "",
			wantVersion: "",
			wantErr:     true,
		},
		{
			name:        "empty path",
			path:        "",
			wantName:    "",
			wantVersion: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotVersion, err := parseMcpPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMcpPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("parseMcpPath() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("parseMcpPath() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}

func TestParseSseMessagePath(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		wantName        string
		wantTag         string
		wantMessagePath string
		wantSessionId   string
		wantErr         bool
	}{
		{
			name:            "valid path",
			path:            "/proxy/message/mcp-fetch/1.0.0/message?sessionId=ace21671-7ddd-403a-901d-1b0c7b7b1f49",
			wantName:        "mcp-fetch",
			wantTag:         "1.0.0",
			wantMessagePath: "/message",
			wantSessionId:   "ace21671-7ddd-403a-901d-1b0c7b7b1f49",
			wantErr:         false,
		},
		{
			name:            "valid path with nested path",
			path:            "/proxy/message/mcp-fetch/2.1.0/api/v1/chat?sessionId=12345678-1234-1234-1234-123456789abc",
			wantName:        "mcp-fetch",
			wantTag:         "2.1.0",
			wantMessagePath: "/api/v1/chat",
			wantSessionId:   "12345678-1234-1234-1234-123456789abc",
			wantErr:         false,
		},
		{
			name:            "valid path with underscore in name",
			path:            "/proxy/message/mcp_fetch/v1.0.0/stream?sessionId=abcdef12-3456-7890-abcd-ef1234567890",
			wantName:        "mcp_fetch",
			wantTag:         "v1.0.0",
			wantMessagePath: "/stream",
			wantSessionId:   "abcdef12-3456-7890-abcd-ef1234567890",
			wantErr:         false,
		},
		{
			name:            "invalid path - missing sessionId",
			path:            "/proxy/message/mcp-fetch/1.0.0/message",
			wantName:        "",
			wantTag:         "",
			wantMessagePath: "",
			wantSessionId:   "",
			wantErr:         true,
		},
		{
			name:            "invalid path - wrong prefix",
			path:            "/api/message/mcp-fetch/1.0.0/message?sessionId=ace21671-7ddd-403a-901d-1b0c7b7b1f49",
			wantName:        "",
			wantTag:         "",
			wantMessagePath: "",
			wantSessionId:   "",
			wantErr:         true,
		},
		{
			name:            "invalid path - invalid sessionId format",
			path:            "/proxy/message/mcp-fetch/1.0.0/message?sessionId=invalid-uuid",
			wantName:        "",
			wantTag:         "",
			wantMessagePath: "",
			wantSessionId:   "",
			wantErr:         true,
		},
		{
			name:            "invalid path - missing name",
			path:            "/proxy/message//1.0.0/message?sessionId=ace21671-7ddd-403a-901d-1b0c7b7b1f49",
			wantName:        "",
			wantTag:         "",
			wantMessagePath: "",
			wantSessionId:   "",
			wantErr:         true,
		},
		{
			name:            "empty path",
			path:            "",
			wantName:        "",
			wantTag:         "",
			wantMessagePath: "",
			wantSessionId:   "",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotTag, gotMessagePath, gotSessionId, err := parseSseMessagePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSseMessagePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("parseSseMessagePath() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotTag != tt.wantTag {
				t.Errorf("parseSseMessagePath() gotTag = %v, want %v", gotTag, tt.wantTag)
			}
			if gotMessagePath != tt.wantMessagePath {
				t.Errorf("parseSseMessagePath() gotMessagePath = %v, want %v", gotMessagePath, tt.wantMessagePath)
			}
			if gotSessionId != tt.wantSessionId {
				t.Errorf("parseSseMessagePath() gotSessionId = %v, want %v", gotSessionId, tt.wantSessionId)
			}
		})
	}
}

func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     *ParsedEndpoint
		wantErr  bool
	}{
		{
			name:     "valid http endpoint with port",
			endpoint: "http://example.com:8080/api",
			want: &ParsedEndpoint{
				Endpoint: "http://example.com:8080/api",
				Host:     "example.com",
				Port:     "8080",
				Path:     "/api",
				Scheme:   "http",
			},
			wantErr: false,
		},
		{
			name:     "valid https endpoint without port",
			endpoint: "https://api.example.com/v1",
			want: &ParsedEndpoint{
				Endpoint: "https://api.example.com/v1",
				Host:     "api.example.com",
				Port:     "443",
				Path:     "/v1",
				Scheme:   "https",
			},
			wantErr: false,
		},
		{
			name:     "valid http endpoint without port and path",
			endpoint: "http://localhost",
			want: &ParsedEndpoint{
				Endpoint: "http://localhost",
				Host:     "localhost",
				Port:     "80",
				Path:     "/",
				Scheme:   "http",
			},
			wantErr: false,
		},
		{
			name:     "valid https endpoint with port and root path",
			endpoint: "https://example.com:8443/",
			want: &ParsedEndpoint{
				Endpoint: "https://example.com:8443/",
				Host:     "example.com",
				Port:     "8443",
				Path:     "/",
				Scheme:   "https",
			},
			wantErr: false,
		},
		{
			name:     "valid endpoint with complex path",
			endpoint: "http://api.example.com:3000/api/v1/chat/completions",
			want: &ParsedEndpoint{
				Endpoint: "http://api.example.com:3000/api/v1/chat/completions",
				Host:     "api.example.com",
				Port:     "3000",
				Path:     "/api/v1/chat/completions",
				Scheme:   "http",
			},
			wantErr: false,
		},
		{
			name:     "invalid endpoint - missing scheme",
			endpoint: "example.com:8080/api",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid endpoint - unsupported scheme",
			endpoint: "ftp://example.com:8080/api",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid endpoint - missing host",
			endpoint: "http://:8080/api",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid endpoint - malformed URL",
			endpoint: "http://example.com:invalid/api",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid endpoint - just scheme",
			endpoint: "http://",
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEndpoint(tt.endpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if got != nil {
					t.Errorf("parseEndpoint() got = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Errorf("parseEndpoint() got = nil, want %v", tt.want)
				return
			}
			if got.Endpoint != tt.want.Endpoint {
				t.Errorf("parseEndpoint() Endpoint = %v, want %v", got.Endpoint, tt.want.Endpoint)
			}
			if got.Host != tt.want.Host {
				t.Errorf("parseEndpoint() Host = %v, want %v", got.Host, tt.want.Host)
			}
			if got.Port != tt.want.Port {
				t.Errorf("parseEndpoint() Port = %v, want %v", got.Port, tt.want.Port)
			}
			if got.Path != tt.want.Path {
				t.Errorf("parseEndpoint() Path = %v, want %v", got.Path, tt.want.Path)
			}
			if got.Scheme != tt.want.Scheme {
				t.Errorf("parseEndpoint() Scheme = %v, want %v", got.Scheme, tt.want.Scheme)
			}
		})
	}
}

func TestParseMessage(t *testing.T) {
	name, tag, messagePath, id, err := parseSseMessagePath("/proxy/message/mcp-fetch/1.0.0/message?sessionId=ace21671-7ddd-403a-901d-1b0c7b7b1f49")
	if err != nil {
		t.Error(err)
	}
	t.Logf("name: %s tag: %s messagePath: %s, sessionId: %s", name, tag, messagePath, id)
}
