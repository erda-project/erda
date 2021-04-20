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
