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

// Package response release api 响应格式
package response

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/dicehub/errcode"
)

// Response 内部约定的Resposne格式, 更多，请参考: https://yuque.antfin-inc.com/terminus_paas_dev/paas/fozwef
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Err     ErrorMsg    `json:"err,omitempty"`
}

// ErrorMsg release api 错误消息格式
type ErrorMsg struct {
	Code    string      `json:"code,omitempty"`
	Msg     string      `json:"msg,omitempty"`
	Context interface{} `json:"ctx,omitempty"`
}

// Success release api 成功时响应
func Success(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)

	response := Response{
		Success: true,
		Data:    data,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		logrus.Errorf("response parse to json err: %v", err)
	}

	w.Write(jsonResponse)
}

// Error release api 失败时响应
func Error(w http.ResponseWriter, statusCode int, code errcode.ImageErrorCode, errorMsg string) {
	w.Header().Set("Content-Type", "applicaton/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)

	response := &Response{
		Success: false,
		Err: ErrorMsg{
			Code: string(code),
			Msg:  errorMsg,
		},
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		logrus.Errorf("response parse to json err: %v", err)
	}

	w.Write(jsonResponse)
}
