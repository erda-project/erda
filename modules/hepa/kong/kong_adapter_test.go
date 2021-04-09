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

package kong

import (
	"net/http"
	"testing"

	"github.com/erda-project/erda/modules/hepa/kong/base"
)

func TestKongAdapterImpl_checkPluginEnabled(t *testing.T) {
	type fields struct {
		kongAddr string
		client   *http.Client
	}
	type args struct {
		pluginName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "case1",
			fields: fields{
				kongAddr: "http://localhost:8001",
				client:   &http.Client{},
			},
			args: args{
				pluginName: "host-check",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "case2",
			fields: fields{
				kongAddr: "http://localhost:8001",
				client:   &http.Client{},
			},
			args: args{
				pluginName: "xxx",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &base.KongAdapterImpl{
				KongAddr: tt.fields.kongAddr,
				Client:   tt.fields.client,
			}
			got, err := impl.CheckPluginEnabled(tt.args.pluginName)
			if (err != nil) != tt.wantErr {
				t.Errorf("KongAdapterImpl.checkPluginEnabled() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("KongAdapterImpl.checkPluginEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
