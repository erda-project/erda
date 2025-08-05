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
	"net/http/httputil"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

// Audit do follow things:
// - print request body for debugging
// - save key request information into database, like: prompt
type Audit struct {
	Stage string // in-log -> in-create-audit -> in-context-parsed -> out
}

const (
	StageInLog           string = "in-log"
	StageInCreateAudit   string = "in-create-audit"
	StageInContextParsed string = "in-context-parsed"

	StageOutLog    string = "out-log"
	StageOutUpdate string = "out-update"
)

var (
	_ filter_define.ProxyRequestRewriter = (*Audit)(nil)
)

var RequestRewriterCreator filter_define.RequestRewriterCreator = func(name string, config json.RawMessage) filter_define.ProxyRequestRewriter {
	if len(config) == 0 {
		return &Audit{}
	}
	var audit Audit
	if err := json.Unmarshal(config, &audit); err != nil {
		panic(err)
	}
	return &audit
}

func init() {
	filter_define.RegisterFilterCreator("audit", RequestRewriterCreator)
}

func (f *Audit) OnProxyRequest(pr *httputil.ProxyRequest) error {
	switch f.Stage {
	case StageInLog:
		return f.requestInLog(pr.In)
	case StageInCreateAudit:
		return f.requestInCreateAudit(pr.In)
	case StageInContextParsed:
		return f.updateAuditAfterContextParsed(pr.In)
	case StageOutLog:
		return f.outLog(pr.Out)
	case StageOutUpdate:
		return f.outUpdate(pr.Out)
	default:
		ctxhelper.MustGetLogger(pr.In.Context()).Warnf("invalid audit stage: %s", f.Stage)
	}
	return nil
}
