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

package profile

import (
	"encoding/json"

	"github.com/pyroscope-io/pyroscope/pkg/ingestion"
	"github.com/pyroscope-io/pyroscope/pkg/storage/segment"
)

type ProfileIngest struct {
	Format      ingestion.Format
	Metadata    ingestion.Metadata
	RawData     []byte
	Key         string
	ContentType string
}

func (p *ProfileIngest) Hash() uint64 {
	return 0
}

func (p *ProfileIngest) GetTags() map[string]string {
	if p.Metadata.Key == nil {
		p.Metadata.Key, _ = segment.ParseKey(p.Key)
	}
	return p.Metadata.Key.Labels()
}

func (p *ProfileIngest) String() string {
	bytes, _ := json.Marshal(p)
	return string(bytes)
}

// Copy instance
func (p *ProfileIngest) Copy() *ProfileIngest {
	copyRawData := make([]byte, len(p.RawData))
	copy(copyRawData, p.RawData)
	copied := &ProfileIngest{
		Format:  p.Format,
		RawData: copyRawData,
		Metadata: ingestion.Metadata{
			StartTime:       p.Metadata.StartTime,
			EndTime:         p.Metadata.EndTime,
			SpyName:         p.Metadata.SpyName,
			SampleRate:      p.Metadata.SampleRate,
			Units:           p.Metadata.Units,
			AggregationType: p.Metadata.AggregationType,
		},
		Key:         p.Key,
		ContentType: p.ContentType,
	}
	p.Metadata.Key, _ = segment.ParseKey(p.Key)
	return copied
}

type TableMain struct {
	Key   string `ch:"k"`
	Value string `ch:"v"`
}

type TableProfile struct {
	Key   string `ch:"k"`
	Value string `ch:"v"`
}

type TableDict struct {
	Key   string `ch:"k"`
	Value string `ch:"v"`
}

type TableDimension struct {
	Key   string `ch:"k"`
	Value string `ch:"v"`
}

type TableSegment struct {
	Key   string `ch:"k"`
	Value string `ch:"v"`
}

type TableTree struct {
	Key   string `ch:"k"`
	Value string `ch:"v"`
}
