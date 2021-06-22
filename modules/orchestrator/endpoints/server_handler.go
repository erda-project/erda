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

package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/orchestrator/spec"
	"github.com/erda-project/erda/modules/orchestrator/utils"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

// Deprecated
func (s *Endpoints) epBulkGetRuntimeStatusDetail(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	rs_ := r.URL.Query().Get("runtimeIds")
	rs := strings.Split(rs_, ",")
	var runtimeIds []uint64
	for _, r := range rs {
		runtimeId, err := strconv.ParseUint(r, 10, 64)
		if err != nil {
			return utils.ErrRespIllegalParam(err, fmt.Sprintf("failed to bulk get runtime StatusDetail, invalid runtimeIds: %s", rs_))
		}
		runtimeIds = append(runtimeIds, runtimeId)
	}
	funcErrMsg := fmt.Sprintf("failed to bulk get runtime StatusDetail, runtimeIds: %v", runtimeIds)
	runtimes, err := s.db.FindRuntimesByIds(runtimeIds)
	if err != nil {
		return utils.ErrResp0101(err, funcErrMsg)
	}
	if _, err := user.GetUserID(r); err != nil {
		return apierrors.ErrGetRuntime.NotLogin().ToResp(), nil
	}
	data := make(map[uint64]interface{})
	for _, r := range runtimes {
		if status, err := s.bdl.GetServiceGroupStatus(r.ScheduleName.Args()); err != nil {
			return utils.ErrResp0101(err, fmt.Sprintf("failed to bulk get runtime StatusDetail, runtimeId: %d", r.ID))
		} else {
			data[r.ID] = status
		}
	}
	return httpserver.OkResp(data)
}

// Deprecated
func (s *Endpoints) epUpdateOverlay(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	appId_ := r.URL.Query().Get("applicationId")
	appId, err := strconv.Atoi(appId_)
	if err != nil {
		return utils.ErrRespIllegalParam(err, fmt.Sprintf("failed to update Overlay, appId invalid: %v", appId_))
	}
	workspace := r.URL.Query().Get("workspace")
	if len(workspace) == 0 {
		return utils.ErrRespIllegalParam(errors.Errorf("workspace invalid: %v", workspace), "failed to update Overlay")
	}
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrUpdateRuntime.NotLogin().ToResp(), nil
	}
	perm, err := s.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  uint64(appId),
		Resource: "runtime-" + strutil.ToLower(workspace),
		Action:   apistructs.OperateAction,
	})
	if err != nil {
		return apierrors.ErrUpdateRuntime.InternalError(err).ToResp(), nil
	}
	if !perm.Access {
		return apierrors.ErrUpdateRuntime.AccessDenied().ToResp(), nil
	}
	runtimeName := r.URL.Query().Get("runtimeName")
	if len(runtimeName) == 0 {
		return utils.ErrRespIllegalParam(errors.Errorf("runtimeName invalid: %v", runtimeName), "failed to update Overlay")
	}
	var overlay apistructs.PreDiceDTO
	if err := json.NewDecoder(r.Body).Decode(&overlay); err != nil {
		return utils.ErrRespIllegalParam(err, "failed to update Overlay, failed to parse req")
	}
	uniqueId := spec.RuntimeUniqueId{
		ApplicationId: uint64(appId),
		Workspace:     workspace,
		Name:          runtimeName,
	}
	funcErrMsg := fmt.Sprintf("failed to update Overlay, uniqueId: %v", uniqueId)
	pre, err := s.db.FindPreDeployment(uniqueId)
	if err != nil {
		return utils.ErrResp0101(err, funcErrMsg)
	}
	var oldOverlay diceyml.Object
	if pre.DiceOverlay != "" {
		if err = json.Unmarshal([]byte(pre.DiceOverlay), &oldOverlay); err != nil {
			return utils.ErrResp0101(err, funcErrMsg)
		}
	}
	// Global Envs
	if overlay.Envs != nil {
		oldOverlay.Envs = overlay.Envs
	}
	var needUpdateServices []string
	for k, v := range overlay.Services {
		oldService, exists := oldOverlay.Services[k]
		if !exists || oldService == nil {
			oldService = &diceyml.Service{}
			if oldOverlay.Services == nil {
				oldOverlay.Services = map[string]*diceyml.Service{}
			}
			oldOverlay.Services[k] = oldService
		}
		// Local Envs
		if v.Envs != nil {
			oldService.Envs = v.Envs
		}
		// record need update scale's service
		if oldService.Resources.CPU != v.Resources.CPU || oldService.Resources.Mem != v.Resources.Mem ||
			oldService.Resources.Disk != v.Resources.Disk || oldService.Deployments.Replicas != v.Deployments.Replicas {
			needUpdateServices = append(needUpdateServices, k)
		}
		// Replicas
		oldService.Deployments.Replicas = v.Deployments.Replicas
		// Resources
		oldService.Resources.CPU = v.Resources.CPU
		oldService.Resources.Mem = v.Resources.Mem
		oldService.Resources.Disk = v.Resources.Disk
	}

	// really update scale
	if len(needUpdateServices) != 0 {
		runtime, err := s.db.FindRuntime(uniqueId)
		if err != nil {
			return utils.ErrResp0101(err, funcErrMsg)
		}
		if runtime == nil {
			return utils.ErrResp0101(errors.Errorf("runtime %s is not existed", uniqueId.Name), funcErrMsg)
		}
		namespace, name := runtime.ScheduleName.Args()
		sg := apistructs.UpdateServiceGroupScaleRequst{
			Name:        name,
			Namespace:   namespace,
			ClusterName: runtime.ClusterName,
		}

		// TODO really update service to k8s deployment
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

		sgb, _ := json.Marshal(&sg)
		logrus.Debugf("scale service group body is %s", string(sgb))
		// TODO: Need to increase the mechanism of failure compensation
		if err := s.bdl.ScaleServiceGroup(sg); err != nil {
			return utils.ErrResp0101(err, funcErrMsg)
		}
	}

	// save changes
	o_, err := json.Marshal(oldOverlay)
	if err != nil {
		return utils.ErrResp0101(err, funcErrMsg)
	}
	pre.DiceOverlay = string(o_)
	if err := s.db.UpdatePreDeployment(pre); err != nil {
		return utils.ErrResp0101(err, funcErrMsg)
	}

	return httpserver.OkResp(nil)
}
