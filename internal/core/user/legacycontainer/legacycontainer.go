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

package legacycontainer

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	singletonsMu sync.RWMutex
	singletons   = make(map[string]any)
)

func Register[T any](instance T) {
	key := getTypeKey[T]()
	singletonsMu.Lock()
	defer singletonsMu.Unlock()
	if _, exists := singletons[key]; exists {
		panic(fmt.Sprintf("singleton %s already registered", key))
	}
	singletons[key] = instance
}

func Get[T any]() T {
	key := getTypeKey[T]()
	singletonsMu.RLock()
	defer singletonsMu.RUnlock()
	val, exists := singletons[key]
	if !exists {
		var zero T
		return zero
	}
	return val.(T)
}

func getTypeKey[T any]() string {
	return reflect.TypeOf((*T)(nil)).Elem().String()
}
