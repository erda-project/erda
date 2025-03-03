package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (r *Service) getRuntimeScaleRecordByRuntimeIds(ids []uint64) ([]dbclient.Runtime, []apistructs.RuntimeScaleRecord, error) {
	rsrs := make([]apistructs.RuntimeScaleRecord, 0)

	runtimes, err := r.db.FindRuntimesByIds(ids)
	if err != nil {
		logrus.Errorf("find runtimes by ids in ps_v2_project_runtimes failed, err: %v", err)
		return runtimes, rsrs, err
	}
	if len(runtimes) == 0 {
		logrus.Errorf("[batch redeploy] find runtimes by ids in ps_v2_project_runtimes failed, err: %v", err)
		return runtimes, rsrs, errors.Errorf("no runtimes found by runtime ids %#v", ids)
	}

	// 确保所有的 ids 都有对应的 Runtime 记录
	notExistedIds := make([]uint64, 0)
	existedIds := make(map[uint64]bool)
	for _, rt := range runtimes {
		existedIds[rt.ID] = true
	}
	for _, id := range ids {
		if _, existed := existedIds[id]; !existed {
			notExistedIds = append(notExistedIds, id)
		}
	}

	if len(notExistedIds) > 0 {
		logrus.Errorf("[batch redeploy] not found runtimes for runtime id:  %#v", notExistedIds)
		return runtimes, rsrs, errors.Errorf("no runtimes found for runtime id: %#v", notExistedIds)
	}

	for _, rt := range runtimes {
		runtimescaleRecord := apistructs.RuntimeScaleRecord{
			ApplicationId: rt.ApplicationID,
			Workspace:     rt.Workspace,
			Name:          rt.Name,
			RuntimeID:     rt.ID,
			PayLoad: apistructs.PreDiceDTO{
				Services: make(map[string]*apistructs.RuntimeInspectServiceDTO),
			},
		}

		uniqueId := spec.RuntimeUniqueId{
			ApplicationId: rt.ApplicationID,
			Workspace:     rt.Workspace,
			Name:          rt.Name,
		}

		pre, err := r.db.FindPreDeployment(uniqueId)
		if err != nil {
			logrus.Errorf("find PreDeployment record in table ps_v2_pre_builds failed for runtime %#v, err: %v", uniqueId, err)
			return runtimes, rsrs, errors.Errorf("find PreDeployment record in table ps_v2_pre_builds failed for runtime %#v, err: %v", uniqueId, err)
		}

		oldOverlay, err := getPreDeploymentDiceOverlay(pre)
		if err != nil {
			logrus.Errorf("getPreDeploymentDiceOverlay failed get r diceyml.Object in table ps_v2_pre_builds for runtime %#v, err: %v", uniqueId, err)
			return runtimes, rsrs, errors.Errorf("no runtimes found for runtime id: %#v", notExistedIds)
		}
		for k, svc := range oldOverlay.Services {
			runtimescaleRecord.PayLoad.Services[k] = &apistructs.RuntimeInspectServiceDTO{
				Deployments: apistructs.RuntimeServiceDeploymentsDTO{
					Replicas: svc.Deployments.Replicas,
				},
				Resources: apistructs.RuntimeServiceResourceDTO{
					CPU: svc.Resources.CPU,
					Mem: svc.Resources.Mem,
				},
			}
		}
		rsrs = append(rsrs, runtimescaleRecord)
	}
	return runtimes, rsrs, nil
}

