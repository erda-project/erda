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
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	storageconfig "github.com/pyroscope-io/pyroscope/pkg/config"
	"github.com/pyroscope-io/pyroscope/pkg/convert/jfr"
	"github.com/pyroscope-io/pyroscope/pkg/convert/pprof"
	convertprofile "github.com/pyroscope-io/pyroscope/pkg/convert/profile"
	"github.com/pyroscope-io/pyroscope/pkg/convert/speedscope"
	"github.com/pyroscope-io/pyroscope/pkg/exporter"
	"github.com/pyroscope-io/pyroscope/pkg/ingestion"
	"github.com/pyroscope-io/pyroscope/pkg/model/appmetadata"
	"github.com/pyroscope-io/pyroscope/pkg/parser"
	"github.com/pyroscope-io/pyroscope/pkg/storage"
	"github.com/pyroscope-io/pyroscope/pkg/storage/metadata"
	"github.com/pyroscope-io/pyroscope/pkg/storage/segment"
	"github.com/pyroscope-io/pyroscope/pkg/storage/tree"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/tools/monitor/core/log"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/core/profile"
	profileinput "github.com/erda-project/erda/internal/tools/monitor/core/profile"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins"
)

var providerName = plugins.WithPrefixProcessor("profile")

type config struct {
	Keypass map[string][]string `file:"keypass"`
}

var _ model.Processor = (*provider)(nil)

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger

	profileParser *parser.Parser
}

func (p *provider) ComponentClose() error {
	return nil
}

func (p *provider) ComponentConfig() interface{} {
	return p.Cfg
}

func (p *provider) ProcessMetric(item *metric.Metric) (*metric.Metric, error) { return item, nil }
func (p *provider) ProcessLog(item *log.Log) (*log.Log, error)                { return item, nil }
func (p *provider) ProcessSpan(item *trace.Span) (*trace.Span, error)         { return item, nil }
func (p *provider) ProcessRaw(item *odata.Raw) (*odata.Raw, error)            { return item, nil }
func (p *provider) ProcessProfile(item *profile.ProfileIngest) (*profile.Output, error) {
	ingestInput := &ingestion.IngestInput{
		Format:   item.Format,
		Metadata: item.Metadata,
	}
	p.genRawProfile(ingestInput, item)
	res := profile.NewOutput()

	if err := p.profileParser.Ingest(context.WithValue(context.Background(), "output", res), ingestInput); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *provider) Put(ctx context.Context, pi *storage.PutInput) error {
	output, ok := ctx.Value("output").(*profile.Output)
	if !ok {
		return fmt.Errorf("failed to get output from context")
	}
	if err := segment.ValidateKey(pi.Key); err != nil {
		return err
	}
	ingest := &profile.OutputIngest{}

	appList := strings.Split(pi.Key.AppName(), ".")
	app := &appmetadata.ApplicationMetadata{
		FQName:          pi.Key.AppName(),
		SpyName:         pi.SpyName,
		SampleRate:      pi.SampleRate,
		Units:           pi.Units,
		AggregationType: pi.AggregationType,
		SampleType:      appList[len(appList)-1],
		OrgID:           pi.Key.Labels()[apistructs.EnvDiceOrgID],
		OrgName:         pi.Key.Labels()[apistructs.EnvDiceOrgName],
		Workspace:       pi.Key.Labels()[apistructs.EnvDiceWorkspace],
		ProjectID:       pi.Key.Labels()[apistructs.EnvDiceProjectID],
		ProjectName:     pi.Key.Labels()[apistructs.EnvDiceProjectName],
		AppID:           pi.Key.Labels()[apistructs.EnvDiceApplicationID],
		AppName:         pi.Key.Labels()[apistructs.EnvDiceApplicationName],
		ClusterName:     pi.Key.Labels()[apistructs.EnvDiceClusterName],
		ServiceName:     pi.Key.Labels()[apistructs.EnvDiceServiceName],
		PodIP:           pi.Key.Labels()[apistructs.EnvPodIP],
	}

	sk := pi.Key.SegmentKey()
	skWithTime := fmt.Sprintf("%s:%d", sk, pi.EndTime.Unix())
	skKey := sk

	st := segment.New()
	st.SetMetadata(metadata.Metadata{
		SpyName:         pi.SpyName,
		SampleRate:      pi.SampleRate,
		Units:           pi.Units,
		AggregationType: pi.AggregationType,
	})
	ingest.SegmentKey = skKey
	ingest.CollectTime = &pi.StartTime
	ingest.App = app

	samples := pi.Val.Samples()
	err := st.Put(pi.StartTime, pi.EndTime, samples, func(depth int, t time.Time, r *big.Rat, addons []segment.Addon) {
		tk := pi.Key.TreeKey()
		cachedTree := tree.New()
		treeClone := pi.Val.Clone(r)
		cachedTree.Lock()
		cachedTree.Merge(treeClone)
		cachedTree.Unlock()
		ingest.Tree = cachedTree
		ingest.TreeKey = tk
	})
	if err != nil {
		return err
	}

	ingest.Segment = st
	output.Add(skWithTime, ingest)
	return nil
}

func (p *provider) genRawProfile(input *ingestion.IngestInput, data *profileinput.ProfileIngest) {
	input.Metadata.Key, _ = segment.ParseKey(data.Key)
	contentType := data.ContentType
	switch {
	default:
		input.Format = ingestion.FormatGroups
	case data.Format == "trie", contentType == "binary/octet-stream+trie":
		input.Format = ingestion.FormatTrie
	case data.Format == "tree", contentType == "binary/octet-stream+tree":
		input.Format = ingestion.FormatTree
	case data.Format == "lines":
		input.Format = ingestion.FormatLines

	case data.Format == "jfr":
		input.Format = ingestion.FormatJFR
		input.Profile = &jfr.RawProfile{
			FormDataContentType: contentType,
			RawData:             data.RawData,
		}

	case data.Format == "pprof":
		input.Format = ingestion.FormatPprof
		input.Profile = &pprof.RawProfile{
			RawData: data.RawData,
		}

	case data.Format == "speedscope":
		input.Format = ingestion.FormatSpeedscope
		input.Profile = &speedscope.RawProfile{
			RawData: data.RawData,
		}

	case strings.Contains(contentType, "multipart/form-data"):
		input.Profile = &pprof.RawProfile{
			FormDataContentType: contentType,
			RawData:             data.RawData,
			StreamingParser:     true,
			//PoolStreamingParser: true,
		}
	}

	if input.Profile == nil {
		input.Profile = &convertprofile.RawProfile{
			Format:  input.Format,
			RawData: data.RawData,
		}
	}
}

func (p *provider) Init(ctx servicehub.Context) error {
	logger := logrus.StandardLogger()
	e, err := exporter.NewExporter(storageconfig.MetricsExportRules{}, prometheus.DefaultRegisterer)
	if err != nil {
		return err
	}
	profileParser := parser.New(logger, p, e)
	p.profileParser = profileParser
	return nil
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},
		Description: "here is description of erda.oap.collector.processor.profile",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
