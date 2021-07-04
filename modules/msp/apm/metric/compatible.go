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

package metric

import "fmt"

func (p *provider) loadCompatibleTKs() error {
	list, err := p.db.ListCompatibleTKs()
	if err != nil {
		return fmt.Errorf("fail to load compatible terminus keys")
	}
	compatibleTKs := make(map[string][]string)
	for _, item := range list {
		if len(item.TerminusKeyRuntime) <= 0 || len(item.TerminusKey) <= 0 || item.TerminusKeyRuntime == item.TerminusKey {
			continue
		}
		compatibleTKs[item.TerminusKey] = append(compatibleTKs[item.TerminusKey], item.TerminusKeyRuntime)
	}
	p.compatibleTKs = compatibleTKs
	if len(compatibleTKs) > 0 {
		p.Log.Debugf("compatible terminus keys: %v", compatibleTKs)
	}
	return nil
}

func (p *provider) getRuntimeTerminusKeys(tk string) (list []string) {
	for _, key := range p.compatibleTKs[tk] {
		list = append(list, key)
	}
	list = append(list, tk)
	return list
}
