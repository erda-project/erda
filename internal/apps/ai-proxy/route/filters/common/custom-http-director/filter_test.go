package custom_http_director

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	structpb "google.golang.org/protobuf/types/known/structpb"

	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	serviceproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func TestHandleJSONPathTemplate(t *testing.T) {
	ctx := buildJSONPathContext()

	t.Run("no placeholder returns original string", func(t *testing.T) {
		got, err := handleJSONPathTemplate(ctx, "plain-text")
		require.NoError(t, err)
		require.Equal(t, "plain-text", got)
	})

	t.Run("reads provider metadata public values", func(t *testing.T) {
		got, err := handleJSONPathTemplate(ctx, "https://${@provider.metadata.public.host}")
		require.NoError(t, err)
		require.Equal(t, "https://api.example.com", got)
	})

	t.Run("multi choice falls back to literal", func(t *testing.T) {
		got, err := handleJSONPathTemplate(ctx, "${@provider.metadata.secret.missing||fallback-value}")
		require.NoError(t, err)
		require.Equal(t, "fallback-value", got)
	})

	t.Run("missing placeholder produces error", func(t *testing.T) {
		_, err := handleJSONPathTemplate(ctx, "${@provider.metadata.secret.missing}")
		require.Error(t, err)
	})
}

func buildJSONPathContext() context.Context {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	provider := &serviceproviderpb.ServiceProvider{
		Metadata: &metadatapb.Metadata{
			Public: map[string]*structpb.Value{
				"host": structpb.NewStringValue("api.example.com"),
			},
			Secret: map[string]*structpb.Value{
				"api_key": structpb.NewStringValue("secret"),
			},
		},
	}
	ctxhelper.PutServiceProvider(ctx, provider)

	model := &modelpb.Model{
		Metadata: &metadatapb.Metadata{
			Public: map[string]*structpb.Value{
				"api_version": structpb.NewStringValue("2025-03-01-preview"),
			},
		},
	}
	ctxhelper.PutModel(ctx, model)
	return ctx
}
