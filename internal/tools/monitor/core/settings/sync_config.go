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

package settings

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/monitor/settings/pb"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
)

func (p *provider) syncCreateOrgMonitorConfig() error {
	allOrgs, err := p.Org.ListOrg(apis.WithInternalClientContext(context.Background(), discover.SvcMonitor), &orgpb.ListOrgRequest{
		PageSize: 9999,
		PageNo:   1,
	})
	if err != nil {
		return err
	}
	defCfg := p.settingsService.getDefaultConfig(apis.Language(context.Background()), "")
	for _, org := range allOrgs.List {
		defSetting := &pb.PutSettingsRequest{
			Data: make(map[string]*pb.ConfigGroups),
		}
		for ns, cfg := range defCfg {
			if ns == "general" {
				continue
			}
			for _, monitorType := range p.Cfg.SyncMonitorTypes {
				isEmpty := p.isEmptyConfig(monitorType, ns, org)
				if !isEmpty {
					continue
				}
				monitorTTL := p.getTTL(monitorType, cfg)
				if monitorTTL == nil {
					continue
				}
				if defSetting.Data[ns] == nil {
					defSetting.Data[ns] = &pb.ConfigGroups{
						Groups: []*pb.ConfigGroup{
							{
								Key: "monitor",
								Items: []*pb.ConfigItem{
									{
										Key:   fmt.Sprintf("%s_ttl", monitorType),
										Value: monitorTTL,
									},
								},
							},
						},
					}
				} else {
					defSetting.Data[ns].Groups[0].Items = append(defSetting.Data[ns].Groups[0].Items, &pb.ConfigItem{
						Key:   fmt.Sprintf("%s_ttl", monitorType),
						Value: monitorTTL,
					})
				}
			}
		}
		if len(defSetting.Data) > 0 {
			defSetting.OrgID = int64(org.ID)
			if _, err := p.settingsService.PutSettings(apis.WithInternalClientContext(context.Background(), discover.SvcMonitor), defSetting); err != nil {
				p.Log.Errorf("failed to create default monitor config for org: %s, err: %v", org.Name, err)
			}
		}
	}
	return nil
}

func (p *provider) isEmptyConfig(monitorType string, ns string, org *orgpb.Org) bool {
	var key string
	switch monitorType {
	case "logs":
		tags := map[string]string{
			lib.DiceOrgNameKey: org.Name,
			lib.DiceWorkspace:  ns,
		}
		key = p.LogRetention.GetConfigKey("container", tags)
	case "metrics":
		tags := map[string]string{
			lib.OrgNameKey:    org.Name,
			lib.DiceWorkspace: ns,
		}
		key = p.MetricRetention.GetConfigKey("docker_container_summary", tags)
	default:
		return false
	}
	return len(key) == 0
}

func (p *provider) getTTL(monitorType string, cfg map[string]map[string]*pb.ConfigItem) *structpb.Value {
	if cfg == nil {
		return nil
	}
	monitorCfg := cfg["monitor"]
	if monitorCfg == nil {
		return nil
	}
	ttl := monitorCfg[fmt.Sprintf("%s_ttl", monitorType)]
	return ttl.Value
}

func (p *provider) autoCreateOrgMonitorConfig(orgID uint64) error {
	p.settingsService.PutSettings(context.Background(), &pb.PutSettingsRequest{
		OrgID: int64(orgID),
	})
	return nil
}
