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

package common

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type ErrInfo struct {
	Code  string `json:"code,omitempty"`
	Msg   string `json:"msg,omitempty"`
	EnMsg string `json:"-"`
}

type StandardResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Err     *ErrInfo    `json:"err"`
}

type StandaredReturnCode interface {
	GetCode() string
	GetMessage() string
}

func NewStandardResult(succ ...bool) *StandardResult {
	isSucc := true
	if len(succ) != 0 {
		isSucc = succ[0]
	}
	return &StandardResult{Success: isSucc}
}

func (result *StandardResult) SwitchLang(c *gin.Context) *StandardResult {
	lang := c.GetHeader("lang")
	if strings.Contains(lang, "en-US") && result.Err.EnMsg != "" {
		result.Err.Msg = result.Err.EnMsg
	}
	return result
}

func (result *StandardResult) SetReturnCode(returnCode StandaredReturnCode) *StandardResult {
	result.Err = &ErrInfo{
		Code: returnCode.GetCode(),
		Msg:  returnCode.GetMessage(),
	}
	return result
}

func (result *StandardResult) SetSuccessAndData(data interface{}) *StandardResult {
	result.Data = data
	result.Success = true
	return result
}

func (result *StandardResult) SetErrorInfo(errInfo *ErrInfo) *StandardResult {
	result.Err = errInfo
	return result
}
