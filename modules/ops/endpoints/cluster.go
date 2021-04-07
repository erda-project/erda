package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/ops/impl/ess"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

func (e *Endpoints) UpgradeEdgeCluster(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("error happened: %+v", err)
			resp, err = mkResponse(apistructs.UpgradeEdgeClusterResponse{
				Header: apistructs.Header{
					Success: false,
					Error:   apistructs.ErrorResponse{Msg: err.Error()},
				},
			})
		}
	}()
	var req apistructs.UpgradeEdgeClusterRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("failed to unmarshal request: %+v", err)
		return
	}

	// get identity info
	i, resp := e.GetIdentity(r)
	if resp != nil {
		err = fmt.Errorf("failed to get User-ID or Org-ID from request header")
		return
	}
	// permission check
	err = e.PermissionCheck(i.UserID, i.OrgID, "", apistructs.UpdateAction)
	if err != nil {
		return
	}

	recordID, status, precheckHint, err := e.clusters.UpgradeEdgeCluster(req, i.UserID, i.OrgID)
	if err != nil {
		err = fmt.Errorf("failed to upgrade cluster: %v", err)
		return
	}
	return mkResponse(apistructs.UpgradeEdgeClusterResponse{
		Header: apistructs.Header{Success: true},
		Data:   apistructs.UpgradeEdgeClusterData{RecordID: recordID, Status: status, PrecheckHint: precheckHint},
	})
}

func (e *Endpoints) BatchUpgradeEdgeCluster(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("error happened: %+v", err)
			resp, err = mkResponse(apistructs.BatchUpgradeEdgeClusterResponse{
				Header: apistructs.Header{
					Success: false,
					Error:   apistructs.ErrorResponse{Msg: err.Error()},
				},
			})
		}
	}()
	var req apistructs.BatchUpgradeEdgeClusterRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("failed to unmarshal request: %+v", err)
		return
	}
	logrus.Debugf("batch upgrade request header:%+v", r.Header)
	// get identity info
	userid := r.Header.Get("User-ID")
	orgid := r.Header.Get("Org-ID")
	if userid == "" && orgid == "" {
		err = fmt.Errorf("failed to get User-ID or Org-ID from request header")
		return
	}
	// permission check
	err = e.PermissionCheck(userid, orgid, "", apistructs.UpdateAction)
	if err != nil {
		return
	}

	go e.clusters.BatchUpgradeEdgeCluster(req, userid)

	return mkResponse(apistructs.BatchUpgradeEdgeClusterResponse{
		Header: apistructs.Header{Success: true},
	})
}

func (e *Endpoints) OrgClusterInfo(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("error happened: %+v", err)
			resp, err = mkResponse(apistructs.OrgClusterInfoResponse{
				Header: apistructs.Header{
					Success: false,
					Error:   apistructs.ErrorResponse{Msg: err.Error()},
				},
				Data: apistructs.OrgClusterInfoData{
					List: []apistructs.OrgClusterInfoBasicData{},
				},
			})
		}
	}()
	req := apistructs.OrgClusterInfoRequest{}
	req.OrgName = r.URL.Query().Get("orgName")
	req.ClusterType = r.URL.Query().Get("clusterType")
	pageSize := r.URL.Query().Get("pageSize")
	pageNo := r.URL.Query().Get("pageNo")
	req.PageSize, err = strconv.Atoi(pageSize)
	if err != nil {
		logrus.Errorf("failed to parse pageSize")
		req.PageSize = 10
	}
	req.PageNo, err = strconv.Atoi(pageNo)
	if err != nil {
		logrus.Errorf("failed to parse pageNo")
		req.PageNo = 1
	}
	if req.PageNo <= 0 {
		req.PageNo = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	data, er := e.clusters.GetOrgClusterInfo(req)
	if er != nil {
		err = fmt.Errorf("list org cluster info failed")
		logrus.Errorf("%s, request:%+v, error:%v", err.Error(), req, er)
	}

	return mkResponse(apistructs.OrgClusterInfoResponse{
		Header: apistructs.Header{Success: true},
		Data:   data,
	})
}

