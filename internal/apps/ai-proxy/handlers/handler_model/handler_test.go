package handler_model

import (
	"context"
	"testing"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	pb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/stretchr/testify/assert"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

func TestDesensitizeModel(t *testing.T) {
	type want struct {
		apiKey      string
		secretIsNil bool
	}

	tests := []struct {
		name         string
		setupCtx     func(context.Context)
		buildPayload func() *pb.Model
		want         want
	}{
		{
			name:     "non-admin hides sensitive fields",
			setupCtx: nil,
			buildPayload: func() *pb.Model {
				return &pb.Model{
					ApiKey: "should-hide",
					Metadata: &metadatapb.Metadata{
						Secret: map[string]*structpb.Value{
							"token": structpb.NewStringValue("value"),
						},
					},
				}
			},
			want: want{
				apiKey:      "",
				secretIsNil: true,
			},
		},
		{
			name: "admin keeps sensitive fields",
			setupCtx: func(ctx context.Context) {
				ctxhelper.PutIsAdmin(ctx, true)
			},
			buildPayload: func() *pb.Model {
				return &pb.Model{
					ApiKey: "keep-me",
					Metadata: &metadatapb.Metadata{
						Secret: map[string]*structpb.Value{
							"token": structpb.NewStringValue("value"),
						},
					},
				}
			},
			want: want{
				apiKey:      "keep-me",
				secretIsNil: false,
			},
		},
		{
			name: "client owned model keeps sensitive fields",
			setupCtx: func(ctx context.Context) {
				ctxhelper.PutClient(ctx, &clientpb.Client{Id: "client-1"})
			},
			buildPayload: func() *pb.Model {
				return &pb.Model{
					ClientId: "client-1",
					ApiKey:   "keep-me",
					Metadata: &metadatapb.Metadata{
						Secret: map[string]*structpb.Value{
							"token": structpb.NewStringValue("value"),
						},
					},
				}
			},
			want: want{
				apiKey:      "keep-me",
				secretIsNil: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
			if tt.setupCtx != nil {
				tt.setupCtx(ctx)
			}

			model := tt.buildPayload()
			desensitizeModel(ctx, model)

			assert.Equal(t, tt.want.apiKey, model.ApiKey)
			if model.Metadata != nil {
				if tt.want.secretIsNil {
					assert.Nil(t, model.Metadata.Secret)
				} else {
					assert.NotNil(t, model.Metadata.Secret)
				}
			}
		})
	}
}