func getPreDeploymentDiceOverlay(pre *dbclient.PreDeployment) (diceyml.Object, error) {
	var oldOverlay diceyml.Object
	var diceObj diceyml.Object
	// 没有 pre.DiceOverlay 信息表示部署之后还未进行过 scale 操作，因此如果当前这次 scale 操作是 scaleDown, 则按 pre.Dice 中的副本数设置 恢复时的副本数
	if err := json.Unmarshal([]byte(pre.Dice), &diceObj); err != nil {
		logrus.Errorf("Unmarshal PreDeployment record dice to diceyml.Object for runtime [application_id: %v workspace: %v name: %v ] failed, err: %v", pre.ApplicationId, pre.Workspace, pre.RuntimeName, err)
		return oldOverlay, errors.Errorf("Unmarshal PreDeployment record dice to diceyml.Object for runtime [application_id: %v workspace: %v name: %v ] failed, err: %v", pre.ApplicationId, pre.Workspace, pre.RuntimeName, err)
	}
	oldOverlay.Services = make(map[string]*diceyml.Service)
	for k, v := range diceObj.Services {
		oldOverlay.Services[k] = &diceyml.Service{
			Deployments: diceyml.Deployments{
				Replicas: v.Deployments.Replicas,
			},
			Resources: diceyml.Resources{
				CPU: v.Resources.CPU,
				Mem: v.Resources.Mem,
			},
		}
	}

	if pre.DiceOverlay != "" {
		var currentDiceObj diceyml.Object
		if err := json.Unmarshal([]byte(pre.DiceOverlay), &oldOverlay); err != nil {
			logrus.Errorf("Unmarshal PreDeployment record dice_overaly to diceyml.Object for runtime [application_id: %v workspace: %v name: %v ] failed, err: %v", pre.ApplicationId, pre.Workspace, pre.RuntimeName, err)
			return oldOverlay, errors.Errorf("Unmarshal PreDeployment record dice_overaly to diceyml.Object for runtime [application_id: %v workspace: %v name: %v ] failed, err: %v", pre.ApplicationId, pre.Workspace, pre.RuntimeName, err)
		}
		for k, v := range currentDiceObj.Services {
			if _, ok := oldOverlay.Services[k]; ok {
				if oldOverlay.Services[k].Deployments.Replicas != v.Deployments.Replicas {
					oldOverlay.Services[k].Deployments.Replicas = v.Deployments.Replicas
				}
				if oldOverlay.Services[k].Resources.CPU != v.Resources.CPU {
					oldOverlay.Services[k].Resources.CPU = v.Resources.CPU
				}
				if oldOverlay.Services[k].Resources.Mem != v.Resources.Mem {
					oldOverlay.Services[k].Resources.Mem = v.Resources.Mem
				}
			}
		}
	}
	return oldOverlay, nil
}