func (e *Endpoints) OfflineEdgeCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.OfflineEdgeClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("failed to unmarshal to apistructs.OfflineEdgeClusterRequest: %v", err)
		return mkResponse(apistructs.OfflineEdgeClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}

	userid := r.Header.Get("User-ID")
	if userid == "" {
		errstr := fmt.Sprintf("failed to get user-id in http header")
		return mkResponse(apistructs.OfflineEdgeClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}

	orgid := r.Header.Get("Org-ID")
	scopeID, err := strconv.ParseUint(orgid, 10, 64)
	if err != nil {
		logrus.Errorf("parse orgid failed, orgid: %v, error: %v", orgid, err)
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: "parse orgid failed"},
			},
		})
	}

	// permission check
	p := apistructs.PermissionCheckRequest{
		UserID:   userid,
		Scope:    apistructs.OrgScope,
		ScopeID:  scopeID,
		Resource: apistructs.CloudResourceResource,
		Action:   apistructs.DeleteAction,
	}
	rspData, err := e.bdl.CheckPermission(&p)
	if err != nil {
		logrus.Errorf("check permission error: %v", err)
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: "check permission internal error"},
			},
		})
	}
	if !rspData.Access {
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: "access denied"},
			},
		})
	}

	recordID, err := e.clusters.OfflineEdgeCluster(req, userid, orgid)
	if err != nil {
		errstr := fmt.Sprintf("failed to offline cluster: %v", err)
		return mkResponse(apistructs.UpgradeEdgeClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	return mkResponse(apistructs.OfflineEdgeClusterResponse{
		Header: apistructs.Header{Success: true},
		Data:   apistructs.OfflineEdgeClusterData{RecordID: recordID},
	})

}

func (e *Endpoints) ClusterInfo(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	clusternames := r.URL.Query().Get("clusterName")
	clusternameList := strutil.Split(clusternames, ",", true)

	if len(clusternameList) == 0 {
		errstr := "empty clusterName arg"
		return httpserver.ErrResp(200, "1", errstr)
	}

	result, err := e.clusters.ClusterInfo(ctx, clusternameList)
	if err != nil {
		errstr := fmt.Sprintf("failed to get clusterinfo: %s, %v", clusternames, err)
		return httpserver.ErrResp(200, "2", errstr)
	}
	return mkResponse(apistructs.OpsClusterInfoResponse{
		Header: apistructs.Header{Success: true},
		Data:   apistructs.OpsClusterInfoData(result),
	})
}

func (e *Endpoints) ClusterUpdate(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	header := r.Header

	var req apistructs.ClusterUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("failed to parse request body: %v", err)
		return httpserver.ErrResp(200, "2", errstr)
	}

	if req.OpsConfig != nil {
		err := e.handleUpdateReq(&req)
		if err != "ok" {
			return httpserver.ErrResp(200, "2", err)
		}
	}

	err := e.bdl.UpdateCluster(req, header)
	if err != nil {
		errstr := fmt.Sprintf("failed to update clusterinfo: %v", err)
		return httpserver.ErrResp(200, "2", errstr)
	}
	return mkResponse(apistructs.OpsClusterInfoResponse{
		Header: apistructs.Header{Success: true},
		Data:   nil,
	})
}

