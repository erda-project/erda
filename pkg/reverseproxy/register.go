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
