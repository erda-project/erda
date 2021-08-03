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

package template

import (
	context "context"
	reflect "reflect"
	testing "testing"

	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/core/dicehub/template/pb"
)

func Test_templateService_ApplyPipelineTemplate(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.PipelineTemplateApplyRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.PipelineTemplateCreateResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.template.TemplateService",
		//			`
		//erda.core.dicehub.template:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.PipelineTemplateApplyRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.PipelineTemplateCreateResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TemplateServiceServer)
			got, err := srv.ApplyPipelineTemplate(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("templateService.ApplyPipelineTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("templateService.ApplyPipelineTemplate() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_templateService_QueryPipelineTemplates(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.PipelineTemplateQueryRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.PipelineTemplateQueryResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.template.TemplateService",
		//			`
		//erda.core.dicehub.template:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.PipelineTemplateQueryRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.PipelineTemplateQueryResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TemplateServiceServer)
			got, err := srv.QueryPipelineTemplates(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("templateService.QueryPipelineTemplates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("templateService.QueryPipelineTemplates() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_templateService_RenderPipelineTemplate(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.PipelineTemplateRenderRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.PipelineTemplateRenderResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.template.TemplateService",
		//			`
		//erda.core.dicehub.template:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.PipelineTemplateRenderRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.PipelineTemplateRenderResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TemplateServiceServer)
			got, err := srv.RenderPipelineTemplate(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("templateService.RenderPipelineTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("templateService.RenderPipelineTemplate() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_templateService_RenderPipelineTemplateBySpec(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.PipelineTemplateRenderSpecRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.PipelineTemplateRender
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.template.TemplateService",
		//			`
		//erda.core.dicehub.template:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.PipelineTemplateRenderSpecRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.PipelineTemplateRender{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TemplateServiceServer)
			got, err := srv.RenderPipelineTemplateBySpec(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("templateService.RenderPipelineTemplateBySpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("templateService.RenderPipelineTemplateBySpec() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_templateService_GetPipelineTemplateVersion(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.PipelineTemplateVersionGetRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.PipelineTemplateVersion
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.template.TemplateService",
		//			`
		//erda.core.dicehub.template:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.PipelineTemplateVersionGetRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.PipelineTemplateVersion{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TemplateServiceServer)
			got, err := srv.GetPipelineTemplateVersion(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("templateService.GetPipelineTemplateVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("templateService.GetPipelineTemplateVersion() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_templateService_QueryPipelineTemplateVersions(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.PipelineTemplateVersionQueryRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.PipelineTemplateVersion
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.template.TemplateService",
		//			`
		//erda.core.dicehub.template:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.PipelineTemplateVersionQueryRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.PipelineTemplateVersion{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TemplateServiceServer)
			got, err := srv.QueryPipelineTemplateVersions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("templateService.QueryPipelineTemplateVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("templateService.QueryPipelineTemplateVersions() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_templateService_QuerySnippetYml(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.QuerySnippetYmlRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.QuerySnippetYmlResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.template.TemplateService",
		//			`
		//erda.core.dicehub.template:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.QuerySnippetYmlRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.QuerySnippetYmlResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TemplateServiceServer)
			got, err := srv.QuerySnippetYml(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("templateService.QuerySnippetYml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("templateService.QuerySnippetYml() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