func (e *Endpoints) handleUpdateReq(req *apistructs.ClusterUpdateRequest) string {
	var as *ess.Ess
	var isEdit bool
	clusterInfo, err := e.bdl.GetCluster(req.Name)
	//TODO: 检验clusterInfo信息
	if err != nil {
		return fmt.Sprintf("failed to get cluster info: %v", err)
	}

	accountInfoReq := apistructs.BasicCloudConf{
		Region:          clusterInfo.OpsConfig.Region,
		AccessKeyId:     clusterInfo.OpsConfig.AccessKey,
		AccessKeySecret: clusterInfo.OpsConfig.SecretKey,
	}
	clusterInfo.OpsConfig.ScaleMode = req.OpsConfig.ScaleMode
	if req.OpsConfig.ScaleMode != "" {
		err = e.validateOpsConfig(clusterInfo.OpsConfig)
		if err != nil {
			return fmt.Sprintf("cluster: %s, error: %v", clusterInfo.Name, err)
		}

		as, err = e.Ess.Init(accountInfoReq, e.Mns, e.nodes)
		if err != nil {
			return fmt.Sprintf("failed to init ess sdk: %v", err)
		}
	}
	if req.OpsConfig.ScaleMode == "auto" {
		if clusterInfo.OpsConfig != nil && clusterInfo.OpsConfig.ScaleMode == apistructs.ScaleModeScheduler {
			_, err := e.bdl.StopPipelineCron(clusterInfo.OpsConfig.ScalePipeLineID)
			if err != nil {
				return fmt.Sprintf("failed to delete pipline cronjob : %v", err)
			}
		}
		name := clusterInfo.Name + ess.EssScaleSchedulerTaskSuff
		err := as.DeleteScheduledTasks(name)
		if err != nil {
			return fmt.Sprintf("failed to delete scheduler task : %v", err)
		}
		err = as.CreateAutoFlow(req.Name, clusterInfo.OpsConfig.VSwitchIDs, clusterInfo.OpsConfig.EcsPassword, clusterInfo.OpsConfig.SgIDs)
		if err != nil {
			return fmt.Sprintf("failed to create auto scale mode: %v", err)
		}
		req.OpsConfig = clusterInfo.OpsConfig
		req.OpsConfig.EssScaleRule = as.Config.EssScaleRule
		req.OpsConfig.EssGroupID = as.Config.EssGroupID
		req.OpsConfig.ScalePipeLineID = 0
	}
	if req.OpsConfig.ScaleMode == apistructs.ScaleModeScheduler {
		if clusterInfo.OpsConfig != nil && clusterInfo.OpsConfig.ScaleMode == apistructs.ScaleModeScheduler {
			if clusterInfo.OpsConfig.ScalePipeLineID != 0 {
				_, err := e.bdl.StopPipelineCron(clusterInfo.OpsConfig.ScalePipeLineID)
				if err != nil {
					return fmt.Sprintf("failed to delete pipline cronjob : %v", err)
				}
			}
			//基于原有的定时伸缩规则进行修改
			if req.OpsConfig.LaunchTime == clusterInfo.OpsConfig.LaunchTime {
				isEdit = true
			}
		}
		clusterInfo.OpsConfig.ScaleNumber = req.OpsConfig.ScaleNumber
		clusterInfo.OpsConfig.LaunchTime = req.OpsConfig.LaunchTime
		clusterInfo.OpsConfig.RepeatMode = req.OpsConfig.RepeatMode
		clusterInfo.OpsConfig.RepeatValue = req.OpsConfig.RepeatValue
		clusterInfo.OpsConfig.ScaleDuration = req.OpsConfig.ScaleDuration
		err := as.CreateSchedulerFlow(apistructs.SchedulerScaleReq{
			ClusterName:     req.Name,
			VSwitchID:       clusterInfo.OpsConfig.VSwitchIDs,
			EcsPassword:     clusterInfo.OpsConfig.EcsPassword,
			SgID:            clusterInfo.OpsConfig.SgIDs,
			OrgID:           clusterInfo.OrgID,
			Region:          clusterInfo.OpsConfig.Region,
			AccessKeyId:     clusterInfo.OpsConfig.AccessKey,
			AccessKeySecret: clusterInfo.OpsConfig.SecretKey,
			Num:             req.OpsConfig.ScaleNumber,
			LaunchTime:      req.OpsConfig.LaunchTime,
			RecurrenceType:  req.OpsConfig.RepeatMode,
			RecurrenceValue: req.OpsConfig.RepeatValue,
			ScaleDuration:   req.OpsConfig.ScaleDuration,
			IsEdit:          isEdit,
			ScheduledTaskId: clusterInfo.OpsConfig.ScheduledTaskId,
		})
		if err != nil {
			return fmt.Sprintf("failed to create scheduler scale mode: %v", err)
		}
		req.OpsConfig = clusterInfo.OpsConfig
		req.OpsConfig.ScheduledTaskId = as.Config.ScheduledTaskId
		req.OpsConfig.ScalePipeLineID = as.Config.ScalePipeLineID
		req.OpsConfig.EssGroupID = as.Config.EssGroupID
	}
	if req.OpsConfig.ScaleMode == "none" {
		if clusterInfo.OpsConfig != nil && clusterInfo.OpsConfig.ScaleMode == apistructs.ScaleModeScheduler {
			_, err := e.bdl.StopPipelineCron(clusterInfo.OpsConfig.ScalePipeLineID)
			if err != nil {
				return fmt.Sprintf("failed to delete pipline cronjob : %v", err)
			}
		}
		name := clusterInfo.Name + ess.EssScaleSchedulerTaskSuff
		err := as.DeleteScheduledTasks(name)
		if err != nil {
			return fmt.Sprintf("failed to delete scheduler task : %v", err)
		}
		req.OpsConfig = clusterInfo.OpsConfig
		req.OpsConfig.ScalePipeLineID = 0
		req.OpsConfig.EssGroupID = as.Config.EssGroupID
	}
	return "ok"
}

func (e *Endpoints) validateOpsConfig(opsConf *apistructs.OpsConfig) error {
	if opsConf == nil {
		err := fmt.Errorf("empty ops config")
		logrus.Error(err.Error())
		return err
	}

	if e.isEmpty(opsConf.AccessKey) || e.isEmpty(opsConf.SecretKey) || e.isEmpty(opsConf.Region) || e.isEmpty(opsConf.EcsPassword) {
		err := fmt.Errorf("invalid ops config")
		return err
	}
	return nil
}

func (e Endpoints) isEmpty(str string) bool {
	return strings.Replace(str, " ", "", -1) == ""
}
