package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
	"github.com/erda-project/erda/modules/ops/impl/aliyun-resources/nat"
	"github.com/erda-project/erda/modules/ops/impl/aliyun-resources/vpc"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

func (e *Endpoints) AddCloudClusters(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("error happened, error:%v", err)
			resp, err = mkResponse(apistructs.CloudClusterResponse{
				Header: apistructs.Header{
					Success: false,
					Error:   apistructs.ErrorResponse{Msg: err.Error()},
				},
			})
		}
	}()

	var req apistructs.CloudClusterRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("failed to unmarshal to apistructs.CloudClusterRequest: %v", err)
		return
	}
	logrus.Debugf("cloud-cluster request: %v", req)

	i, resp := e.GetIdentity(r)
	if resp != nil {
		err = fmt.Errorf("failed to get User-ID or Org-ID from request header")
		return
	}

	// permission check
	err = e.PermissionCheck(i.UserID, i.OrgID, "", apistructs.CreateAction)
	if err != nil {
		return
	}

	// add cluster using cloud resource, vpc already exist
	if req.VpcID != "" {
		ak_ctx, resp := e.mkCtx(ctx, i.OrgID)
		if resp != nil {
			return errorresp.ErrResp(fmt.Errorf("gte cloud resource context failed, err:%v", err))
		}
		req.AccessKey = ak_ctx.AccessKeyID
		req.SecretKey = ak_ctx.AccessSecret
		ak_ctx.Region = req.Region
		rsp, er := vpc.DescribeVPCs(ak_ctx, aliyun_resources.DefaultPageOption, "", req.VpcID)
		if er != nil {
			err = fmt.Errorf("describe vpc failed, error:%v", er)
			logrus.Error(err.Error())
			return errorresp.ErrResp(err)
		}
		// nat gateway already exist
		if len(rsp.Vpcs.Vpc) > 0 && len(rsp.Vpcs.Vpc[0].NatGatewayIds.NatGatewayIds) > 0 {
			// assign nat gateway id to prevent to recreate
			req.NatGatewayID = rsp.Vpcs.Vpc[0].NatGatewayIds.NatGatewayIds[0]
			rsp, er := nat.DescribeResource(ak_ctx, aliyun_resources.DefaultPageOption, req.NatGatewayID)
			if er != nil {
				err = fmt.Errorf("describe nat gateway failed, error:%v", er)
				logrus.Error(err.Error())
				return errorresp.ErrResp(err)
			}
			if len(rsp.NatGateways.NatGateway) > 0 {
				gateway := rsp.NatGateways.NatGateway[0]
				fTableId := gateway.ForwardTableIds.ForwardTableId[0]
				sTableId := gateway.SnatTableIds.SnatTableId[0]
				isVswBoundSnat, er := nat.IsVswitchBoundSnat(ak_ctx, sTableId, req.VSwitchID)
				if er != nil {
					err = fmt.Errorf("check vsw snat bound info failed, error:%v", er)
					return errorresp.ErrResp(err)
				}
				// vsw not bound with snat, pass snat table id to bound it
				if !isVswBoundSnat {
					req.ForwardTableID = fTableId
					req.SnatTableID = sTableId
				}
			} else {
				err = fmt.Errorf("empty vpc natgateway table id")
				logrus.Errorf(err.Error())
				return errorresp.ErrResp(err)
			}
		}
	}

	recordID, err := e.clusters.AddClusters(req, i.UserID)
	if err != nil {
		err = fmt.Errorf("failed to add clusters: %v", err)
		return
	}
	return mkResponse(apistructs.CloudClusterResponse{
		Header: apistructs.Header{Success: true},
		Data:   apistructs.AddNodesData{RecordID: recordID},
	})
}

// lock cluster for auto scale
func (e *Endpoints) LockCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.LockCluster
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errStr := fmt.Sprintf("failed to unmarshal to apistructs.LockCluster: %v", err)
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errStr},
			},
		})
	}
	userid := r.Header.Get("User-ID")
	if userid == "" {
		errStr := fmt.Sprintf("failed to get user-id in http header")
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errStr},
			},
		})
	}

	err := e.Mns.LockCluster(req.ClusterName)
	if err != nil {
		errStr := fmt.Errorf("lock cluster failed, request: %v, error: %v", req, err)
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errStr.Error()},
			},
		})
	}
	return mkResponse(apistructs.CloudClusterResponse{
		Header: apistructs.Header{
			Success: true,
		},
	})
}

// unlock cluster for auto scale
func (e *Endpoints) UnlockCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.LockCluster
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errStr := fmt.Sprintf("failed to unmarshal to apistructs.LockCluster: %v", err)
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errStr},
			},
		})
	}
	userid := r.Header.Get("User-ID")
	if userid == "" {
		errStr := fmt.Sprintf("failed to get user-id in http header")
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errStr},
			},
		})
	}

	err := e.Mns.UnlockCluster(req.ClusterName)
	if err != nil {
		err := fmt.Errorf("unlock cluster failed, request: %v, error: %v", req, err)
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: err.Error()},
			},
		})
	}
	return mkResponse(apistructs.CloudClusterResponse{
		Header: apistructs.Header{
			Success: true,
		},
	})
}
