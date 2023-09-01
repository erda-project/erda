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

package metadata

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

func (m *Metadata) Scan(src any) error {
	if src == nil {
		return nil
	}
	v, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("invalid src type for metadata, got %T", src)
	}
	if len(v) == 0 {
		return nil
	}
	return json.Unmarshal(v, m)
}

func (m Metadata) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}
