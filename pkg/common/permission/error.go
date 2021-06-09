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

package permission

import (
	"fmt"
	"net/http"

	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
)

// PermError .
type PermError struct {
	method string
	msg    string
}

var _ transhttp.Error = (*PermError)(nil)

// NewPermError .
func NewPermError(method, msg string) error {
	return &PermError{
		method: method,
		msg:    msg,
	}
}

func (e *PermError) HTTPStatus() int {
	return http.StatusUnauthorized
}

func (e *PermError) Error() string {
	if len(e.msg) > 0 {
		return fmt.Sprintf("permission denied to call %q: %s", e.method, e.msg)
	}
	return fmt.Sprintf("permission denied to call %q", e.method)
}
