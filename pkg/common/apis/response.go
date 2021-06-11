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
	"reflect"
	"runtime"
	"strconv"

	validator "github.com/mwitkow/go-proto-validators"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/modcom"
	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	"github.com/erda-project/erda-infra/providers/i18n"
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

// I18n .
var I18n i18n.I18n

func init() {
	modcom.RegisterHubListener(&servicehub.DefaultListener{
		BeforeInitFunc: func(h *servicehub.Hub, config map[string]interface{}) error {
			if _, ok := config["i18n"]; !ok {
				config["i18n"] = nil // i18n is required
			}
			return nil
		},
		AfterInitFunc: func(h *servicehub.Hub) error {
			I18n = h.Service("i18n").(i18n.I18n)
			return nil
		},
	})
}

var validateErrorType = reflect.TypeOf(validator.FieldError("", nil))

func encodeError(w http.ResponseWriter, r *http.Request, err error) {
	var status int
	if e, ok := err.(transhttp.Error); ok {
		status = e.HTTPStatus()
	} else {
		typ := reflect.TypeOf(err)
		if typ == validateErrorType {
			status = http.StatusBadRequest
		} else {
			status = http.StatusInternalServerError
		}
	}
	var msg string
	if e, ok := err.(i18n.Internationalizable); I18n != nil && ok {
		msg = e.Translate(I18n.Translator("apis"), api.Language(r))
	} else {
		msg = err.Error()
	}
	w.WriteHeader(status)
	byts, _ := json.Marshal(&Response{
		Success: false,
		Err: &Error{
			Code: strconv.Itoa(status),
			Msg:  msg,
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
		if resp != nil {
			val := reflect.ValueOf(resp)
			for val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			if val.Kind() == reflect.Struct {
				var fields int
				for i, n := 0, val.NumField(); i < n; i++ {
					field := val.Field(i)
					if field.CanSet() {
						fields++
					}
				}
				if fields == 1 {
					field := val.FieldByName("Data")
					if field.IsValid() {
						resp = field.Interface()
					}
				}
			}
		}
		return &Response{
			Success: true,
			Data:    resp,
		}, nil
	}
}

func validRequest(h interceptor.Handler) interceptor.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		if v, ok := req.(validator.Validator); ok {
			err := v.Validate()
			if err != nil {
				return nil, err
				// return nil, errors.ParseValidateError(err)
			}
		}
		return h(ctx, req)
	}
}

// Options .
func Options() transport.ServiceOption {
	return transport.ServiceOption(func(opts *transport.ServiceOptions) {
		transport.WithInterceptors(validRequest)(opts)
		transport.WithHTTPOptions(transhttp.WithInterceptor(wrapResponse))(opts)
		transport.WithHTTPOptions(transhttp.WithErrorEncoder(encodeError))(opts)
	})
}

func getMethodFullName(method interface{}) string {
	if method == nil {
		return ""
	}
	name, ok := method.(string)
	if ok {
		return name
	}
	val := reflect.ValueOf(method)
	if val.Kind() != reflect.Func {
		panic(fmt.Errorf("method %V not function", method))
	}
	fn := runtime.FuncForPC(val.Pointer())
	return fn.Name()
}
