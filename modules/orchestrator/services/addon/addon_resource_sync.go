// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package addon

import (
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
)

func (a *Addon) SyncAddonResources() {
	orgidList := a.getAllOrgIDs()
	for _, orgid := range orgidList {
		addons, err := a.ListAddonInstanceByOrg(orgid)
		if err != nil {
			logrus.Errorf("ListAddonInstanceByOrg: %v", err)
			continue
		}
		for _, addon := range addons {
			pods, err := a.bdl.GetPodInfo(apistructs.PodInfoRequest{
				OrgID:   strconv.FormatUint(orgid, 10),
				AddonID: addon.ID,
			})
			if err != nil {
				logrus.Errorf("failed to getpodinfo(orgid:%d, addonid:%s): %v", orgid, addon.ID, err)
				break
			}
			if !pods.Success {
				logrus.Errorf("failed to getpodinfo(orgid:%d, addonid:%s):%v", orgid, addon.ID, *pods)
			}
			if err := a.updateAddonInstanceResource(addon, pods.Data); err != nil {
				logrus.Errorf("UpdateAddonInstanceResource: %v", err)
				break
			}
		}

	}
}

func (a *Addon) getAllOrgIDs() []uint64 {
	orgs, err := a.bdl.ListDopOrgs(&apistructs.OrgSearchRequest{PageSize: 99999})
	if err != nil {
		return nil
	}
	var orgids []uint64
	for _, org := range orgs.List {
		orgids = append(orgids, org.ID)
	}
	return orgids
}

func (a *Addon) updateAddonInstanceResource(addon dbclient.AddonInstance, pods apistructs.PodInfoDataList) error {
	var cpurequest, cpulimit float64
	var memrequest, memlimit int
	for _, pod := range pods {
		cpurequest += pod.CpuRequest
		cpulimit += pod.CpuLimit
		memrequest += pod.MemRequest
		memlimit += pod.MemLimit
	}
	return a.db.UpdateAddonInstanceResource(addon.ID, cpurequest, cpulimit, memrequest, memlimit)
}
