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
