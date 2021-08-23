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

package spec

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func TestPipeline_EnsureGC(t *testing.T) {
	var ttl uint64 = 10
	var archive bool = true
	p := Pipeline{
		PipelineExtra: PipelineExtra{
			Extra: PipelineExtraInfo{
				GC: apistructs.PipelineGC{
					ResourceGC: apistructs.PipelineResourceGC{
						SuccessTTLSecond: &ttl,
						FailedTTLSecond:  nil,
					},
					DatabaseGC: apistructs.PipelineDatabaseGC{
						Analyzed: apistructs.PipelineDBGCItem{
							NeedArchive: nil,
							TTLSecond:   &ttl,
						},
						Finished: apistructs.PipelineDBGCItem{
							NeedArchive: &archive,
							TTLSecond:   nil,
						},
					},
				},
			},
		},
	}
	p.EnsureGC()
}
