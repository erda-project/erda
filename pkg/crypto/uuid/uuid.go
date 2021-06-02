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

package uuid

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

// Generate 不要再调用这个函数，太丑了，找时间废除.
func Generate() string {
	u := uuid.NewV4()
	return fmt.Sprintf("%x%x%x%x%x", u[:4], u[4:6], u[6:8], u[8:10], u[10:])
}

// UUID 返回 uuid.
func UUID() string {
	return Generate()
}
