// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package logic

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/actionagent"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/apitestsv2"
	"github.com/erda-project/erda/pkg/apitestsv2/cookiejar"
	"github.com/erda-project/erda/pkg/encoding/jsonparse"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

const (
	ResultSuccess = "success"
	ResultFailed  = "failed"
)

const (
	MetaKeyResult           = "result"
	metaKeyAPIRequest       = "api_request"
	metaKeyAPIResponse      = "api_response"
	metaKeyAPISetCookie     = "api_set_cookie"
	metaKeyAPIAssertSuccess = "api_assert_success" // true; false
	metaKeyAPIAssertDetail  = "api_assert_detail"
)

type Meta struct {
	Result          string
	AssertResult    bool
	AssertDetail    string
	Req             *apistructs.APIRequestInfo
	Resp            *apistructs.APIResp
	OutParamsDefine []apistructs.APIOutParam
	CookieJar       cookiejar.Cookies
	OutParamsResult map[string]interface{}
}

func NewMeta() *Meta {
	return &Meta{
		OutParamsResult: map[string]interface{}{},
	}
}

type KVs []kv
type kv struct {
	k string
	v string
}

func (kvs *KVs) add(k, v string) {
	*kvs = append(*kvs, kv{k, v})
}

func writeMetaFile(ctx context.Context, task *spec.PipelineTask, meta *Meta) {
	log := clog(ctx)

	// kvs 保证顺序
	kvs := &KVs{}

	kvs.add(MetaKeyResult, meta.Result)
	if meta.AssertDetail != "" {
		kvs.add(metaKeyAPIAssertSuccess, strconv.FormatBool(meta.AssertResult))
		kvs.add(metaKeyAPIAssertDetail, meta.AssertDetail)
	}
	if meta.Req != nil {
		kvs.add(metaKeyAPIRequest, jsonOneLine(ctx, meta.Req))
	}
	if meta.Resp != nil {
		kvs.add(metaKeyAPIResponse, jsonOneLine(ctx, meta.Resp))
		if len(meta.Resp.Headers) > 0 {
			if headerSetCookie := meta.Resp.Headers[apitestsv2.HeaderSetCookie]; len(headerSetCookie) > 0 {
				kvs.add(metaKeyAPISetCookie, jsonparse.JsonOneLine(headerSetCookie)) // format: []string
			}
		}
	}
	if len(meta.OutParamsResult) > 0 {
		for _, define := range meta.OutParamsDefine {
			v, ok := meta.OutParamsResult[define.Key]
			if !ok {
				continue
			}
			kvs.add(define.Key, jsonOneLine(ctx, v))
		}
	}

	var fields []*apistructs.MetadataField
	for _, kv := range *kvs {
		fields = append(fields, &apistructs.MetadataField{Name: kv.k, Value: kv.v})
	}

	var cb actionagent.Callback
	cb.AppendMetadataFields(fields)
	cb.PipelineID = task.PipelineID
	cb.PipelineTaskID = task.ID

	cbData, _ := json.Marshal(&cb)
	var cbReq apistructs.PipelineCallbackRequest
	cbReq.Type = string(apistructs.PipelineCallbackTypeOfAction)
	cbReq.Data = cbData

	// update task result through internal api
	var resp apistructs.PipelineCallbackResponse
	r, err := httpclient.New().
		Post("localhost"+conf.ListenAddr()).
		Path("/api/pipelines/actions/callback").
		Header("Internal-Client", "action executor").
		JSONBody(&cbReq).
		Do().
		JSON(&resp)
	if err != nil {
		log.Errorf("failed to callback, err: %v", err)
		return
	}
	if !r.IsOK() || !resp.Success {
		log.Errorf("failed to callback, status-code %d, resp %#v", r.StatusCode(), resp)
		return
	}
	return
}
