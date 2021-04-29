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
	"encoding/json"
	"fmt"
	"net/http"
)

var _ error = (*HTTPError)(nil)

// HTTPError represents an error that occurred while handling a request.
type HTTPError struct {
	Code     int
	Message  interface{}
	Internal error // Stores the error returned by an external dependency
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("code:%d msg:%q", e.Code, e.Message)
}

// WriteError .
func WriteError(err error, rw http.ResponseWriter) (int, error) {
	if err != nil {
		if herr, ok := err.(*HTTPError); ok {
			rw.Header().Set("Content-Type", "application/json")
			byts, _ := json.Marshal(herr)
			return rw.Write(byts)
		}
		return rw.Write([]byte(err.Error()))
	}
	return 0, nil
}
