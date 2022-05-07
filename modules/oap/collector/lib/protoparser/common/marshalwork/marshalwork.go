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

package marshalwork

import (
	"sync"

	"github.com/erda-project/erda/modules/oap/collector/lib"
)

type MarshalWork interface {
	Marshal()
}

func Schedule(uw MarshalWork) {
	marshalWorkCh <- uw
}

func Start() {
	gomaxProcs := lib.AvailableCPUs()
	marshalWorkCh = make(chan MarshalWork, gomaxProcs)
	marshalWorkWG.Add(gomaxProcs)
	for i := 0; i < gomaxProcs; i++ {
		go func() {
			defer marshalWorkWG.Done()
			for uw := range marshalWorkCh {
				uw.Marshal()
			}
		}()
	}
}

func Stop() {
	close(marshalWorkCh)
	marshalWorkWG.Wait()
	marshalWorkCh = nil
}

var (
	marshalWorkCh chan MarshalWork
	marshalWorkWG sync.WaitGroup
)
