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
	"sync"
	"time"

	"github.com/pyroscope-io/pyroscope/pkg/ingestion"
	"github.com/pyroscope-io/pyroscope/pkg/model/appmetadata"
	"github.com/pyroscope-io/pyroscope/pkg/storage/segment"
	"github.com/pyroscope-io/pyroscope/pkg/storage/tree"
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

type Output struct {
	sync.Mutex
	profiles map[string]*OutputIngest
}

func NewOutput() *Output {
	return &Output{
		profiles: make(map[string]*OutputIngest),
	}
}

func (o *Output) Add(k string, v *OutputIngest) {
	o.Lock()
	defer o.Unlock()
	o.profiles[k] = v
}

func (o *Output) Profiles() map[string]*OutputIngest {
	o.Lock()
	defer o.Unlock()
	return o.profiles
}

type OutputIngest struct {
	TreeKey     string
	SegmentKey  string
	CollectTime *time.Time
	Tree        *tree.Tree
	Segment     *segment.Segment
	App         *appmetadata.ApplicationMetadata
}

func (o *Output) Hash() uint64 {
	return 0
}

func (o *Output) GetTags() map[string]string {
	if len(o.profiles) == 0 {
		return map[string]string{}
	}
	for _, v := range o.profiles {
		if v.App != nil {
			return map[string]string{
				"org_name":            v.App.OrgName,
				"DICE_ORG_ID":         v.App.OrgID,
				"DICE_ORG_NAME":       v.App.OrgName,
				"POD_IP":              v.App.PodIP,
				"DICE_PROJECT_ID":     v.App.ProjectID,
				"DICE_APPLICATION_ID": v.App.AppID,
				"DICE_SERVICE":        v.App.ServiceName,
			}
		}
	}
	return map[string]string{}
}

func (o *Output) String() string {
	bytes, _ := json.Marshal(o)
	return string(bytes)
}

type TableGeneral struct {
	K         string    `db:"k" ch:"k"`
	V         string    `db:"v" ch:"v"`
	Timestamp time.Time `db:"timestamp" ch:"timestamp"`
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
