// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package apitestsv2

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/customhttp"
)

func Test_getK8sNamespace(t *testing.T) {
	type args struct {
		k8sHost string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				k8sHost: "a.default.svc.cluster.local",
			},
			want: "default",
		},
		{
			args: args{
				k8sHost: "web.n1.svc.cluster.local",
			},
			want: "n1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getK8sNamespace(tt.args.k8sHost); got != tt.want {
				t.Errorf("getK8sNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_useNetportal(t *testing.T) {
	type args struct {
		url          string
		netportalOpt *netportalOption
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "no netportal info",
			args: args{
				netportalOpt: nil,
			},
			want: false,
		},
		{
			name: "empty netportal url",
			args: args{
				netportalOpt: &netportalOption{url: ""},
			},
			want: false,
		},
		{
			name: "a valid service host",
			args: args{
				url: "http://web.n1.svc.cluster.local:8080",
				netportalOpt: &netportalOption{
					url:                           "netportal:80",
					blacklistOfK8sNamespaceAccess: nil,
				},
			},
			want: true,
		},
		{
			name: "cannot access ns in blacklist",
			args: args{
				url: "http://web.n1.svc.cluster.local:8080",
				netportalOpt: &netportalOption{
					url:                           "netportal:80",
					blacklistOfK8sNamespaceAccess: []string{"n2", "n1"},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := useNetportal(tt.args.url, tt.args.netportalOpt); got != tt.want {
				t.Errorf("useNetportal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleCustomNetportalRequest(t *testing.T) {
	inetAddr := "netportal.default.svc.cluster.local:80"
	netportalURL := "inet://xxx.yyy"
	customhttp.SetInetAddr(inetAddr)
	type args struct {
		apiReq       *apistructs.APIRequestInfo
		netportalOpt *netportalOption
	}
	type want struct {
		schema string
		host   string
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "public network url, do not use netportal",
			args: args{
				apiReq:       &apistructs.APIRequestInfo{URL: "https://www.erda.cloud"},
				netportalOpt: nil,
			},
			want: want{
				schema: "https",
				host:   "www.erda.cloud",
			},
			wantErr: false,
		},
		{
			name: "internal service url, but in blacklist, do not use netportal",
			args: args{
				apiReq: &apistructs.APIRequestInfo{URL: "http://web.n1.svc.cluster.local:8080"},
				netportalOpt: &netportalOption{
					url:                           netportalURL,
					blacklistOfK8sNamespaceAccess: []string{"n2", "n1"},
				},
			},
			want: want{
				schema: "http",
				host:   "web.n1.svc.cluster.local:8080",
			},
			wantErr: false,
		},
		{
			name: "internal service url, NOT in blacklist, use netportal",
			args: args{
				apiReq: &apistructs.APIRequestInfo{URL: "http://web.n1.svc.cluster.local:8080", Headers: make(http.Header)},
				netportalOpt: &netportalOption{
					url:                           netportalURL,
					blacklistOfK8sNamespaceAccess: []string{"n2"},
				},
			},
			want: want{
				schema: "http",
				host:   inetAddr,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleCustomNetportalRequest(tt.args.apiReq, tt.args.netportalOpt)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleCustomNetportalRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_got := want{
				schema: got.URL.Scheme,
				host:   got.Host,
			}
			if !reflect.DeepEqual(_got, tt.want) {
				t.Errorf("handleCustomNetportalRequest() got = %v, want %v", _got, tt.want)
			}
		})
	}
}
