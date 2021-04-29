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

package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
)

// Response .
type Response struct {
	Success bool        `json:"success,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Err     *Error      `json:"err,omitempty"`
}

var _ error = (*Error)(nil)

// Error .
type Error struct {
	Code interface{} `json:"code,omitempty"`
	Msg  interface{} `json:"msg,omitempty"`
	Ctx  interface{} `json:"ctx,omitempty"`
}

func (s *Error) Error() string {
	if s.Ctx != nil {
		return fmt.Sprintf("{code: %s, msg: %s, ctx: %s}", s.Code, s.Msg, s.Ctx)
	}
	return fmt.Sprintf("{code: %s, msg: %s}", s.Code, s.Msg)
}

func encodeError(w http.ResponseWriter, r *http.Request, err error) {
	var status int
	if e, ok := err.(transhttp.Error); ok {
		status = e.HTTPStatus()
	} else {
		status = http.StatusInternalServerError
	}
	w.WriteHeader(status)
	byts, _ := json.Marshal(&Response{
		Success: false,
		Err: &Error{
			Code: strconv.Itoa(status),
			Msg:  err.Error(),
			Ctx:  r.URL.String(),
		},
	})
	w.Write(byts)
}

func wrapResponse(h interceptor.Handler) interceptor.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		resp, err := h(ctx, req)
		if err != nil {
			return resp, err
		}
		return &Response{
			Success: true,
			Data:    resp,
		}, nil
	}
}

// Options .
func Options() transport.ServiceOption {
	return transport.ServiceOption(func(opts *transport.ServiceOptions) {
		transport.WithInterceptors(wrapResponse)(opts)
		transport.WithHTTPOptions(transhttp.WithErrorEncoder(encodeError))(opts)
	})
}
