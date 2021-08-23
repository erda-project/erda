// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
