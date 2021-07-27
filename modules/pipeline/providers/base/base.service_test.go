package base

import (
	context "context"
	reflect "reflect"
	testing "testing"

	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
)

func Test_baseService_PipelineCreate(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.PipelineCreateRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.PipelineCreateResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			"case 1",
			"erda.core.pipeline.base.BaseService",
			`
erda.core.pipeline.base:
`,
			args{
				context.TODO(),
				&pb.PipelineCreateRequest{
					// TODO: setup fields
				},
			},
			&pb.PipelineCreateResponse{
				// TODO: setup fields.
			},
			false,
		},
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
			srv := hub.Service(tt.service).(pb.BaseServiceServer)
			got, err := srv.PipelineCreate(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("baseService.PipelineCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("baseService.PipelineCreate() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
