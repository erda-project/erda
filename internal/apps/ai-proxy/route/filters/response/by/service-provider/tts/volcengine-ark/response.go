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

package volcengine_ark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/polling"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/tts/ttsutil"
)

type (
	BytedanceTTSSubmitResponse struct {
		Code    int64      `json:"code"` // e.g., 20000000
		Message string     `json:"message"`
		Data    SubmitData `json:"data"`
	}

	SubmitData struct {
		TaskId string `json:"task_id"`
	}

	QueryRequest struct {
		TaskId string `json:"task_id"`
	}

	QueryResponse struct {
		Code    int64             `json:"code"`
		Message string            `json:"message"`
		Data    QueryResponseData `json:"data"`
	}

	QueryResponseData struct {
		TaskStatus int64  `json:"task_status"` // 1: running, 2: success, 3: failure
		AudioURL   string `json:"audio_url"`
	}
)

const (
	SuccessCode = 20000000
)

func (f *VolcengineTTSConverter) OnPeekChunkBeforeHeaders(resp *http.Response, peekBytes []byte) error {
	// parse bytedance response
	var bytedanceSubmitResp BytedanceTTSSubmitResponse
	if err := json.Unmarshal(peekBytes, &bytedanceSubmitResp); err != nil {
		return fmt.Errorf("failed to unmarshal bytedance submit response: %w", err)
	}

	// if upstream returns error, return err
	if bytedanceSubmitResp.Code != SuccessCode {
		return fmt.Errorf("bytedance submit failed, code: %d, message: %s", bytedanceSubmitResp.Code, bytedanceSubmitResp.Message)
	}
	f.taskID = bytedanceSubmitResp.Data.TaskId
	if f.taskID == "" {
		return fmt.Errorf("missing task id from bytedance submit response")
	}

	// record task id for audit
	audithelper.Note(resp.Request.Context(), "doubao_seed_tts_task_id", bytedanceSubmitResp.Data.TaskId)

	// replace body with audio and set headers
	format, _ := ctxhelper.GetAudioTTSResponseFormat(resp.Request.Context())
	contentType := ttsutil.ContentTypeFromFormat(format)
	resp.Header.Set("Content-Type", contentType)

	return nil
}

func (f *VolcengineTTSConverter) OnBodyChunk(resp *http.Response, chunk []byte, index int64) (out []byte, err error) {
	// poll and download audio before sending headers
	audioBinary, err := f.loopQueryTaskStatus(resp.Request.Context(), f.taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to poll audio: %w", err)
	}

	return audioBinary, nil
}

func (f *VolcengineTTSConverter) loopQueryTaskStatus(ctx context.Context, taskId string) ([]byte, error) {
	result, err := polling.Poll(ctx, polling.DefaultConfig(), func(ctx context.Context) polling.Result {
		audioBinary, err := f.onceQueryTaskStatus(ctx, taskId)
		if err != nil {
			return polling.Result{Done: true, Err: fmt.Errorf("query task status failed (task_id: %s): %w", taskId, err)}
		}
		if audioBinary != nil {
			return polling.Result{Done: true, Data: audioBinary}
		}
		return polling.Result{Done: false}
	})
	if err != nil {
		return nil, fmt.Errorf("%w (task_id: %s)", err, taskId)
	}
	return result.([]byte), nil
}

// onceQueryTaskStatus return audio content and error.
func (f *VolcengineTTSConverter) onceQueryTaskStatus(ctx context.Context, taskId string) ([]byte, error) {
	queryURL := "https://openspeech.bytedance.com/api/v3/tts/query"
	var reqBody bytes.Buffer
	if err := json.NewEncoder(&reqBody).Encode(QueryRequest{TaskId: taskId}); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, queryURL, &reqBody)
	if err != nil {
		return nil, err
	}
	model := ctxhelper.MustGetModel(ctx)
	modelMeta := metadata.FromProtobuf(model.Metadata)
	req.Header.Set("x-api-app-id", modelMeta.MustGetValueByKey("app_id"))
	req.Header.Set("x-api-access-key", modelMeta.MustGetValueByKey("access_key"))
	req.Header.Set("x-api-resource-id", modelMeta.MustGetValueByKey("model_name"))
	req.Header.Set("x-api-request-id", ctxhelper.MustGetGeneratedCallID(ctx))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, err
	}
	if respBody.Code != SuccessCode {
		return nil, fmt.Errorf("failed to query task status, code: %d, message: %s", respBody.Code, respBody.Message)
	}
	switch respBody.Data.TaskStatus {
	case 1: // running
		return nil, nil
	case 2: // success
		if respBody.Data.AudioURL == "" {
			return nil, fmt.Errorf("missing audio url from bytedance query response")
		}
		return ttsutil.DownloadAudio(ctx, respBody.Data.AudioURL)
	case 3: // failure
		return nil, fmt.Errorf("task failed, code: %d, message: %s", respBody.Code, respBody.Message)
	default:
		return nil, fmt.Errorf("unknown task status: %d", respBody.Data.TaskStatus)
	}
}