// batchRuntimeReDeploy 批量重新部署
func (r *Service) batchRuntimeReDeploy(ctx context.Context, userID user.ID, runtimes []dbclient.Runtime, runtimeScaleRecords apistructs.RuntimeScaleRecords) apistructs.BatchRuntimeReDeployResults {
	batchRuntimeReDeployResult := apistructs.BatchRuntimeReDeployResults{
		Total:           len(runtimeScaleRecords.Runtimes),
		Success:         0,
		Failed:          0,
		ReDeployed:      make([]apistructs.RuntimeDeployDTO, 0),
		ReDeployedIds:   make([]uint64, 0),
		UnReDeployed:    make([]apistructs.RuntimeDTO, 0),
		UnReDeployedIds: make([]uint64, 0),
		ErrMsg:          make([]string, 0),
	}

	// 根据请求 RuntimeScaleRecords.IDs  执行重新部署
	if len(runtimeScaleRecords.IDs) > 0 {
		for _, runtime := range runtimes {
			redep, err := r.RedeployPipeline(ctx, userID, runtime.OrgID, runtime.ID)
			if err != nil {
				logrus.Errorf("[batch redeploy] redeploy failed for runtime %s for runtime instance: %#v, error: %v", runtime.Name, runtime, err)
				errMsg := fmt.Sprintf("redeploy redeploy failed for runtime %s for runtime instance: %#v, error: %v", runtime.Name, runtime, err)
				batchRuntimeReDeployResult.ErrMsg = append(batchRuntimeReDeployResult.ErrMsg, errMsg)
				batchRuntimeReDeployResult.Failed++
				batchRuntimeReDeployResult.UnReDeployedIds = append(batchRuntimeReDeployResult.UnReDeployedIds, runtime.ID)
				curr := apistructs.RuntimeDTO{
					Name:          runtime.Name,
					Workspace:     runtime.Workspace,
					ApplicationID: runtime.ApplicationID,
				}
				batchRuntimeReDeployResult.UnReDeployed = append(batchRuntimeReDeployResult.UnReDeployed, curr)
				continue
			}

			logrus.Infof("[batch redeploy] redeploy successed for rruntime %s for runtime instance: %#v", runtime.Name, runtime)
			batchRuntimeReDeployResult.Success++
			batchRuntimeReDeployResult.ReDeployedIds = append(batchRuntimeReDeployResult.ReDeployedIds, runtime.ID)
			batchRuntimeReDeployResult.ReDeployed = append(batchRuntimeReDeployResult.ReDeployed, *redep)
		}
	} else {
		// 根据请求 RuntimeScaleRecords.Runtimes  执行重新部署
		for index := range runtimeScaleRecords.Runtimes {
			uniqueId := spec.RuntimeUniqueId{
				ApplicationId: runtimeScaleRecords.Runtimes[index].ApplicationId,
				Workspace:     runtimeScaleRecords.Runtimes[index].Workspace,
				Name:          runtimeScaleRecords.Runtimes[index].Name,
			}

			curr := apistructs.RuntimeDTO{
				Name:          uniqueId.Name,
				Workspace:     uniqueId.Workspace,
				ApplicationID: uniqueId.ApplicationId,
			}

			runtime, err := r.db.FindRuntime(uniqueId)
			if err != nil {
				logrus.Errorf("[batch redeploy] find runtime record in ps_v2_project_runtimes for runtime %#v failed, err: %v", uniqueId, err)
				errMsg := fmt.Sprintf("find runtime record in ps_v2_project_runtimes for runtime %#v failed, err: %v", uniqueId, err)
				batchRuntimeReDeployResult.ErrMsg = append(batchRuntimeReDeployResult.ErrMsg, errMsg)
				batchRuntimeReDeployResult.Failed++
				batchRuntimeReDeployResult.UnReDeployedIds = append(batchRuntimeReDeployResult.UnReDeployedIds, runtime.ID)
				batchRuntimeReDeployResult.UnReDeployed = append(batchRuntimeReDeployResult.UnReDeployed, curr)
				continue
			}
			if runtime == nil {
				logrus.Errorf("[batch redeploy] runtime %s is not existed for runtime %#v, can not redeploy", uniqueId.Name, uniqueId)
				errMsg := fmt.Sprintf("runtime %s is not existed for runtime %#v, can not redeploy", uniqueId.Name, uniqueId)
				batchRuntimeReDeployResult.ErrMsg = append(batchRuntimeReDeployResult.ErrMsg, errMsg)
				batchRuntimeReDeployResult.Failed++
				batchRuntimeReDeployResult.UnReDeployedIds = append(batchRuntimeReDeployResult.UnReDeployedIds, runtime.ID)
				batchRuntimeReDeployResult.UnReDeployed = append(batchRuntimeReDeployResult.UnReDeployed, curr)
				continue
			}

			redep, err := r.RedeployPipeline(ctx, userID, runtime.OrgID, runtime.ID)
			if err != nil {
				logrus.Errorf("[batch redeploy] redeploy failed for runtime %s for %#v", uniqueId.Name, uniqueId)
				errMsg := fmt.Sprintf("redeploy failed for runtime %s for %#v", uniqueId.Name, uniqueId)
				batchRuntimeReDeployResult.ErrMsg = append(batchRuntimeReDeployResult.ErrMsg, errMsg)
				batchRuntimeReDeployResult.Failed++
				batchRuntimeReDeployResult.UnReDeployedIds = append(batchRuntimeReDeployResult.UnReDeployedIds, runtime.ID)
				batchRuntimeReDeployResult.UnReDeployed = append(batchRuntimeReDeployResult.UnReDeployed, curr)
				continue
			}

			logrus.Infof("[batch redeploy] redeploy successed for runtime %s for %#v", uniqueId.Name, uniqueId)
			batchRuntimeReDeployResult.Success++
			batchRuntimeReDeployResult.ReDeployedIds = append(batchRuntimeReDeployResult.ReDeployedIds, runtime.ID)
			batchRuntimeReDeployResult.ReDeployed = append(batchRuntimeReDeployResult.ReDeployed, *redep)
		}
	}
	return batchRuntimeReDeployResult
}

