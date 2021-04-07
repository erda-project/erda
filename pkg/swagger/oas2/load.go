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

package oas2

import (
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi2"
)

func LoadFromData(data []byte) (*openapi2.Swagger, error) {
	var v2 openapi2.Swagger
	if err := json.Unmarshal(data, &v2); err != nil {
		return nil, err
	}
	return &v2, nil
}
