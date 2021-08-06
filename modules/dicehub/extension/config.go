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

package extension

import "encoding/json"

func (c *config) ConvertConfigExtensionMenu() (map[string][]string, error) {
	var mp map[string][]string
	err := json.Unmarshal([]byte(c.ExtensionMenu), &mp)
	if err != nil {
		return nil, err
	}
	if mp == nil {
		mp = make(map[string][]string)
	}
	return mp, nil
}