// batchRuntimeRecovery 批量恢复（副本数 0 ----> N）
func (r *Service) batchRuntimeRecovery(runtimeScaleRecords *apistructs.RuntimeScaleRecords) {
	// len(runtimeScaleRecords.IDs) > 0, 则已经在 getRuntimeScaleRecordByRuntimeIds 逻辑中构建了 runtimeScaleRecords.Runtimes，此处无需重新构建

	// 根据请求 RuntimeScaleRecords.IDs  执行重新部署
	if len(runtimeScaleRecords.IDs) == 0 {
		for index := range runtimeScaleRecords.Runtimes {
			uniqueId := spec.RuntimeUniqueId{
				ApplicationId: runtimeScaleRecords.Runtimes[index].ApplicationId,
				Workspace:     runtimeScaleRecords.Runtimes[index].Workspace,
				Name:          runtimeScaleRecords.Runtimes[index].Name,
			}
			runtime, err := r.db.FindRuntime(uniqueId)
			if err != nil {
				logrus.Errorf("[batch delete] find runtime record in ps_v2_project_runtimes for runtime %#v failed, err: %v", uniqueId, err)
				continue
			}
			if runtime != nil {
				runtimeScaleRecords.Runtimes[index].RuntimeID = runtime.ID
			}

			pre, err := r.db.FindPreDeployment(uniqueId)
			if err != nil {
				logrus.Errorf("[batch scale up] find PreDeployment record in table ps_v2_pre_builds failed for runtime %#v, err: %v", uniqueId, err)
				continue
			}
			var oldOverlay diceyml.Object
			if pre.DiceOverlay != "" {
				if err = json.Unmarshal([]byte(pre.DiceOverlay), &oldOverlay); err != nil {
					logrus.Errorf("[batch scale up] Unmarshal PreDeployment record dice_overaly to diceyml.Object for runtime %#v failed, err: %v", uniqueId, err)
					continue
				}
			} else {
				var diceObj diceyml.Object
				if err = json.Unmarshal([]byte(pre.Dice), &diceObj); err != nil {
					logrus.Errorf("[batch scale up] Unmarshal PreDeployment record dice to diceyml.Object for runtime %#v failed, err: %v", uniqueId, err)
					continue
				}
				oldOverlay.Services = make(map[string]*diceyml.Service)
				for k, v := range diceObj.Services {
					oldOverlay.Services[k] = &diceyml.Service{
						Deployments: diceyml.Deployments{
							Replicas: v.Deployments.Replicas,
						},
						Resources: diceyml.Resources{
							CPU: v.Resources.CPU,
							Mem: v.Resources.Mem,
						},
					}
				}
			}

			for k, v := range runtimeScaleRecords.Runtimes[index].PayLoad.Services {
				if _, ok := oldOverlay.Services[k]; ok {
					v.Deployments.Replicas = oldOverlay.Services[k].Deployments.Replicas
				}
			}
		}
	}
}

// batchRuntimeScaleDownAddRuntimeIDs 为 RuntimeScaleRecord 对象加入 runtimeID
func (r *Service) batchRuntimeScaleAddRuntimeIDs(runtimeScaleRecords *apistructs.RuntimeScaleRecords, action string) {
	for index := range runtimeScaleRecords.Runtimes {

		if action == apistructs.ScaleActionDown {
			for _, v := range runtimeScaleRecords.Runtimes[index].PayLoad.Services {
				v.Deployments.Replicas = 0
			}
		}

		if runtimeScaleRecords.Runtimes[index].RuntimeID == 0 {
			uniqueId := spec.RuntimeUniqueId{
				ApplicationId: runtimeScaleRecords.Runtimes[index].ApplicationId,
				Workspace:     runtimeScaleRecords.Runtimes[index].Workspace,
				Name:          runtimeScaleRecords.Runtimes[index].Name,
			}

			runtime, err := r.db.FindRuntime(uniqueId)
			logrus.Errorf("[batch redeploy] find runtime record in ps_v2_project_runtimes for runtime %#v, runtime %#v, error: %v", uniqueId, runtime, err)
			if err != nil {
				logrus.Errorf("[batch redeploy] find runtime record in ps_v2_project_runtimes for runtime %#v failed, err: %v", uniqueId, err)
				continue
			}
			if runtime != nil {
				runtimeScaleRecords.Runtimes[index].RuntimeID = runtime.ID
			}
		}
	}
}

