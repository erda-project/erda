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
	"time"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

type AuditResponse struct {
	Stage string // in-log -> in-update -> in-parsed -> out-log

	inUpdate inUpdate
	inParsed inParsed
}

type inUpdate struct {
	firstResponseAt time.Time
	allChunks       []byte
}

type inParsed struct {
	allChunks                []byte
	completion               string
	responseFunctionCallName string
}

const (
	StageInLog    = "in-log"
	StageInUpdate = "in-update"
	StageInParsed = "in-parsed"
	StageOutLog   = "out-log"
)

var (
	_ filter_define.ProxyResponseModifier = (*AuditResponse)(nil)
)

var ResponseModifierCreator filter_define.ResponseModifierCreator = func(name string, config json.RawMessage) filter_define.ProxyResponseModifier {
	if len(config) == 0 {
		return &AuditResponse{}
	}
	var f AuditResponse
	if err := json.Unmarshal(config, &f); err != nil {
		panic(err)
	}
	return &f
}

func init() {
	filter_define.RegisterFilterCreator("audit", ResponseModifierCreator)
}

func (f *AuditResponse) OnHeaders(resp *http.Response) error {
	switch f.Stage {
	case StageInLog:
		return f.inLogOnHeaders(resp)
	case StageInUpdate:
		return f.inUpdateOnHeaders(resp)
	case StageInParsed:
		return f.inParsedOnHeaders(resp)
	case StageOutLog:
		return f.outLogOnHeaders(resp)
	default:
		ctxhelper.MustGetLogger(resp.Request.Context()).Warnf("invalid audit stage: %s", f.Stage)
	}
	return nil
}

func (f *AuditResponse) OnBodyChunk(resp *http.Response, chunk []byte) (out []byte, err error) {
	switch f.Stage {
	case StageInLog:
		return f.inLogOnBodyChunk(resp, chunk)
	case StageInUpdate:
		return f.inUpdateOnBodyChunk(resp, chunk)
	case StageInParsed:
		return f.inParsedOnBodyChunk(resp, chunk)
	case StageOutLog:
		return f.outLogOnBodyChunk(resp, chunk)
	default:
		ctxhelper.MustGetLogger(resp.Request.Context()).Warnf("invalid audit stage: %s", f.Stage)
	}
	return chunk, nil
}

func (f *AuditResponse) OnComplete(resp *http.Response) ([]byte, error) {
	switch f.Stage {
	case StageInLog:
		return f.inLogOnComplete(resp)
	case StageInUpdate:
		return f.inUpdateOnComplete(resp)
	case StageInParsed:
		return f.inParsedOnComplete(resp)
	case StageOutLog:
		return f.outLogOnComplete(resp)
	default:
		ctxhelper.MustGetLogger(resp.Request.Context()).Warnf("invalid audit stage: %s", f.Stage)
	}
	return nil, nil
}
