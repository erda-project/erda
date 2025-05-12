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
	"time"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
)

var (
	Cfg = &Config{}
)

type Config struct {
	// default 2h
	PipelineDBGCDuration time.Duration `file:"pipeline_dbgc_duration" env:"PIPELINE_DBGC_DURATION" default:"2h"`
	// default 1 day
	AnalyzedPipelineArchiveDefaultRetainHour time.Duration `file:"analyzed_pipeline_archive_default_retain_hour" env:"ANALYZED_PIPELINE_ARCHIVE_RETAIN_HOUR" default:"24h"`
	// default 30 day
	FinishedPipelineArchiveDefaultRetainHour time.Duration `file:"finished_pipeline_archive_default_retain_hour" env:"FINISHED_PIPELINE_ARCHIVE_RETAIN_HOUR" default:"720h"`

	// default database gc ttl for analyzed pipeline record: 1 day
	AnalyzedPipelineDefaultDatabaseGCTTLDuration time.Duration `file:"analyzed_pipeline_default_database_gc_ttl_duration" env:"ANALYZED_PIPELINE_DEFAULT_DATABASE_GC_TTL_DURATION" default:"24h"`
	// default database gc ttl for finished pipeline record: 60 day
	FinishedPipelineDefaultDatabaseGCTTLDuration time.Duration `file:"finished_pipeline_default_database_gc_ttl_duration" env:"FINISHED_PIPELINE_DEFAULT_DATABASE_GC_TTL_DURATION" default:"1440h"`
}

func EnsureGCConfig(gc **basepb.PipelineDatabaseGC) {
	if *gc == nil {
		*gc = &basepb.PipelineDatabaseGC{}
	}
	// analyzed part
	if (*gc).Analyzed == nil {
		(*gc).Analyzed = &basepb.PipelineDBGCItem{}
	}
	if (*gc).Analyzed.NeedArchive == nil {
		(*gc).Analyzed.NeedArchive = &[]bool{false}[0]
	}
	if (*gc).Analyzed.TTLSecond == nil {
		(*gc).Analyzed.TTLSecond = &[]uint64{uint64(Cfg.AnalyzedPipelineDefaultDatabaseGCTTLDuration.Seconds())}[0]
	}
	// finished part
	if (*gc).Finished == nil {
		(*gc).Finished = &basepb.PipelineDBGCItem{}
	}
	if (*gc).Finished.NeedArchive == nil {
		(*gc).Finished.NeedArchive = &[]bool{true}[0]
	}
	if (*gc).Finished.TTLSecond == nil {
		(*gc).Finished.TTLSecond = &[]uint64{uint64(Cfg.FinishedPipelineDefaultDatabaseGCTTLDuration.Seconds())}[0]
	}
}
