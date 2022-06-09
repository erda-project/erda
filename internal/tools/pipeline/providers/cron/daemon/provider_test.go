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

package daemon

import (
	"context"
	"net/http"
	"testing"

	"github.com/coreos/etcd/clientv3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	db2 "github.com/erda-project/erda/internal/tools/pipeline/providers/cron/db"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/edgepipeline_register"
)

type EdgePipelineRegister struct {
}

func (e EdgePipelineRegister) ClusterAccessKey() string {
	panic("implement me")
}

func (e EdgePipelineRegister) GetAccessToken(req apistructs.OAuth2TokenGetRequest) (*apistructs.OAuth2Token, error) {
	panic("implement me")
}

func (e EdgePipelineRegister) GetOAuth2Token(req apistructs.OAuth2TokenGetRequest) (*apistructs.OAuth2Token, error) {
	panic("implement me")
}

func (e EdgePipelineRegister) GetEdgePipelineEnvs() apistructs.ClusterManagerClientDetail {
	panic("implement me")
}

func (e EdgePipelineRegister) CheckAccessToken(token string) error {
	panic("implement me")
}

func (e EdgePipelineRegister) CheckAccessTokenFromHttpRequest(req *http.Request) error {
	panic("implement me")
}

func (e EdgePipelineRegister) IsEdge() bool {
	return false
}

func (e EdgePipelineRegister) IsCenter() bool {
	panic("implement me")
}

func (e EdgePipelineRegister) CanProxyToEdge(source apistructs.PipelineSource, clusterName string) bool {
	return false
}

func (e EdgePipelineRegister) GetEdgeBundleByClusterName(clusterName string) (*bundle.Bundle, error) {
	panic("implement me")
}

func (e EdgePipelineRegister) ClusterIsEdge(clusterName string) (bool, error) {
	panic("implement me")
}

func (e EdgePipelineRegister) OnEdge(f func(context.Context)) {
	panic("implement me")
}

func (e EdgePipelineRegister) OnCenter(f func(context.Context)) {
	panic("implement me")
}

type mockKV struct{}

func (o mockKV) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return nil, nil
}
func (o mockKV) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	panic("implement me")
}
func (o mockKV) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	panic("implement me")
}
func (o mockKV) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	panic("implement me")
}
func (o mockKV) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	panic("implement me")
}
func (o mockKV) Txn(ctx context.Context) clientv3.Txn {
	panic("implement me")
}

func Test_provider_AddIntoPipelineCrond(t *testing.T) {
	type fields struct {
		EtcdClient           *clientv3.Client
		EdgePipelineRegister edgepipeline_register.Interface
	}
	type args struct {
		cron *db2.PipelineCron
	}

	etcdClient := &clientv3.Client{
		KV: mockKV{},
	}
	edgePipelineRegister := EdgePipelineRegister{}
	enable := false
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test with add",
			fields: fields{
				EtcdClient:           etcdClient,
				EdgePipelineRegister: edgePipelineRegister,
			},
			args: args{
				cron: &db2.PipelineCron{
					ID:       1,
					CronExpr: "*/1 * * * *",
					Enable:   &enable,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				EtcdClient:           tt.fields.EtcdClient,
				EdgePipelineRegister: tt.fields.EdgePipelineRegister,
			}
			if err := p.AddIntoPipelineCrond(tt.args.cron); (err != nil) != tt.wantErr {
				t.Errorf("AddIntoPipelineCrond() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_provider_DeleteFromPipelineCrond(t *testing.T) {
	type fields struct {
		EtcdClient           *clientv3.Client
		EdgePipelineRegister edgepipeline_register.Interface
	}
	type args struct {
		cron *db2.PipelineCron
	}

	etcdClient := &clientv3.Client{KV: mockKV{}}
	edgePipelineRegister := EdgePipelineRegister{}
	enable := false
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test with delete",
			fields: fields{
				EtcdClient:           etcdClient,
				EdgePipelineRegister: edgePipelineRegister,
			},
			args: args{
				cron: &db2.PipelineCron{
					ID:       1,
					CronExpr: "*/1 * * * *",
					Enable:   &enable,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				EtcdClient:           tt.fields.EtcdClient,
				EdgePipelineRegister: tt.fields.EdgePipelineRegister,
			}
			if err := p.DeleteFromPipelineCrond(tt.args.cron); (err != nil) != tt.wantErr {
				t.Errorf("DeleteFromPipelineCrond() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
