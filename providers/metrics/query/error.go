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

package query

import (
	"encoding/json"
	"fmt"
)

type returnedErrorFormat struct {
	Success string            `json:"success"`
	Err     map[string]string `json:"err"`
}

type ServerError struct {
	errorCode string
	message   string
	context   string
}

func NewServerError(body []byte) error {
	data := &returnedErrorFormat{}
	if err := json.Unmarshal(body, data); err != nil {
		return err
	}
	return &ServerError{
		errorCode: data.Err["code"],
		message:   data.Err["msg"],
		context:   data.Err["context"],
	}
}

func (e ServerError) Error() string {
	return fmt.Sprintf("SDK.ServerError:\nErrorCode: %s\nContext: %s\nMessage: %s", e.errorCode, e.context, e.message)
}