func (r *Service) batchRuntimeDelete(userID user.ID, runtimes []dbclient.Runtime, runtimeScaleRecords apistructs.RuntimeScaleRecords) apistructs.BatchRuntimeDeleteResults {
	batchRuntimeDeleteResult := apistructs.BatchRuntimeDeleteResults{
		Total:        len(runtimeScaleRecords.Runtimes),
		Success:      0,
		Failed:       0,
		Deleted:      make([]apistructs.RuntimeDTO, 0),
		DeletedIds:   make([]uint64, 0),
		UnDeleted:    make([]apistructs.RuntimeDTO, 0),
		UnDeletedIds: make([]uint64, 0),
		ErrMsg:       make([]string, 0),
	}

	// 根据请求 RuntimeScaleRecords.IDs  执行删除
	if len(runtimeScaleRecords.IDs) > 0 {
		for _, runtime := range runtimes {
			logrus.Debugf("[batch delete] deleting runtime %d, operator %s", runtime.ID, userID)
			runtimedto, err := r.DeleteRuntime(userID, runtime.OrgID, runtime.ID)
			if err != nil {
				logrus.Errorf("[batch delete] delete runtime Id %v  runtime %#v failed, error: %v", runtime.ID, runtime, err)
				errMsg := fmt.Sprintf("delete runtime Id %v  runtime %#v failed, error: %v", runtime.ID, runtime, err)
				batchRuntimeDeleteResult.Failed++
				batchRuntimeDeleteResult.ErrMsg = append(batchRuntimeDeleteResult.ErrMsg, errMsg)
				batchRuntimeDeleteResult.UnDeletedIds = append(batchRuntimeDeleteResult.UnDeletedIds, runtime.ID)
				continue
			}

			batchRuntimeDeleteResult.Success++
			batchRuntimeDeleteResult.DeletedIds = append(batchRuntimeDeleteResult.DeletedIds, runtime.ID)
			batchRuntimeDeleteResult.Deleted = append(batchRuntimeDeleteResult.Deleted, *runtimedto)
		}
	} else {
		// 根据请求 RuntimeScaleRecords.Runtimes  执行删除
		for index := range runtimeScaleRecords.Runtimes {
			uniqueId := spec.RuntimeUniqueId{
				ApplicationId: runtimeScaleRecords.Runtimes[index].ApplicationId,
				Workspace:     runtimeScaleRecords.Runtimes[index].Workspace,
				Name:          runtimeScaleRecords.Runtimes[index].Name,
			}

			curr := apistructs.RuntimeDTO{
				Name:          uniqueId.Name,
				Workspace:     uniqueId.Workspace,
				ApplicationID: uniqueId.ApplicationId,
			}

			runtime, err := r.db.FindRuntime(uniqueId)
			if err != nil {
				logrus.Errorf("[batch delete] find runtime record in ps_v2_project_runtimes for runtime %#v failed, err: %v", uniqueId, err)
				errMsg := fmt.Sprintf("find runtime record in ps_v2_project_runtimes for runtime %#v failed, err: %v", uniqueId, err)
				batchRuntimeDeleteResult.Failed++
				batchRuntimeDeleteResult.ErrMsg = append(batchRuntimeDeleteResult.ErrMsg, errMsg)
				batchRuntimeDeleteResult.UnDeletedIds = append(batchRuntimeDeleteResult.UnDeletedIds, runtime.ID)
				batchRuntimeDeleteResult.UnDeleted = append(batchRuntimeDeleteResult.UnDeleted, curr)
				continue
			}
			if runtime == nil {
				logrus.Errorf("[batch delete] runtime %s is not existed for runtime %#v", uniqueId.Name, uniqueId)
				errMsg := fmt.Sprintf("no record in ps_v2_project_runtimes for runtime %#v failed, err: %v", uniqueId, err)
				batchRuntimeDeleteResult.Failed++
				batchRuntimeDeleteResult.ErrMsg = append(batchRuntimeDeleteResult.ErrMsg, errMsg)
				batchRuntimeDeleteResult.UnDeletedIds = append(batchRuntimeDeleteResult.UnDeletedIds, runtime.ID)
				batchRuntimeDeleteResult.UnDeleted = append(batchRuntimeDeleteResult.UnDeleted, curr)
				continue
			}

			logrus.Debugf("[batch delete] deleting runtime %d, operator %s", runtime.ID, userID)
			runtimedto, err := r.DeleteRuntime(userID, runtime.OrgID, runtime.ID)
			if err != nil {
				logrus.Errorf("[batch delete] runtime %s is delete failed %#v, error: %v", uniqueId.Name, uniqueId, err)
				errMsg := fmt.Sprintf("runtime %s is delete failed %#v, error: %v", uniqueId.Name, uniqueId, err)
				batchRuntimeDeleteResult.Failed++
				batchRuntimeDeleteResult.ErrMsg = append(batchRuntimeDeleteResult.ErrMsg, errMsg)
				batchRuntimeDeleteResult.UnDeletedIds = append(batchRuntimeDeleteResult.UnDeletedIds, runtime.ID)
				batchRuntimeDeleteResult.UnDeleted = append(batchRuntimeDeleteResult.UnDeleted, curr)
				continue
			}

			batchRuntimeDeleteResult.Success++
			batchRuntimeDeleteResult.DeletedIds = append(batchRuntimeDeleteResult.DeletedIds, runtime.ID)
			batchRuntimeDeleteResult.Deleted = append(batchRuntimeDeleteResult.Deleted, *runtimedto)
		}
	}
	return batchRuntimeDeleteResult
}

