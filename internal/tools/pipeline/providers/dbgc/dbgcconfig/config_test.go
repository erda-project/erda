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

package dbgcconfig

import (
	"testing"
	"time"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
)

func TestEnsurePipelineRecordGCConfig(t *testing.T) {
	Cfg = &Config{
		AnalyzedPipelineDefaultDatabaseGCTTLDuration: 24 * time.Hour,
		FinishedPipelineDefaultDatabaseGCTTLDuration: 720 * time.Hour,
	}
	cases := []struct {
		name string
		gc   *basepb.PipelineDatabaseGC
	}{
		{name: "nil", gc: nil},
		{name: "empty", gc: &basepb.PipelineDatabaseGC{}},
		{name: "analyzed", gc: &basepb.PipelineDatabaseGC{
			Analyzed: &basepb.PipelineDBGCItem{
				NeedArchive: &[]bool{false}[0],
			},
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			EnsureGCConfig(&c.gc)
			if c.gc.Analyzed.NeedArchive == nil {
				t.Errorf("Analyzed.NeedArchive is nil")
			}
			if c.gc.Analyzed.TTLSecond == nil {
				t.Errorf("Analyzed.TTLSecond is nil")
			}
			if c.gc.Finished.NeedArchive == nil {
				t.Errorf("Finished.NeedArchive is nil")
			}
			if c.gc.Finished.TTLSecond == nil {
				t.Errorf("Finished.TTLSecond is nil")
			}
		})
	}
}
