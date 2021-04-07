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
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/modules/scheduler/instanceinfo"
)

// updateStatelessServiceDeploymentService 更新 stateless-service类型的deployment 到db中
// TODO: 还没有处理 'cluster', 'namespace', 'name'  这 3 个字段
func updateStatelessServiceDeployment(dbclient *instanceinfo.Client, deploylist *appsv1.DeploymentList, delete bool) error {
	r := dbclient.ServiceReader()
	w := dbclient.ServiceWriter()

	for _, deploy := range deploylist.Items {
		var (
			cluster         string
			orgName         string
			orgID           string
			projectName     string
			projectID       string
			applicationName string
			applicationID   string
			runtimeName     string
			runtimeID       string
			serviceName     string
			workspace       string
			phase           instanceinfo.ServicePhase
			message         string
			startedAt       time.Time
			namespace       string
			name            string

			// 参考 terminus.io/dice/dice/docs/dice-env-vars.md
			envmap = map[string]*string{
				"DICE_CLUSTER_NAME":     &cluster,
				"DICE_ORG_NAME":         &orgName,
				"DICE_ORG_ID":           &orgID,
				"DICE_PROJECT_NAME":     &projectName,
				"DICE_PROJECT_ID":       &projectID,
				"DICE_APPLICATION_NAME": &applicationName,
				"DICE_APPLICATION_ID":   &applicationID,
				"DICE_RUNTIME_NAME":     &runtimeName,
				"DICE_RUNTIME_ID":       &runtimeID,
				"DICE_SERVICE_NAME":     &serviceName,
				"DICE_WORKSPACE":        &workspace,
			}
		)
		// -------------------------------------
		// 1. 从 deployment 获取所有需要的信息
		// -------------------------------------
		for _, env := range deploy.Spec.Template.Spec.Containers[0].Env {
			if variable, ok := envmap[env.Name]; ok {
				*variable = env.Value
			}
		}
		// 如果 envmap 中的内容有为空的情况, 则不更新这个 deployment 到 DB
		// 因为走 dice 部署流程发起的服务都应该有這些环境变量
		skipThisDeploy := false
		for _, v := range envmap {
			if *v == "" {
				skipThisDeploy = true
				break
			}
		}
		if skipThisDeploy {
			continue
		}
		phase = instanceinfo.ServicePhaseUnHealthy
		for _, cond := range deploy.Status.Conditions {
			if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionFalse {
				phase = instanceinfo.ServicePhaseHealthy
				break
			}
		}

		startedAt = deploy.ObjectMeta.CreationTimestamp.Time

		if runtimeID != "" && workspace != "" {
			namespace = "services"
			name = workspace + "-" + runtimeID
		}

		// -------------------------------
		// 2. 更新或创建 ServiceInfo 记录
		// -------------------------------

		svcs, err := r.ByOrgName(orgName).
			ByProjectName(projectName).
			ByApplicationName(applicationName).
			ByRuntimeName(runtimeName).
			ByService(serviceName).
			ByWorkspace(workspace).
			ByServiceType("stateless-service").
			Do()
		if err != nil {
			return err
		}
		serviceinfo := instanceinfo.ServiceInfo{
			Cluster:         cluster,
			Namespace:       namespace,
			Name:            name,
			OrgName:         orgName,
			OrgID:           orgID,
			ProjectName:     projectName,
			ProjectID:       projectID,
			ApplicationName: applicationName,
			ApplicationID:   applicationID,
			RuntimeName:     runtimeName,
			RuntimeID:       runtimeID,
			ServiceName:     serviceName,
			Workspace:       workspace,
			ServiceType:     "stateless-service",
			Phase:           phase,
			Message:         message,
			StartedAt:       startedAt,
		}
		switch len(svcs) {
		case 0: // DB 中没有这个 service
			if delete {
				break
			}
			if err := w.Create(&serviceinfo); err != nil {
				return err
			}
		default: // 会有 len(svcs)> 1 的情况吗?
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

// updateAddonDeploymentService 更新 addon 类型的 deployment到 db中
func updateAddonDeployment(dbclient *instanceinfo.Client, deploylist *appsv1.DeploymentList, delete bool) error {
	r := dbclient.ServiceReader()
	w := dbclient.ServiceWriter()

	for _, deploy := range deploylist.Items {
		var (
			cluster     string
			orgName     string
			orgID       string
			projectName string
			projectID   string
			workspace   string
			addonName   string
			phase       instanceinfo.ServicePhase
			message     string
			startedAt   time.Time

			envmap = map[string]*string{
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
		// 1. 从 deployment 中获取所有需要的信息
		// -------------------------------
		for _, env := range deploy.Spec.Template.Spec.Containers[0].Env {
			if variable, ok := envmap[env.Name]; ok {
				*variable = env.Value
			}
		}

		skipThisDeploy := false
		for _, v := range envmap {
			if *v == "" {
				skipThisDeploy = true
				break
			}
		}
		if skipThisDeploy {
			continue
		}
		phase = instanceinfo.ServicePhaseUnHealthy
		for _, cond := range deploy.Status.Conditions {
			if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionFalse {
				phase = instanceinfo.ServicePhaseHealthy
				break
			}
		}
		startedAt = deploy.ObjectMeta.CreationTimestamp.Time

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

func updateDeploymentOnWatch(db *instanceinfo.Client) (func(*appsv1.Deployment), func(*appsv1.Deployment)) {
	addOrUpdateFunc := func(deploy *appsv1.Deployment) {
		if err := updateStatelessServiceDeployment(db,
			&appsv1.DeploymentList{Items: []appsv1.Deployment{*deploy}}, false); err != nil {
			logrus.Errorf("failed to update deployment: %v", err)
		}
		if err := updateAddonDeployment(db,
			&appsv1.DeploymentList{Items: []appsv1.Deployment{*deploy}}, false); err != nil {
			logrus.Errorf("failed to update deployment: %v", err)
		}
	}
	deleteFunc := func(deploy *appsv1.Deployment) {
		if err := updateStatelessServiceDeployment(db,
			&appsv1.DeploymentList{Items: []appsv1.Deployment{*deploy}}, true); err != nil {
			logrus.Errorf("failed to update(delete) deployment: %v", err)
		}
		if err := updateAddonDeployment(db,
			&appsv1.DeploymentList{Items: []appsv1.Deployment{*deploy}}, true); err != nil {
			logrus.Errorf("failed to update(delete) deployment: %v", err)
		}
	}

	return addOrUpdateFunc, deleteFunc
}
