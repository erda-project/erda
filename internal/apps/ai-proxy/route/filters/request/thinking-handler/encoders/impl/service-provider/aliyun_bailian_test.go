package service_provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

func TestBailianThinkingEncoder(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutModelProvider(ctx, &modelproviderpb.ModelProvider{Type: types.ModelProviderTypeAliyunDashScope.String()})

	encoder := &BailianThinkingEncoder{}

	assert.True(t, encoder.CanEncode(ctx))

	result, err := encoder.Encode(ctx, types.CommonThinking{Mode: types.ModePtr(types.ModeOn)})
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		types.FieldEnableThinking: true,
		types.FieldThinkingBudget: 1024,
	}, result)
}

func TestBailianThinkingEncoder_CanEncodeVariants(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutModelProvider(ctx, &modelproviderpb.ModelProvider{Type: types.ModelProviderTypeAliyunBailian.String()})

	encoder := &BailianThinkingEncoder{}

	assert.True(t, encoder.CanEncode(ctx))

	ctx = ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutModelProvider(ctx, &modelproviderpb.ModelProvider{Type: "Unknown"})
	assert.False(t, encoder.CanEncode(ctx))
}
