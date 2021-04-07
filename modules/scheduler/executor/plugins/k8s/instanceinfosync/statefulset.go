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

package instanceinfosync

import (
	"time"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/erda-project/erda/modules/scheduler/instanceinfo"
	"github.com/erda-project/erda/pkg/strutil"
)

func updateAddonStatefulSet(dbclient *instanceinfo.Client, stslist *appsv1.StatefulSetList, delete bool) error {
	r := dbclient.ServiceReader()
	w := dbclient.ServiceWriter()

	for _, sts := range stslist.Items {
		var (
			cluster     string
			orgName     string
			orgID       string
			projectName string
			projectID   string
			workspace   string
			addonName   string

			phase     instanceinfo.ServicePhase
			message   string
			startedAt time.Time

			// 参考 terminus.io/dice/dice/docs/dice-env-vars.md
			envsuffixmap = map[string]*string{
				"DICE_CLUSTER_NAME": &cluster,
				"DICE_ORG_NAME":     &orgName,
				"DICE_ORG_ID":       &orgID,
				"DICE_PROJECT_NAME": &projectName,
				"DICE_PROJECT_ID":   &projectID,
				"DICE_WORKSPACE":    &workspace,
				"DICE_ADDON_NAME":   &addonName,
			}
		)

		// -------------------------------
		// 1. 从 statefulset 中获取所有需要的信息
		// -------------------------------
		for _, env := range sts.Spec.Template.Spec.Containers[0].Env {
			for k, v := range envsuffixmap {
				if strutil.HasSuffixes(env.Name, k) {
					*v = env.Value
				}
			}
		}

		// 如果 envsuffixmap 中的内容有为空的情况, 则不更新这个 sts 到 DB
		// 因为走 dice 部署流程发起的服务都应该有這些环境变量
		skipThisSts := false
		for _, v := range envsuffixmap {
			if *v == "" {
				skipThisSts = true
				break
			}
		}
		if skipThisSts {
			continue
		}
		phase = instanceinfo.ServicePhaseUnHealthy
		if sts.Spec.Replicas == nil || *sts.Spec.Replicas == sts.Status.ReadyReplicas {
			phase = instanceinfo.ServicePhaseHealthy
		}
		startedAt = sts.ObjectMeta.CreationTimestamp.Time

		// -------------------------------
		// 2. 更新或创建 ServiceInfo 记录
		// -------------------------------
		svcs, err := r.ByOrgName(orgName).
			ByProjectName(projectName).
			ByWorkspace(workspace).
			ByServiceType("addon").
			Do()
		if err != nil {
			return err
		}
		serviceinfo := instanceinfo.ServiceInfo{
			Cluster:     cluster,
			OrgName:     orgName,
			OrgID:       orgID,
			ProjectName: projectName,
			ProjectID:   projectID,
			Workspace:   workspace,
			ServiceType: "addon",
			Phase:       phase,
			Message:     message,
			StartedAt:   startedAt,
		}
		switch len(svcs) {
		case 0:
			if delete {
				break
			}
			if err := w.Create(&serviceinfo); err != nil {
				return err
			}
		default:
			for _, svc := range svcs {
				serviceinfo.ID = svc.ID
				if delete {
					if err := w.Delete(svc.ID); err != nil {
						return err
					}
				} else {
					if err := w.Update(serviceinfo); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func updateAddonStatefulSetOnWatch(db *instanceinfo.Client) (func(*appsv1.StatefulSet), func(*appsv1.StatefulSet)) {
	addOrUpdateFunc := func(sts *appsv1.StatefulSet) {
		if err := updateAddonStatefulSet(db,
			&appsv1.StatefulSetList{Items: []appsv1.StatefulSet{*sts}}, false); err != nil {
			logrus.Errorf("failed to update statefulset: %v", err)
		}
	}
	deleteFunc := func(sts *appsv1.StatefulSet) {
		if err := updateAddonStatefulSet(db,
			&appsv1.StatefulSetList{Items: []appsv1.StatefulSet{*sts}}, true); err != nil {
			logrus.Errorf("failed to update(delete) statefulset: %v", err)
		}
	}
	return addOrUpdateFunc, deleteFunc
}
