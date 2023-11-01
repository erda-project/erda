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

package bailian_director

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func (f *BailianDirector) OnResponseChunk(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (signal reverseproxy.Signal, err error) {
	f.DefaultResponseFilter.Buffer.Write(chunk)
	return reverseproxy.Continue, nil
}

func (f *BailianDirector) OnResponseEOF(ctx context.Context, _ reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (err error) {
	// convert bailian response to openai response
	var bailianResp CompletionResponse
	if err := json.Unmarshal(f.DefaultResponseFilter.Buffer.Bytes(), &bailianResp); err != nil {
		return fmt.Errorf("failed to unmarshal bailian response, err: %v", err)
	}
	model, _ := ctxhelper.GetModel(ctx)
	openaiResp, err := convertToOpenaiResponse(model, bailianResp)
	if err != nil {
		return fmt.Errorf("failed to convert bailian response to openai response, err: %v", err)
	}
	b, err := json.Marshal(openaiResp)
	if err != nil {
		return fmt.Errorf("failed to marshal openai response, err: %v", err)
	}
	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("failed to write openai response, err: %v", err)
	}
	return err
}

func convertToOpenaiResponse(model *modelpb.Model, bailianResponse CompletionResponse) (*openai.CompletionResponse, error) {
	openaiResp := &openai.CompletionResponse{
		ID:    *bailianResponse.Data.ResponseId,
		Model: model.Name,
		Choices: []openai.CompletionChoice{
			{
				Text:         *bailianResponse.Data.Text,
				Index:        0,
				FinishReason: string(openai.FinishReasonStop),
			},
		},
	}
	return openaiResp, nil
}
