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

package nexus

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

const (
	HeaderAuthorization = "Authorization"
	HeaderContentType   = "Content-Type"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrMissingRepoFormat = errors.New("missing repo format")
)

func (n *Nexus) basicAuthBase64Value() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(n.Username+":"+n.Password))
}

func printJSON(o interface{}) {
	b, _ := json.MarshalIndent(o, "", "  ")
	fmt.Println(string(b))
}

func ErrNotOK(statusCode int, body string) error {
	if statusCode == http.StatusNotFound {
		return ErrNotFound
	}
	return errors.Errorf("status code: %d, err: %v", statusCode, body)
}
