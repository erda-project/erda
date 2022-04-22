// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package edgepipeline_register

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func Test_parseDialerEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
		wantErr  bool
	}{
		{
			name:     "invalid endpoint",
			endpoint: "xxx",
			want:     "xxx",
			wantErr:  false,
		},
		{
			name:     "http endpoint",
			endpoint: "http://cluster-dialer:80",
			want:     "ws://cluster-dialer:80",
			wantErr:  false,
		},
		{
			name:     "https endpoint",
			endpoint: "https://cluster-dialer:80",
			want:     "wss://cluster-dialer:80",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		p := &provider{
			Cfg: &Config{
				IsEdge:              true,
				ClusterDialEndpoint: tt.endpoint,
			},
		}
		got, err := p.parseDialerEndpoint()
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. provider.parseDialerEndpoint() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. provider.parseDialerEndpoint() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestGetAccessToken(t *testing.T) {
	p := &provider{
		Cfg: &Config{
			IsEdge:      true,
			accessToken: "xxx",
		},
	}
	accessToken, err := p.GetAccessToken(apistructs.OAuth2TokenGetRequest{})
	assert.NoError(t, err)
	assert.Equal(t, p.Cfg.accessToken, accessToken.AccessToken)
}

func TestGetOAuth2Token(t *testing.T) {
	p := &provider{
		Cfg: &Config{
			IsEdge:      true,
			accessToken: "xxx",
		},
	}
	oauth2Token, err := p.GetOAuth2Token(apistructs.OAuth2TokenGetRequest{})
	assert.NoError(t, err)
	assert.Equal(t, p.Cfg.accessToken, oauth2Token.AccessToken)
}

func TestCheckAccessToken(t *testing.T) {
	p := &provider{
		Cfg: &Config{
			IsEdge:      true,
			accessToken: "xxx",
		},
	}
	tests := []struct {
		name        string
		accessToken string
		wantErr     bool
	}{
		{
			name:        "valid access token",
			accessToken: "xxx",
			wantErr:     false,
		},
		{
			name:        "invalid access token",
			accessToken: "yyy",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.CheckAccessToken(tt.accessToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("provider.CheckAccessToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEdgePipelineEnvs(t *testing.T) {
	p := &provider{
		Cfg: &Config{
			IsEdge:       true,
			accessToken:  "xxx",
			PipelineAddr: "pipeline:3081",
			PipelineHost: "pipeline.default.svc.cluster.local",
		},
	}
	envs := p.GetEdgePipelineEnvs()
	assert.Equal(t, "pipeline:3081", envs.Get(apistructs.ClusterDialerDataKeyPipelineAddr))
	assert.Equal(t, "pipeline.default.svc.cluster.local", envs.Get(apistructs.ClusterDialerDataKeyPipelineHost))
}

func TestCheckAccessTokenFromHttpRequest(t *testing.T) {
	p := &provider{
		Cfg: &Config{
			IsEdge:      true,
			accessToken: "xxx",
		},
	}
	tests := []struct {
		name        string
		accessToken string
		wantErr     bool
	}{
		{
			name:        "valid access token",
			accessToken: "xxx",
			wantErr:     false,
		},
		{
			name:        "invalid access token",
			accessToken: "yyy",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{
				Header: http.Header{
					"Authorization": []string{tt.accessToken},
				},
			}
			err := p.CheckAccessTokenFromHttpRequest(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("provider.CheckAccessTokenFromHttpRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsEdge(t *testing.T) {
	p := &provider{
		Cfg: &Config{
			IsEdge:      true,
			accessToken: "xxx",
		},
	}
	assert.Equal(t, true, p.IsEdge())
}