// processRuntimeScaleRecord 处理单个 runtime 对应的 scale 操作
func (r *Service) processRuntimeScaleRecord(rsc apistructs.RuntimeScaleRecord, action string) (apistructs.PreDiceDTO, error, string) {
	uniqueId := spec.RuntimeUniqueId{
		ApplicationId: rsc.ApplicationId,
		Workspace:     rsc.Workspace,
		Name:          rsc.Name,
	}
	serviceManualScale := false
	logrus.Errorf("process runtime scale for runtime %#v", uniqueId)

	appliedScaledObjects, vpaObjects, err := r.AppliedScaledObjects(uniqueId)
	if err != nil {
		logrus.Warnf("get applied hpa rules for RuntimeUniqueId %#v failed: %v", uniqueId, err)
	}

	for svcName := range rsc.PayLoad.Services {
		if _, ok := appliedScaledObjects[svcName]; ok {
			errMsg := fmt.Sprintf("hpa rules has applied for RuntimeUniqueId %#v, can not do this scale action, please delete or canel the applied rules first", uniqueId)
			return apistructs.PreDiceDTO{}, errors.Errorf("hpa rules has applied for RuntimeUniqueId %#v, can not do this scale action, please delete or canel the applied rules first", uniqueId), errMsg
		}
		if _, ok := vpaObjects[svcName]; ok {
			errMsg := fmt.Sprintf("vpa rules has applied for RuntimeUniqueId %#v, can not do this scale action, please delete or canel the applied rules first", uniqueId)
			return apistructs.PreDiceDTO{}, errors.Errorf("vpa rules has applied for RuntimeUniqueId %#v, can not do this scale action, please delete or canel the applied rules first", uniqueId), errMsg
		}
	}

	pre, err := r.db.FindPreDeployment(uniqueId)
	if err != nil {
		logrus.Errorf("process runtime scale failed, find PreDeployment record in table ps_v2_pre_builds failed for runtime %#v, err: %v", uniqueId, err)
		errMsg := fmt.Sprintf("find PreDeployment record in table ps_v2_pre_builds failed for runtime %#v, err: %v", uniqueId, err)
		return apistructs.PreDiceDTO{}, err, errMsg
	}

	oldOverlay, err := getPreDeploymentDiceOverlay(pre)
	if err != nil {
		return apistructs.PreDiceDTO{}, err, fmt.Sprintf("%s", err)
	}

	// Global Envs
	if rsc.PayLoad.Envs != nil {
		oldOverlay.Envs = rsc.PayLoad.Envs
	}
	var (
		needUpdateServices     []string
		oldOverlayDataForAudit = apistructs.PreDiceDTO{
			Services: make(map[string]*apistructs.RuntimeInspectServiceDTO, 0),
		}
	)
	oldServiceReplicas := make(map[string]int)
	for k, v := range rsc.PayLoad.Services {
		oldService, exists := oldOverlay.Services[k]
		if !exists || oldService == nil {
			logrus.Errorf("process runtime scale failed, can not found service %s in table ps_v2_pre_builds in filed dice or dice_overlay", k)
			errMsg := fmt.Sprintf("process runtime scale failed, can not found service %s in table ps_v2_pre_builds in filed dice or dice_overlay", k)
			return apistructs.PreDiceDTO{}, err, errMsg
		}
		// Local Envs
		if v.Envs != nil {
			oldService.Envs = v.Envs
		}
		oldOverlayDataForAudit.Services[k] = genOverlayDataForAudit(oldService)
		needUpdateServices = append(needUpdateServices, k)
		// Replicas
		oldService.Deployments.Replicas = v.Deployments.Replicas
		oldOverlay.Services[k].Deployments.Replicas = v.Deployments.Replicas
		// Resources
		oldService.Resources.CPU = v.Resources.CPU
		oldOverlay.Services[k].Resources.CPU = v.Resources.CPU
		oldService.Resources.Mem = v.Resources.Mem
		oldOverlay.Services[k].Resources.Mem = v.Resources.Mem
		oldService.Resources.Disk = v.Resources.Disk
	}

	// really update scale
	if len(needUpdateServices) != 0 {
		runtime, err := r.db.FindRuntime(uniqueId)
		if err != nil {
			logrus.Errorf("process runtime scale failed, find runtime record in ps_v2_project_runtimes for runtime %#v failed, err: %v", uniqueId, err)
			errMsg := fmt.Sprintf("find runtime record in ps_v2_project_runtimes for runtime %#v failed, err: %v", uniqueId, err)
			return apistructs.PreDiceDTO{}, err, errMsg
		}
		if runtime == nil {
			logrus.Errorf("process runtime scale failed, runtime %s is not existed for runtime %#v", uniqueId.Name, uniqueId)
			errMsg := fmt.Sprintf("runtime %s is not existed for runtime %#v", uniqueId.Name, uniqueId)
			return apistructs.PreDiceDTO{}, errors.Errorf("runtime %s is not existed", uniqueId.Name), errMsg
		}

		namespace, name := runtime.ScheduleName.Args()
		sg := &apistructs.ServiceGroup{
			ClusterName: runtime.ClusterName,
			Dice: apistructs.Dice{
				ID:       name,
				Type:     namespace,
				Services: make([]apistructs.Service, 0),
			},
		}
		for _, svcName := range needUpdateServices {
			sg.Services = append(sg.Services, apistructs.Service{
				Name:  svcName,
				Scale: oldOverlay.Services[svcName].Deployments.Replicas,
				Resources: apistructs.Resources{
					Cpu:  oldOverlay.Services[svcName].Resources.CPU,
					Mem:  float64(oldOverlay.Services[svcName].Resources.Mem),
					Disk: float64(oldOverlay.Services[svcName].Resources.Disk),
				},
			})
		}

		// 未设置 action，则需要自行判断用于执行 scale 操作之后的 runtime 状态更新
		if action == "" {
			serviceManualScale = true
			// 如果所有的 serice 的副本数都调整为0，则是停止操作，Runtime 状态应该为 Stopped，否则不是停止操作，Runtime 状态应该为 Healthy
			action = apistructs.ScaleActionDown
			for _, svc := range sg.Services {
				if svc.Scale > 0 {
					action = apistructs.ScaleActionUp
					break
				}
			}
		}

		sgb, _ := json.Marshal(&sg)
		logrus.Debugf("scale service group body is %s", string(sgb))

		if _, err := r.scheduler.Httpendpoints.ServiceGroupImpl.Scale(sg); err != nil {
			logrus.Errorf("process runtime scale failed, ScaleServiceGroup failed for runtime %s for runtime %#v failed, err: %v", uniqueId.Name, uniqueId, err)
			errMsg := fmt.Sprintf("ScaleServiceGroup failed for runtime %s for runtime %#v failed, err: %v", uniqueId.Name, uniqueId, err)
			return apistructs.PreDiceDTO{}, err, errMsg
		}

		if (action == apistructs.ScaleActionDown || action == apistructs.ScaleActionUp) && !serviceManualScale {
			addons, err := r.db.GetUnDeletableAttachMentsByRuntimeID(runtime.OrgID, runtime.ID)
			if err != nil {
				logrus.Warnf("process runtime scale successed, but update runtime referenced addon attact_count for runtime %#v failed, err: %v", uniqueId, err)
			}
			switch action {
			case apistructs.ScaleActionDown:
				runtime.Status = apistructs.RuntimeStatusStopped
				if err := r.db.UpdateRuntime(runtime); err != nil {
					logrus.Warnf("process runtime scale successed, but update runtime_status to 'Stopped' in %s for runtime %#v failed, err: %v", runtime.TableName(), uniqueId, err)
				}

				for _, att := range *addons {
					addonRouting, err := r.db.GetInstanceRouting(att.RoutingInstanceID)
					if err != nil {
						logrus.Warnf("process runtime scale successed, but update runtime referenced addon attact_count for runtime %#v failed, err: %v", uniqueId, err)
						continue
					}

					if addonRouting != nil && addonRouting.PlatformServiceType == 0 {
						att.Deleted = apistructs.AddonScaleDown
						if err := r.db.UpdateAttachment(&att); err != nil {
							logrus.Warnf("process runtime scale successed, but update table %s for runtime %#v failed, err: %v", att.TableName(), uniqueId, err)
						}
					}
				}

			case apistructs.ScaleActionUp:
				runtime.Status = apistructs.RuntimeStatusHealthy
				if err := r.db.UpdateRuntime(runtime); err != nil {
					logrus.Warnf("process runtime scale successed, but update runtime_status to 'Healthy' in %s for runtime %#v failed, err: %v", runtime.TableName(), uniqueId, err)
				}

				for _, att := range *addons {
					addonRouting, err := r.db.GetInstanceRouting(att.RoutingInstanceID)
					if err != nil {
						logrus.Warnf("process runtime scale successed, but update runtime referenced addon attact_count for runtime %#v failed, err: %v", uniqueId, err)
						continue
					}

					if addonRouting != nil && addonRouting.PlatformServiceType == 0 {
						att.Deleted = apistructs.AddonNotDeleted
						if err := r.db.UpdateAttachment(&att); err != nil {
							logrus.Warnf("process runtime scale successed, but update table %s for runtime %#v failed, err: %v", att.TableName(), uniqueId, err)
						}
					}
				}
			}
		}
	}

	// 保留非停止状态下的副本数
	if action != apistructs.ScaleActionDown {
		for k := range oldOverlay.Services {
			if oldOverlay.Services[k].Deployments.Replicas == 0 {
				if replicas, ok := oldServiceReplicas[k]; ok {
					if replicas > 0 {
						oldOverlay.Services[k].Deployments.Replicas = replicas
					}
				}
			}
		}
		// save changes
		o_, err := json.Marshal(oldOverlay)
		if err != nil {
			logrus.Errorf("process runtime scale failed, ScaleServiceGroup for runtime %s for runtime %#v successfully, but json.Marshal diceyml.Object oldOverlay %#v failed, err: %v", uniqueId.Name, uniqueId, oldOverlay, err)
			errMsg := fmt.Sprintf("ScaleServiceGroup for runtime %s for runtime %#v successfully, but json.Marshal diceyml.Object oldOverlay %#v failed, err: %v", uniqueId.Name, uniqueId, oldOverlay, err)
			return apistructs.PreDiceDTO{}, err, errMsg

		}
		pre.DiceOverlay = string(o_)
		if err := r.db.UpdatePreDeployment(pre); err != nil {
			logrus.Errorf("process runtime scale failed, ScaleServiceGroup for runtime %s for runtime %#v successfully, but update table ps_v2_pre_builds failed, err: %v", uniqueId.Name, uniqueId, err)
			errMsg := fmt.Sprintf("ScaleServiceGroup for runtime %s for runtime %#v successfully, but update table ps_v2_pre_builds failed, err: %v", uniqueId.Name, uniqueId, err)
			return apistructs.PreDiceDTO{}, err, errMsg
		}
	}

	logrus.Errorf("process runtime scale for runtime %#v successfully", uniqueId)
	return oldOverlayDataForAudit, nil, ""
}

func genOverlayDataForAudit(oldService *diceyml.Service) *apistructs.RuntimeInspectServiceDTO {
	return &apistructs.RuntimeInspectServiceDTO{
		Resources: apistructs.RuntimeServiceResourceDTO{
			CPU:  oldService.Resources.CPU,
			Mem:  oldService.Resources.Mem,
			Disk: oldService.Resources.Disk,
		},
		Deployments: apistructs.RuntimeServiceDeploymentsDTO{
			Replicas: oldService.Deployments.Replicas,
		},
	}
}
