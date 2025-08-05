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

package audit

import (
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

// updateAuditAfterContextParsed is called after the request context has been parsed, and before director invoked.
// Called once after all context parsers have been executed to reduce the number of database operations.
func (f *Audit) updateAuditAfterContextParsed(in *http.Request) error {
	auditRecID, ok := ctxhelper.GetAuditID(in.Context())
	if !ok || auditRecID == "" {
		return nil
	}

	var updateReq pb.AuditUpdateRequestAfterSpecificContextParsed
	updateReq.AuditId = auditRecID
	// prompt
	prompt, _ := ctxhelper.GetUserPrompt(in.Context())
	updateReq.Prompt = prompt

	// metadata, routing by request path
	switch ctxhelper.MustGetPathMatcher(in.Context()).Pattern {
	case common.RequestPathPrefixV1ChatCompletions, common.RequestPathPrefixV1Completions:
		// TODO get from ctxhelper CtxKeyMap directly, parsed at concrete context parser
		//var openaiReq openai.ChatCompletionRequest
		//bodyCopy, err := body_util.SmartCloneBody(&pr.In.Body, body_util.MaxSample)
		//if err != nil {
		//	return fmt.Errorf("failed to clone request body: %w", err)
		//}
		//if err := json.NewDecoder(bodyCopy).Decode(&openaiReq); err != nil {
		//	goto Next
		//}
		//updateReq.RequestFunctionCallName = func() string {
		//	// switch type
		//	switch openaiReq.FunctionCall.(type) {
		//	case string:
		//		return openaiReq.FunctionCall.(string)
		//	case map[string]interface{}:
		//		var reqFuncCall openai.FunctionCall
		//		cputil.MustObjJSONTransfer(openaiReq.FunctionCall, &reqFuncCall)
		//		return reqFuncCall.Name
		//	case openai.FunctionCall:
		//		return openaiReq.FunctionCall.(openai.FunctionCall).Name
		//	case nil:
		//		return ""
		//	default:
		//		return fmt.Sprintf("%v", openaiReq.FunctionCall)
		//	}
		//}()
	case common.RequestPathPrefixV1Audio:
		audioInfo, ok := ctxhelper.GetAudioInfo(in.Context())
		if ok {
			updateReq.AudioFileName = audioInfo.FileName
			updateReq.AudioFileSize = audioInfo.FileSize.String()
			updateReq.AudioFileHeaders = func() string {
				b, err := json.Marshal(audioInfo.FileHeaders)
				if err != nil {
					return err.Error()
				}
				return string(b)
			}()
		}
	case common.RequestPathPrefixV1Images:
		imageInfo, ok := ctxhelper.GetImageInfo(in.Context())
		if ok {
			updateReq.ImageQuality = imageInfo.ImageQuality
			updateReq.ImageSize = imageInfo.ImageSize
			updateReq.ImageStyle = imageInfo.ImageStyle
		}
	case common.RequestPathPrefixV1Assistants:

	default:
		// do nothing
	}

	// update audit into db
	_, err := ctxhelper.MustGetDBClient(in.Context()).AuditClient().UpdateAfterSpecificContextParsed(in.Context(), &updateReq)
	if err != nil {
		// log it
		l := ctxhelper.MustGetLogger(in.Context())
		l.Warnf("failed to update audit (id=%s): %v", auditRecID, err)
	}
	return nil
}
