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

package httputils

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/http/httpclient"
)

type Resp struct {
	Success bool
	Data    json.RawMessage
	Err     *Err
}

type Page struct {
	Total int
	List  json.RawMessage
}

type Err struct {
	Code string
	Msg  string
	Ctx  interface{}
}

func (resp *Resp) ParseData(o interface{}) error {
	if !resp.Success {
		return errors.Errorf("response failed, code=%v,msg=%v,ctx=%v", resp.Err.Code, resp.Err.Msg, resp.Err.Ctx)
	}
	if o == nil {
		return nil
	}
	if err := json.Unmarshal([]byte(resp.Data), o); err != nil {
		return errors.Wrapf(err, "parse data, resp data not json, data=%v", string(resp.Data))
	}
	return nil
}

func (resp *Resp) ParsePagingListData(o interface{}) error {
	if !resp.Success {
		return errors.Errorf(" get paging resp , code=%v,msg=%v,ctx=%v", resp.Err.Code, resp.Err.Msg, resp.Err.Ctx)
	}
	var page Page
	if o == nil {
		return nil
	}
	if err := json.Unmarshal([]byte(resp.Data), &page); err != nil {
		return errors.Wrapf(err, "parse paging resp paging data not json, data=%v", string(resp.Data))
	}
	if err := json.Unmarshal([]byte(page.List), &o); err != nil {
		return errors.Wrapf(err, "parse paging list data not json, data=%v", string(page.List))
	}
	return nil
}

func DoResp(r *httpclient.Request) (Resp, error) {
	var b bytes.Buffer
	response, err := r.Do().Body(&b)
	if err != nil {
		return Resp{}, err
	}
	if !response.IsOK() {
		return Resp{}, fmt.Errorf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
			response.StatusCode(), response.ResponseHeader("Content-Type"), b.String())
	}
	var resp Resp

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return Resp{}, fmt.Errorf("response not json, raw body %s", b.String())
	}
	return resp, nil
}
