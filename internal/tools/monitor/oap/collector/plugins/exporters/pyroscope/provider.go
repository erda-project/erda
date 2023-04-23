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

package pyroscope

import (
	"context"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	storageconfig "github.com/pyroscope-io/pyroscope/pkg/config"
	"github.com/pyroscope-io/pyroscope/pkg/convert/jfr"
	"github.com/pyroscope-io/pyroscope/pkg/convert/pprof"
	"github.com/pyroscope-io/pyroscope/pkg/convert/profile"
	"github.com/pyroscope-io/pyroscope/pkg/convert/speedscope"
	"github.com/pyroscope-io/pyroscope/pkg/exporter"
	"github.com/pyroscope-io/pyroscope/pkg/health"
	"github.com/pyroscope-io/pyroscope/pkg/ingestion"
	"github.com/pyroscope-io/pyroscope/pkg/parser"
	"github.com/pyroscope-io/pyroscope/pkg/storage"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/tools/monitor/core/log"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	profileinput "github.com/erda-project/erda/internal/tools/monitor/core/profile"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins"
)

var providerName = plugins.WithPrefixExporter("pyroscope")

type config struct{}

// +provider
type provider struct {
	Cfg           *config
	profileParser *parser.Parser
}

var _ model.Exporter = (*provider)(nil)

func (p *provider) Init(ctx servicehub.Context) error {
	logger := logrus.StandardLogger()
	st, err := storage.New(storage.NewConfig(&storageconfig.Server{
		MaxNodesSerialization: 2048,
	}).WithInMemory(), logger, prometheus.DefaultRegisterer, new(health.Controller), storage.NoopApplicationMetadataService{})
	e, err := exporter.NewExporter(storageconfig.MetricsExportRules{}, prometheus.DefaultRegisterer)
	if err != nil {
		return err
	}
	profileParser := parser.New(logger, st, e)
	p.profileParser = profileParser
	return nil
}

func (p *provider) ComponentClose() error {
	return nil
}

func (p *provider) ExportRaw(items ...*odata.Raw) error { return nil }
func (p *provider) ExportLog(items ...*log.Log) error   { return nil }

func (p *provider) ExportProfile(items ...*profileinput.ProfileIngest) error {
	for _, item := range items {
		ingestInput := &ingestion.IngestInput{
			Format:   item.Format,
			Metadata: item.Metadata,
		}
		p.genRawProfile(ingestInput, item)
		if err := p.profileParser.Ingest(context.Background(), ingestInput); err != nil {
			return err
		}
	}
	return nil
}

func (p *provider) ExportMetric(items ...*metric.Metric) error {
	return nil
}

func (p *provider) ExportSpan(items ...*trace.Span) error {
	return nil
}

func (p *provider) ComponentConfig() interface{} {
	return p.Cfg
}

func (p *provider) Connect() error {
	return nil
}

func (p *provider) genRawProfile(input *ingestion.IngestInput, data *profileinput.ProfileIngest) {
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
			PoolStreamingParser: true,
		}
	}

	if input.Profile == nil {
		input.Profile = &profile.RawProfile{
			Format:  input.Format,
			RawData: data.RawData,
		}
	}
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},
		Description:  "here is description of erda.oap.collector.exporter.pyroscope",
		Dependencies: []string{"clickhouse", "clickhouse.table.initializer"},
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
