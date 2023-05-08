// Copyright (c) 2023 Terminus, Inc.
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

package reverseproxy

import (
	"encoding/json"
	"fmt"
	"sync"
)

var (
	filterCreators             = make(map[string]FilterCreator)
	registerFilterCreatorMutex = new(sync.Mutex)
)

type FilterCreator func(json.RawMessage) (Filter, error)

func RegisterFilterCreator(name string, creator FilterCreator) {
	registerFilterCreatorMutex.Lock()
	filterCreators[name] = creator
	registerFilterCreatorMutex.Unlock()
}

func GetFilterCreator(name string) (FilterCreator, bool) {
	registerFilterCreatorMutex.Lock()
	defer registerFilterCreatorMutex.Unlock()
	creator, ok := filterCreators[name]
	return creator, ok
}

func MustGetFilterCreator(name string) FilterCreator {
	creator, ok := GetFilterCreator(name)
	if !ok {
		panic(fmt.Sprintf("no such filter creator named %s, do you import it ?", name))
	}
	return creator
}
