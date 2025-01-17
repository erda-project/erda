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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda-proto-go/core/monitor/settings/pb"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/tools/monitor/common/db"
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
	client := db.New(p.DB)
	defaultConfig := p.settingsService.getDefaultConfig(apis.Language(context.Background()), "")
	for _, org := range allOrgs.List {
		var registers []db.SpMonitorConfigRegister
		monitorRegisters, err := client.MonitorConfigRegister.ListRegisterByOrgId(strconv.FormatUint(org.ID, 10))
		if err != nil {
			p.Log.Errorf("failed to get monitor config register by orgId: %d, err: %v", org.ID, err)
			return err
		}

		logRegisters, err := client.MonitorConfigRegister.ListRegisterByType("log")
		if err != nil {
			p.Log.Errorf("failed to get monitor config register by type: log, err: %v", err)
			return err
		}
		registers = make([]db.SpMonitorConfigRegister, 0, len(logRegisters)+len(monitorRegisters))
		registers = append(registers, monitorRegisters...)
		registers = append(registers, logRegisters...)

		for _, register := range registers {
			if !p.isEmptyConfig(&register, org) {
				continue
			}

			nsConfig := defaultConfig[register.Namespace]
			defConfig := nsConfig["monitor"]

			req := &pb.PutSettingsWithTypeRequest{
				OrgID: int64(org.ID),
				Data: &pb.ConfigGroup{
					Key:   "monitor",
					Items: make([]*pb.ConfigItem, 0),
				},
				Namespace:   register.Namespace,
				MonitorType: register.Type,
			}

			var ttlItem *pb.ConfigItem
			var hotTTLItem *pb.ConfigItem

			switch register.Type {
			case "log":
				ttlItem = defConfig[LogsTTLKey]
				hotTTLItem = defConfig[LogsHotTTLKey]
			case "metric":
				ttlItem = defConfig[MetricsTTLKey]
				hotTTLItem = defConfig[MetricsHotTTLKey]
			}

			if ttlItem == nil {
				err = fmt.Errorf("ttl item is nil, monitor type: %s", register.Type)
				p.Log.Error(err)
				return err
			}
			if hotTTLItem == nil {
				err = fmt.Errorf("hot_ttl item is nil, monitor type: %s", register.Type)
				p.Log.Error(err)
				return err
			}

			req.Data.Items = append(req.Data.Items, ttlItem)
			req.Data.Items = append(req.Data.Items, hotTTLItem)

			if _, err := p.settingsService.PutSettingsWithType(context.Background(), req); err != nil {
				p.Log.Errorf("failed to put settings with monitor type: %s, err: %v", register.Type, err)
				return err
			}
		}
	}

	return nil
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (p *provider) isEmptyConfig(register *db.SpMonitorConfigRegister, org *orgpb.Org) bool {
	var kvs []KeyValue
	if err := json.Unmarshal([]byte(register.Filters), &kvs); err != nil {
		p.Log.Errorf("failed to unmarshal filters, filters: %s, err: %v", register.Filters, err)
		return false
	}
	filters := make(map[string]string)
	for _, kv := range kvs {
		filters[kv.Key] = kv.Value
	}

	names := strings.Split(register.Names, ",")

	switch register.Type {
	case "log":
		orgName, err := p.settingsService.getOrgName(int64(org.ID))
		if err != nil {
			p.Log.Errorf("failed to get org name, err: %v", err)
			return false
		}
		filters["dice_org_name"] = orgName
		for _, name := range names {
			return len(p.LogRetention.GetConfigKey(name, filters)) == 0
		}

	case "metric":
		for _, name := range names {
			return len(p.MetricRetention.GetConfigKey(name, filters)) == 0
		}
	default:
		return false
	}
	return false
}
