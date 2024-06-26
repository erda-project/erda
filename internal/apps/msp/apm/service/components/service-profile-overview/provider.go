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

package service_profile_overview

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pyroscope-io/pyroscope/pkg/model/appmetadata"
	"github.com/pyroscope-io/pyroscope/pkg/service"
	"github.com/pyroscope-io/pyroscope/pkg/structs/flamebearer"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/topn"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/topn/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/i18n"
	mysql "github.com/erda-project/erda-infra/providers/mysql/v2"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

const (
	MemTop5 string = "peripheralInterfaceChart"
	CPUTop5 string = "pathClientRpsMaxTop5"
	Span    string = "24"
)

const (
	languageJava string = "java"
	languageGo   string = "go"
)

const (
	cpuUsageTop5I18n = "cpuUsageTop5"
	memUsageTop5I18n = "memUsageTop5"
)

var (
	languageCpuAlias = map[string]string{
		languageJava: "itimer",
		languageGo:   "cpu",
	}
	languageMemAlias = map[string]string{
		languageJava: "alloc_in_new_tlab_bytes",
		languageGo:   "alloc_space",
	}
)

type provider struct {
	impl.DefaultTop
	base.DefaultProvider
	Log                logs.Logger
	I18n               i18n.Translator `autowired:"i18n" translator:"msp-i18n"`
	DB                 mysql.Interface
	applicationService service.ApplicationMetadataService
	bdl                *bundle.Bundle
}

func init() {
	servicehub.Register("service-profile-overview", &servicehub.Spec{
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
	cpregister.RegisterProviderComponent("service-profile-overview", "service-profile-overview", &provider{})
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithMonitor())
	p.applicationService = service.NewApplicationMetadataService(p.DB.DB())
	return nil
}

// RegisterRenderingOp .
func (p *provider) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return p.RegisterInitializeOp()
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return p
}

// RegisterInitializeOp .
func (p *provider) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) cptype.IStdStructuredPtr {
		lang := sdk.Lang
		startTime := int64(p.StdInParamsPtr.Get("startTime").(float64))
		endTime := int64(p.StdInParamsPtr.Get("endTime").(float64))
		serviceID := p.StdInParamsPtr.Get("serviceId").(string)
		workspace := p.StdInParamsPtr.Get("workspace").(string)
		projectID := p.StdInParamsPtr.Get("projectID").(string)

		query := url.Values{}
		query.Set("projectID", projectID)
		query.Set("workspace", workspace)
		profileApps, err := p.applicationService.List(context.WithValue(context.Background(), "query", query))
		if err != nil {
			p.Log.Errorf("Failed to query apps, err: %v", err)
			return nil
		}

		cpuAlias := languageCpuAlias[judgeAppLanguage(serviceID, profileApps)]
		if cpuAlias == "" {
			p.Log.Errorf("Failed to get service: %s cpu alias", serviceID)
			return nil
		}
		memAlias := languageMemAlias[judgeAppLanguage(serviceID, profileApps)]
		if memAlias == "" {
			p.Log.Errorf("Failed to get service: %s mem alias", serviceID)
			return nil
		}
		switch sdk.Comp.Name {
		case MemTop5:
			var records []topn.Record
			pathRpsMaxTop5, err := p.getTop5(serviceID, projectID, workspace, memAlias, startTime, endTime)
			if err != nil {
				p.Log.Error(err)
			}
			pathRpsMaxTop5Records := topn.Record{Title: p.I18n.Text(lang, memUsageTop5I18n), Span: Span}
			pathRpsMaxTop5Records.Items = pathRpsMaxTop5
			records = append(records, pathRpsMaxTop5Records)
			return &impl.StdStructuredPtr{StdDataPtr: &topn.Data{List: records}}
		case CPUTop5:
			var records []topn.Record
			pathRpsMaxTop5, err := p.getTop5(serviceID, projectID, workspace, cpuAlias, startTime, endTime)
			if err != nil {
				p.Log.Error(err)
			}
			pathRpsMaxTop5Records := topn.Record{Title: p.I18n.Text(lang, cpuUsageTop5I18n), Span: Span}
			pathRpsMaxTop5Records.Items = pathRpsMaxTop5
			records = append(records, pathRpsMaxTop5Records)
			return &impl.StdStructuredPtr{StdDataPtr: &topn.Data{List: records}}
		}
		return &impl.StdStructuredPtr{StdDataPtr: &topn.Data{}}
	}
}

func (p *provider) getTop5(serviceID string, projectID, workspace, spyAlias string, startTime, endTime int64) ([]topn.Item, error) {
	segmentKey := fmt.Sprintf("%s.%s{DICE_WORKSPACE=\"%s\",DICE_PROJECT_ID=\"%s\"}", serviceID, spyAlias, workspace, projectID)
	topRender, err := p.bdl.ProfileRender(&apistructs.ProfileRenderRequest{
		Query:    segmentKey,
		From:     fmt.Sprintf("%d", startTime),
		Until:    fmt.Sprintf("%d", endTime),
		MaxNodes: 8192,
	})
	if err != nil {
		return []topn.Item{}, err
	}
	if topRender.Flamebearer.NumTicks <= 0 {
		return []topn.Item{}, nil
	}
	formattedData := flamebearer.GenerateCellTable(topRender.Flamebearer, int(topRender.Metadata.SampleRate), topRender.Metadata.Units.String())
	var items []topn.Item
	var count int
	for _, node := range formattedData {
		if count >= 5 {
			break
		}
		value, unit := getValue(node.FormatTotal)
		items = append(items, topn.Item{
			Name:  node.Name,
			Value: value,
			Unit:  unit,
		})
		count++
	}
	return items, nil
}

func getValue(formatTotal string) (float64, string) {
	dat := strings.Split(formatTotal, " ")
	if len(dat) != 2 {
		return 0, ""
	}
	value, err := strconv.ParseFloat(dat[0], 64)
	if err != nil {
		return 0, ""
	}
	return value, dat[1]
}

func judgeAppLanguage(serviceName string, apps []appmetadata.ApplicationMetadata) string {
	for _, app := range apps {
		if app.ServiceName == serviceName {
			switch app.SpyName {
			case "javaspy":
				return languageJava
			case "gospy":
				return languageGo
			default:
				return ""
			}
		}
	}
	return ""
}
