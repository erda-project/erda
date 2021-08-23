package accesskey

import (
	context "context"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/core/services/accesskey/pb"
	reflect "reflect"
	testing "testing"
)

func Test_accessKeyService_ListAccessKeys(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ListAccessKeysRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ListAccessKeysResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			"case 1",
			"erda.core.services.accesskey.AccessKeyService",
			`
erda.core.services.accesskey:
`,
			args{
				context.TODO(),
				&pb.ListAccessKeysRequest{
					// TODO: setup fields
				},
			},
			&pb.ListAccessKeysResponse{
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
			srv := hub.Service(tt.service).(pb.AccessKeyServiceServer)
			got, err := srv.ListAccessKeys(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accessKeyService.ListAccessKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("accessKeyService.ListAccessKeys() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
