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

package types

import (
	"context"
	"net/http"
)

type Responser interface {
	GetStatus() int
	GetContent() interface{}
	Raw() bool
}

type Endpoint struct {
	Path    string
	Method  string
	Handler func(context.Context, *http.Request, map[string]string) (Responser, error)
}

type ErrorResponse struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Ctx  interface{} `json:"ctx"`
}

type HTTPResponse struct {
	Status     int
	Error      *ErrorResponse
	Content    interface{}
	Compose    bool // compose response with common structure
	RawContent bool // return raw Content
}

func (r HTTPResponse) GetStatus() int {
	if r.Status == 0 {
		r.Status = 200
	}
	return r.Status
}
func (r HTTPResponse) Raw() bool {
	return r.RawContent
}
func (r HTTPResponse) GetContent() interface{} {
	if r.Compose {
		c := struct {
			Success bool          `json:"success"`
			Data    interface{}   `json:"data"`
			Error   ErrorResponse `json:"err"`
		}{
			Success: r.Error == nil,
		}

		if c.Success {
			c.Data = r.Content
		} else {
			c.Error = *r.Error
		}
		return c
	}
	return r.Content
}
func ErrorResp(code string, msg string) HTTPResponse {
	return HTTPResponse{
		Error: &ErrorResponse{
			Code: code,
			Msg:  msg,
		},
		Compose: true,
	}
}
