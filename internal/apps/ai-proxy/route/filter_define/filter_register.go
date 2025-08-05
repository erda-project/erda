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

package filter_define

import (
	"encoding/json"
	"fmt"
	"sync"
)

type FilterFactoryMap struct {
	RequestFilters  map[string]RequestRewriterCreator
	ResponseFilters map[string]ResponseModifierCreator
}

var FilterFactory = &FilterFactoryMap{
	RequestFilters:  make(map[string]RequestRewriterCreator),
	ResponseFilters: make(map[string]ResponseModifierCreator),
}

type RequestRewriterCreator func(name string, _ json.RawMessage) ProxyRequestRewriter
type ResponseModifierCreator func(name string, _ json.RawMessage) ProxyResponseModifier

var registerFilterCreatorLock sync.Mutex

func RegisterFilterCreator[T RequestRewriterCreator | ResponseModifierCreator](name string, creator T) {
	registerFilterCreatorLock.Lock()
	defer registerFilterCreatorLock.Unlock()

	switch c := any(creator).(type) {
	case RequestRewriterCreator:
		if _, exist := FilterFactory.RequestFilters[name]; exist {
			panic(fmt.Errorf("request filter %s duplicated", name))
		}
		FilterFactory.RequestFilters[name] = c
	case ResponseModifierCreator:
		if _, exist := FilterFactory.ResponseFilters[name]; exist {
			panic(fmt.Errorf("response filter %s duplicated", name))
		}
		FilterFactory.ResponseFilters[name] = c
	default:
		panic(fmt.Errorf("unsupported filter creator type for %s", name))
	}
}
