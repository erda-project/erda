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

package assetsvc

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/bdl"
	"github.com/erda-project/erda/modules/dop/dbclient"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

func (svc *Service) UpdateAPIAsset(req *apistructs.UpdateAPIAssetReq) *errorresp.APIError {
	// parameters validation
	if req.URIParams == nil {
		return apierrors.UpdateAPIAsset.MissingParameter("URI parameters")
	}
	if req.URIParams.AssetID == "" {
		return apierrors.UpdateAPIAsset.MissingParameter("assetID")
	}
	if err := apistructs.ValidateAPIAssetID(req.URIParams.AssetID); err != nil {
		return apierrors.UpdateAPIAsset.InvalidParameter("assetID")
	}
	if req.Keys == nil {
		return apierrors.UpdateAPIAsset.MissingParameter("missing request body")
	}
	if req.OrgID == 0 {
		return apierrors.UpdateAPIAsset.MissingParameter("orgID")
	}

	// retrieve asset
	var asset apistructs.APIAssetsModel
	if err := svc.FirstRecord(&asset, map[string]interface{}{
		"org_id":   req.OrgID,
		"asset_id": req.URIParams.AssetID,
	}); err != nil {
		logrus.Errorf("failed to FirstRecord asset, err: %v", err)
		return apierrors.UpdateAPIAsset.InternalError(err)
	}

	// authentication
	rolesSet := bdl.FetchAssetRolesSet(req.OrgID, req.Identity.UserID)
	if written := writePermission(rolesSet, &asset); !written {
		return apierrors.UpdateAPIAsset.AccessDenied()
	}

	var (
		where = map[string]interface{}{
			"org_id":   req.OrgID,
			"asset_id": req.URIParams.AssetID,
		}
		cascadeUpdates = map[string]interface{}{"asset_name": req.Keys["assetName"]}
	)

	req.Keys["updater_id"] = req.Identity.UserID
	req.Keys["updated_at"] = time.Now()

	// validate the length of assetName
	if assetName, ok := req.Keys["assetName"]; ok {
		assetNameS, ok := assetName.(string)
		if !ok {
			return apierrors.UpdateAPIAsset.InternalError(errors.New("assetName must be string"))
		}
		if err := strutil.Validate(assetNameS, strutil.MinLenValidator(1), strutil.MaxLenValidator(191)); err != nil {
			return apierrors.UpdateAPIAsset.InvalidParameter(errors.Wrapf(err, "assetName: %s", assetNameS))
		}
	}

	// check whether the user is doing the operation of clearing the associated project or application
	// (the key without a value)
	var (
		projectID, ok1 = req.Keys["projectID"]
		appID, ok2     = req.Keys["appID"]
		clearProject   = ok1 && projectID == nil
		clearApp       = ok2 && appID == nil
	)
	if clearProject {
		req.Keys["project_name"] = nil
	}
	if clearApp {
		req.Keys["app_name"] = nil
	}

	// update db
	tx := dbclient.Tx()
	defer tx.RollbackUnlessCommitted()

	if err := tx.Model(new(apistructs.APIAssetsModel)).
		Where(where).
		Updates(req.Keys).
		Error; err != nil {
		return apierrors.UpdateAPIAsset.InternalError(err)
	}

	// cascade update assetName in table access
	if cascadeUpdates["asset_name"] == nil {
		return nil
	}
	if err := tx.Model(new(apistructs.APIAccessesModel)).Where(where).Updates(cascadeUpdates).Error; err != nil {
		return apierrors.UpdateAPIAsset.InternalError(errors.Wrap(err, "failed to UpdateAccess"))
	}

	// cascade update assetName in table version
	if err := tx.Model(new(apistructs.APIAssetVersionsModel)).Where(where).Updates(cascadeUpdates).Error; err != nil {
		return apierrors.UpdateAPIAsset.InternalError(errors.Wrap(err, "failed to UpdateVersion"))
	}

	// cascade update assetName in table contract
	if err := tx.Model(new(apistructs.ContractModel)).Where(where).Updates(cascadeUpdates).Error; err != nil {
		return apierrors.UpdateAPIAsset.InternalError(errors.Wrap(err, "failed to UpdateContract"))
	}

	tx.Commit()

	return nil
}

func (svc *Service) UpdateInstantiation(req *apistructs.UpdateInstantiationReq) (*apistructs.InstantiationModel, *errorresp.APIError) {
	// parameters validation
	if req.URIParams == nil {
		return nil, apierrors.UpdateInstantiation.MissingParameter("URI parameters")
	}
	if err := apistructs.ValidateAPIAssetID(req.URIParams.AssetID); err != nil {
		return nil, apierrors.UpdateInstantiation.InvalidParameter("assetID")
	}
	if req.Body == nil {
		return nil, apierrors.UpdateInstantiation.InvalidParameter("missing request body")
	}
	if req.OrgID == 0 {
		return nil, apierrors.UpdateInstantiation.InvalidParameter("OrgID")
	}

	var (
		asset  apistructs.APIAssetsModel
		access apistructs.APIAccessesModel
	)
	// retrieve asset
	if err := svc.FirstRecord(&asset, map[string]interface{}{
		"org_id":   req.OrgID,
		"asset_id": req.URIParams.AssetID,
	}); err != nil {
		logrus.Errorf("failed to FirstRecord asset, assetID: %s, err: %v", req.URIParams.AssetID, err)
		return nil, apierrors.UpdateInstantiation.InternalError(err)
	}

	// authentication
	rolesSet := bdl.FetchAssetRolesSet(req.OrgID, req.Identity.UserID)
	if written := writePermission(rolesSet, &asset); !written {
		return nil, apierrors.UpdateInstantiation.AccessDenied()
	}

	// can not update if there is an Access
	if err := svc.FirstRecord(&access, map[string]interface{}{
		"org_id":          req.OrgID,
		"asset_id":        req.URIParams.AssetID,
		"swagger_version": req.URIParams.SwaggerVersion,
		"minor":           req.URIParams.Minor,
	}); err == nil {
		return nil, apierrors.UpdateInstantiation.InternalError(errors.New("???????????????????????????, ????????????"))
	}

	// parse the url of runtime instance
	var (
		urlStr = req.Body.URL
	)
	if urlStr == "" {
		// delete the association relationship with the instance if the giving url is ""
		if err := dbclient.Sq().Delete(new(apistructs.InstantiationModel), map[string]interface{}{
			"id": req.URIParams.InstantiationID,
		}).Error; err != nil {
			return nil, apierrors.UpdateInstantiation.InternalError(errors.Wrap(err, "failed to clear instance url"))
		}
		return nil, nil
	}

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") && !strings.HasPrefix(urlStr, "//") {
		urlStr = "//" + strings.TrimLeft(urlStr, "/")
	}
	_, err := url.Parse(urlStr)
	if err != nil {
		return nil, apierrors.UpdateInstantiation.InternalError(errors.Wrap(err, "invalid instance url"))
	}

	if err := dbclient.UpdateInstantiation(req); err != nil {
		return nil, apierrors.UpdateInstantiation.InternalError(err)
	}

	instantiations, ok, apiError := svc.GetInstantiation(&apistructs.GetInstantiationsReq{
		OrgID:    req.OrgID,
		Identity: req.Identity,
		URIParams: &apistructs.GetInstantiationsURIParams{
			AssetID:        req.URIParams.AssetID,
			SwaggerVersion: req.URIParams.SwaggerVersion,
			Minor:          req.URIParams.Minor,
		},
	})
	if apiError != nil {
		return nil, apiError
	}
	if !ok {
		return nil, apierrors.UpdateInstantiation.InternalError(errors.New("failed to GetInstantiations"))
	}

	return instantiations, nil
}

func (svc *Service) UpdateClient(req *apistructs.UpdateClientReq) (*apistructs.ClientModel, *apistructs.SK, *errorresp.APIError) {
	if req == nil || req.URIParams == nil || req.QueryParams == nil || req.Body == nil {
		return nil, nil, apierrors.UpdateClient.InvalidParameter("invalid parameter or body")
	}

	var (
		model apistructs.ClientModel
		where = map[string]interface{}{
			"org_id": req.OrgID,
			"id":     req.URIParams.ClientID,
		}
	)
	if err := svc.FirstRecord(&model, where); err != nil {
		return nil, nil, apierrors.UpdateClient.InternalError(err)
	}

	var sk apistructs.SK
	if req.QueryParams.ResetClientSecret {
		credentials, err := bdl.Bdl.ResetClientCredentials(model.ClientID)
		if err != nil {
			return nil, nil, apierrors.UpdateClient.InternalError(err)
		}
		sk.ClientID = credentials.ClientId
		sk.ClientSecret = credentials.ClientSecret
	} else {
		credentials, err := bdl.Bdl.GetClientCredentials(model.ClientID)
		if err != nil {
			return nil, nil, apierrors.UpdateClient.InternalError(err)
		}
		sk.ClientID = credentials.ClientId
		sk.ClientSecret = credentials.ClientSecret
	}

	var (
		updates = map[string]interface{}{
			"displayName": req.Body.DisplayName,
			"desc":        req.Body.Desc,
			"updater_id":  req.Identity.UserID,
			"updated_at":  time.Now(),
		}
	)
	if err := dbclient.Sq().Model(&model).Where(where).Updates(updates).Error; err != nil {
		return nil, nil, apierrors.UpdateClient.InternalError(err)
	}
	return &model, &sk, nil
}

func (svc *Service) UpdateContract(req *apistructs.UpdateContractReq) (*apistructs.ClientModel, *apistructs.ContractModel, *errorresp.APIError) {
	if req == nil || req.URIParams == nil || req.Body == nil {
		return nil, nil, apierrors.UpdateContract.InvalidParameter("invalid parameter")
	}

	contractID, err := strconv.ParseUint(req.URIParams.ContractID, 10, 64)
	if err != nil {
		return nil, nil, apierrors.UpdateContract.InvalidParameter("invalid contract id")
	}

	var (
		client   apistructs.ClientModel
		asset    apistructs.APIAssetsModel
		contract apistructs.ContractModel
		access   apistructs.APIAccessesModel
		where    = map[string]interface{}{
			"org_id": req.OrgID,
			"id":     contractID,
		}
	)

	// retrieve the client
	if err := svc.FirstRecord(&client, map[string]interface{}{
		"org_id": req.OrgID,
		"id":     req.URIParams.ClientID,
	}); err != nil {
		logrus.Errorf("failed to FirstRecord client, err: %v", err)
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil, apierrors.UpdateContract.InternalError(errors.New("??????????????????"))
		}
		return nil, nil, apierrors.UpdateContract.InternalError(err)
	}

	// retrieve the contract
	if err := svc.FirstRecord(&contract, where); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil, apierrors.UpdateContract.InternalError(errors.New("contract not found"))
		}
		return nil, nil, apierrors.UpdateContract.InternalError(err)
	}

	// retrieve the  asset
	if err := svc.FirstRecord(&asset, map[string]interface{}{
		"org_id":   req.OrgID,
		"asset_id": contract.AssetID,
	}); err != nil {
		logrus.Errorf("failed to FirstRecord asset, err: %v", err)
		return nil, nil, apierrors.UpdateContract.InternalError(err)
	}

	// authentication: can the user approve it ?
	// ?????? ????????????????????????????????????
	rolesSet := bdl.FetchAssetRolesSet(req.OrgID, req.Identity.UserID)
	written := writePermission(rolesSet, &asset)

	// retrieve access
	if err := svc.FirstRecord(&access, map[string]interface{}{
		"org_id":          req.OrgID,
		"asset_id":        contract.AssetID,
		"swagger_version": contract.SwaggerVersion,
	}); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil, apierrors.UpdateContract.InternalError(errors.New("access not found"))
		}
		return nil, nil, apierrors.UpdateContract.InternalError(err)
	}

	defer func() {
		err := svc.createOrUpdateClientLimits(access.EndpointID, client.ClientID, contractID)
		if err != nil {
			logrus.Errorf("createOrUpdateClientLimits failed, err:%+v", err)
		}
	}()

	switch {
	case req.Body.RequestSLAID != nil:
		// If the current user is not an org administrator or the client creator, he cannot apply for a new SLA
		// ???????????????????????????, ???????????????????????????, ????????????????????? SLA
		if !inSlice(strconv.FormatUint(req.OrgID, 10), rolesSet.RolesOrgs(bdl.OrgMRoles...)) && req.Identity.UserID != contract.CreatorID {
			return nil, nil, apierrors.UpdateContract.AccessDenied()
		}
		if err := svc.updateContractRequestSLA(req, &contract, &access, &asset); err != nil {
			logrus.Errorf("failed to updateContractRequestSLA, err: %v", err)
			return nil, nil, apierrors.UpdateContract.InternalError(errors.New("?????? SLA ??????"))
		}

	case req.Body.Status != nil:
		if !written {
			return nil, nil, apierrors.UpdateContract.AccessDenied()
		}

		if err := svc.updateContractStatus(req, &client, &access, &contract); err != nil {
			logrus.Errorf("failed to updateContractStatus, err: %v", err)
			return nil, nil, apierrors.UpdateContract.InternalError(errors.New("??????????????????????????????"))
		}

	case req.Body.CurSLAID != nil:
		if !written {
			return nil, nil, apierrors.UpdateContract.AccessDenied()
		}

		if err := svc.updateContractCurSLA(req, &contract, &client, &access); err != nil {
			logrus.Errorf("failed to updateContractCurSLA, err: %v", err)
			return nil, nil, apierrors.UpdateContract.InternalError(errors.New("?????? SLA ??????"))
		}

	default:
		return nil, nil, apierrors.UpdateContract.InvalidParameter("??????????????????")
	}

	return &client, &contract, nil
}

// the manager modifies the contract status (approve the contract)
func (svc *Service) updateContractStatus(req *apistructs.UpdateContractReq, client *apistructs.ClientModel, access *apistructs.APIAccessesModel,
	contract *apistructs.ContractModel) error {
	if req.Body.Status == nil {
		return nil
	}

	// do something depends on contract status
	status := req.Body.Status.ToLower()
	switch status {
	case apistructs.ContractApproved:
		// the manager change the contract status to "approved", call the api-gateway to grant to the client
		if err := bdl.Bdl.GrantEndpointToClient(client.ClientID, access.EndpointID); err != nil {
			return err
		}
	case apistructs.ContractDisapproved:
		// do nothing with api-gateway
	case apistructs.ContractUnapproved:
		//  call the api-gateway to revoke the grant
		if err := bdl.Bdl.RevokeEndpointFromClient(client.ClientID, access.EndpointID); err != nil {
			return err
		}
	default:
		return errors.New("invalid contract status")
	}

	timeNow := time.Now()
	tx := dbclient.Tx()
	defer tx.RollbackUnlessCommitted()

	if err := tx.Model(contract).
		Where(map[string]interface{}{"org_id": req.OrgID, "id": req.URIParams.ContractID}).
		Updates(map[string]interface{}{"status": status, "updated_at": timeNow}).
		Error; err != nil {
		return errors.Wrap(err, "failed to Updates contractModel")
	}

	action := status2Action(status)
	if err := tx.Create(&apistructs.ContractRecordModel{
		ID:         0,
		OrgID:      req.OrgID,
		ContractID: contract.ID,
		Action:     fmt.Sprintf("%s???????????????????????????", action),
		CreatorID:  req.Identity.UserID,
		CreatedAt:  timeNow,
	}).Error; err != nil {
		return errors.Wrap(err, "failed to Create contractRecordModel")
	}

	tx.Commit()

	// notification by mail and in-site letter
	go svc.contractMsgToUser(req.OrgID, contract.CreatorID, access.AssetName, client, ApprovalResultFromStatus(status))

	return nil
}

// ??????????????????????????? SLA
func (svc *Service) updateContractCurSLA(req *apistructs.UpdateContractReq, contract *apistructs.ContractModel, client *apistructs.ClientModel,
	access *apistructs.APIAccessesModel) error {
	if req.Body.CurSLAID == nil {
		return nil
	}

	// ?????? SLA
	sla := unlimitedSLA(access)
	if *req.Body.CurSLAID != 0 {
		if err := svc.FirstRecord(sla, map[string]interface{}{"id": *req.Body.CurSLAID}); err != nil {
			return errors.Wrap(err, "failed to FirstRecord sla")
		}
	}

	// update contract
	var (
		timeNow = time.Now()
		updates = map[string]interface{}{
			"cur_sla_id":       *req.Body.CurSLAID,
			"updated_at":       timeNow,
			"updater_id":       req.Identity.UserID,
			"sla_committed_at": timeNow,
		}
		action = fmt.Sprintf("??? SLA ????????? %s", sla.Name)
	)

	tx := dbclient.Tx()
	defer tx.RollbackUnlessCommitted()

	if contract.RequestSLAID != nil && *contract.RequestSLAID == *req.Body.CurSLAID {
		updates["request_sla_id"] = nil
		action = fmt.Sprintf("????????????????????? %s ??? SLA ?????????", sla.Name)
	}

	if err := tx.Model(&contract).Updates(updates).Error; err != nil {
		return errors.Wrap(err, "failed to Updates contract")
	}

	// ??????????????????
	if err := tx.Create(&apistructs.ContractRecordModel{
		ID:         0,
		OrgID:      req.OrgID,
		ContractID: contract.ID,
		Action:     action,
		CreatorID:  req.Identity.UserID,
		CreatedAt:  timeNow,
	}).Error; err != nil {
		return errors.Wrap(err, "failed to Create contractRecordModel")
	}

	tx.Commit()

	go svc.contractMsgToUser(req.OrgID, contract.CreatorID, access.AssetName, client, ApprovalResultSLAUpdated(sla.Name))

	return nil
}

func (svc *Service) updateContractRequestSLA(req *apistructs.UpdateContractReq, contract *apistructs.ContractModel, access *apistructs.APIAccessesModel,
	asset *apistructs.APIAssetsModel) error {
	// ???????????????????????????????????????(????????????????????????????????????), ????????????
	if contract.RequestSLAID != nil {
		if *req.Body.RequestSLAID == *contract.RequestSLAID {
			return nil
		}
		if *req.Body.RequestSLAID == 0 {
			return errors.New("SLAID ??????")
		}
	}

	// ?????? SLA
	var (
		sla      apistructs.SLAModel
		whereSLA = map[string]interface{}{
			"id": *req.Body.RequestSLAID,
		}
		client      apistructs.ClientModel
		whereClient = map[string]interface{}{"id": contract.ClientID}
	)
	if err := svc.FirstRecord(&sla, whereSLA); err != nil {
		return errors.Wrap(err, "failed to FirstRecord sla")
	}
	if err := svc.FirstRecord(&client, whereClient); err != nil {
		return errors.Wrap(err, "failed to FirstRecord client")
	}

	var (
		timeNow       = time.Now()
		whereContract = map[string]interface{}{"id": contract.ID}
		updates       = map[string]interface{}{
			"request_sla_id": req.Body.RequestSLAID,
			"updated_at":     timeNow,
			"updater_id":     req.Identity.UserID,
		}
		record = apistructs.ContractRecordModel{
			ID:         0,
			OrgID:      req.OrgID,
			ContractID: contract.ID,
			Action:     fmt.Sprintf("?????????????????? %s ??? SLA", sla.Name),
			CreatorID:  req.Identity.UserID,
			CreatedAt:  timeNow,
		}
	)

	tx := dbclient.Tx()
	defer tx.RollbackUnlessCommitted()

	// ?????????????????????????????????
	if sla.Approval.ToLower() == apistructs.AuthorizationAuto {
		updates["cur_sla_id"] = *req.Body.RequestSLAID
		updates["sla_committed_at"] = timeNow
		updates["request_sla_id"] = nil
	}

	// ???????????????????????????(????????????????????????????????????)
	if contract.CurSLAID != nil && *contract.CurSLAID == *req.Body.RequestSLAID {
		updates["request_sla_id"] = nil
	}

	if err := tx.Model(&contract).
		Where(whereContract).
		Updates(updates).
		Error; err != nil {
		return errors.Wrap(err, "failed to Updates contract")
	}

	if err := tx.Create(&record).Error; err != nil {
		return errors.Wrap(err, "failed to Create record")
	}

	tx.Commit()

	go svc.contractMsgToManager(req.OrgID, req.Identity.UserID, asset, access, RequestItemSLA(sla.Name), updates["request_sla_id"] == nil)

	return nil
}

func (svc *Service) UpdateAccess(req *apistructs.UpdateAccessReq) (*apistructs.APIAccessesModel, *errorresp.APIError) {
	if req == nil || req.Body == nil {
		return nil, apierrors.UpdateAccess.InvalidParameter("invalid parameters")
	}

	var (
		asset  apistructs.APIAssetsModel
		access apistructs.APIAccessesModel
		where  = map[string]interface{}{
			"org_id": req.OrgID,
			"id":     req.URIParams.AccessID,
		}
		updates = map[string]interface{}{
			"minor":           req.Body.Minor,
			"workspace":       req.Body.Workspace,
			"authentication":  req.Body.Authentication,
			"authorization":   req.Body.Authorization,
			"bindDomain":      strings.Join(req.Body.BindDomain, ""),
			"addonInstanceID": req.Body.AddonInstanceID,
		}
	)

	// ???????????? access
	if err := svc.FirstRecord(&access, where); err != nil {
		logrus.Errorf("failed to FirstRecord access, where: %v, err: %v", where, err)
		if gorm.IsRecordNotFoundError(err) {
			return nil, apierrors.UpdateAccess.InternalError(errors.New("access not found"))
		}
		return nil, apierrors.UpdateAccess.InternalError(err)
	}

	// ??????????????? asset
	var whereFirstAsset = map[string]interface{}{
		"org_id":   req.OrgID,
		"asset_id": access.AssetID,
	}
	if err := svc.FirstRecord(&asset, whereFirstAsset); err != nil {
		logrus.Errorf("failed to FirstRecord asset, where: %v, err: %v", whereFirstAsset, err)
		return nil, apierrors.UpdateAccess.InternalError(errors.Errorf("???????????????API, API ??????: %s", access.AssetName))

	}

	// ?????? ?????????????????????????????? access ?????????
	rolesSet := bdl.FetchAssetRolesSet(req.OrgID, req.Identity.UserID)
	if written := req.Identity.UserID == access.CreatorID || writePermission(rolesSet, &asset); !written {
		return nil, apierrors.UpdateAccess.AccessDenied()
	}

	// ???????????? access ???????????????
	var (
		exInstantiation  apistructs.InstantiationModel
		whereFirstExInst = map[string]interface{}{
			"org_id":          req.OrgID,
			"asset_id":        access.AssetID,
			"swagger_version": access.SwaggerVersion,
			"minor":           access.Minor,
		}
	)
	if err := svc.FirstRecord(&exInstantiation, whereFirstExInst); err != nil {
		logrus.Errorf("failed to FirstRecord exInstantiation, where: %v, err: %v", whereFirstExInst, err)
		return nil, apierrors.UpdateAccess.InternalError(err)
	}

	// ?????? access ??? projectName
	if project, err := bdl.Bdl.GetProject(access.ProjectID); err == nil {
		updates["project_name"] = project.Name
	}

	// ???????????????????????????
	var (
		instantiations apistructs.InstantiationModel
		whereFirstInst = map[string]interface{}{
			"asset_id":        access.AssetID,
			"swagger_version": access.SwaggerVersion,
			"minor":           req.Body.Minor,
		}
	)
	if err := svc.FirstRecord(&instantiations, whereFirstInst); err != nil {
		logrus.Errorf("failed to FirstRecord instantiation, ")
		if gorm.IsRecordNotFoundError(err) {
			return nil, apierrors.UpdateAccess.InternalError(errors.New("no instantiation"))
		}
		return nil, apierrors.UpdateAccess.InternalError(err)
	}

	// ????????????????????? workspace ??????????????? (???????????????)
	if exInstantiation.Workspace != instantiations.Workspace {
		return nil, apierrors.UpdateAccess.InternalError(errors.Errorf("????????? %v.%v.* [??????:%s] ????????? %v.%v.* [??????:%s], ???????????????????????????",
			exInstantiation.Major, exInstantiation.Minor, exInstantiation.Workspace,
			instantiations.Major, exInstantiation.Minor, exInstantiation.Workspace))
	}

	if strings.ToLower(instantiations.Type) == InsTypeDice {
		// ??????????????????????????????????????????????????? (??????????????????)
		if instantiations.Workspace != req.Body.Workspace {
			return nil, apierrors.UpdateAccess.InternalError(errors.Errorf("????????????, ?????????????????????[%s]?????????????????????[%s]?????????",
				instantiations.Workspace, req.Body.Workspace))
		}
	}

	tx := dbclient.Tx()
	defer tx.RollbackUnlessCommitted()

	// ??????
	if err := tx.Model(&access).Where(where).Updates(updates).Error; err != nil {
		return nil, apierrors.UpdateInstantiation.InternalError(err)
	}

	// ??????????????? url
	var (
		urlStr     = instantiations.URL
		host, path string
	)
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") && !strings.HasPrefix(urlStr, "//") {
		urlStr = "//" + strings.TrimLeft(urlStr, "/")
	}
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, apierrors.UpdateAccess.InternalError(errors.Wrap(err, "invalid instance url"))
	}
	host = parsedURL.Host
	path = parsedURL.Path
	if path == "" {
		path = "/"
	}

	// ??????????????????
	authType := apistructs.AT_KEY_AUTH
	switch req.Body.Authentication.ToLower() {
	case apistructs.AuthenticationKeyAuth:
	case apistructs.AuthenticationSignAuth:
		authType = apistructs.AT_SIGN_AUTH
	case apistructs.AuthenticationOAuth2:
		return nil, apierrors.CreateAccess.InvalidParameter("???????????? OAuth2 ??????")
	default:
		return nil, apierrors.CreateAccess.InvalidParameter("????????????????????????")
	}
	if err := bdl.Bdl.UpdateEndpoint(access.EndpointID, apistructs.PackageDto{
		Name:             fmt.Sprintf("%s_%d_%s", access.AssetID, access.Major, access.Workspace),
		BindDomain:       req.Body.BindDomain,
		AuthType:         authType,
		AclType:          apistructs.ACL_ON,
		Scene:            apistructs.OPENAPI_SCENE,
		Description:      "creation of endpoint from apim",
		NeedBindCloudapi: false,
	}); err != nil {
		return nil, apierrors.UpdateAccess.InternalError(errors.Wrap(err, "failed to UpdateEndpoint"))
	}

	// ??????????????????
	if err := bdl.Bdl.CreateOrUpdateEndpointRootRoute(access.EndpointID, host, path); err != nil {
		return nil, apierrors.UpdateAccess.InternalError(errors.Wrap(err, "failed to UpdateEndpointRootRoute"))
	}

	tx.Commit()

	return &access, nil
}

func (svc *Service) UpdateSLA(req *apistructs.UpdateSLAReq) *errorresp.APIError {
	if req == nil || req.URIParams == nil || req.Body == nil {
		return apierrors.UpdateSLA.InvalidParameter("???????????????")
	}
	if len(req.Body.Limits) != 1 {
		return apierrors.UpdateSLA.InvalidParameter("???????????????????????????????????????")
	}
	limit := req.Body.Limits[0]
	if limit.Limit == 0 {
		return apierrors.UpdateSLA.InvalidParameter("??????????????? 0")
	}
	if !limit.Unit.Valid() {
		return apierrors.UpdateSLA.InvalidParameter("????????????????????????")
	}

	var (
		sla    apistructs.SLAModel
		asset  apistructs.APIAssetsModel
		access apistructs.APIAccessesModel
		where  = map[string]interface{}{
			"id": req.URIParams.SLAID,
		}
		updates = make(map[string]interface{})
		timeNow = time.Now()
	)
	// ?????? asset
	if err := svc.FirstRecord(&asset, map[string]interface{}{
		"org_id":   req.OrgID,
		"asset_id": req.URIParams.AssetID,
	}); err != nil {
		logrus.Errorf("failed to FirstRecord asset, err: %v", err)
		return apierrors.UpdateSLA.InternalError(errors.New("?????? API ??????"))
	}

	// SLA ?????????????????? API Asset ??? W ????????????
	rolesSet := bdl.FetchAssetRolesSet(req.OrgID, req.Identity.UserID)
	if written := writePermission(rolesSet, &asset); !written {
		return apierrors.UpdateSLA.AccessDenied()
	}

	// ?????? access
	if err := svc.FirstRecord(&access, map[string]interface{}{
		"org_id":          req.OrgID,
		"asset_id":        req.URIParams.AssetID,
		"swagger_version": req.URIParams.SwaggerVersion,
	}); err != nil {
		logrus.Errorf("failed to FirstRecord access, err: %v", err)
		return apierrors.UpdateSLA.InternalError(errors.New("????????????????????????"))
	}

	// ?????? SLA
	if err := svc.FirstRecord(&sla, where); err != nil {
		logrus.Errorf("failed to FirstRecord sla, err: %v", err)
		return apierrors.UpdateSLA.InternalError(errors.New("?????? SLA ??????"))
	}

	// ?????????????????????, ???????????????????????????????????? SLA
	if req.Body.Name != nil &&
		strings.Replace(*req.Body.Name, " ", "", -1) == strings.Replace(apistructs.UnlimitedSLAName, " ", "", -1) {
		return apierrors.CreateSLA.InternalError(errors.Errorf("??????????????? %s: ????????????", *req.Body.Name))
	}
	var (
		exName apistructs.SLAModel
	)
	if err := dbclient.Sq().Model(&exName).Where(map[string]interface{}{
		"access_id": access.ID,
		"name":      req.Body.Name,
	}).Where("id != ?", req.URIParams.SLAID).
		Find(&exName).Error; err == nil {
		return apierrors.UpdateSLA.InvalidParameter(errors.New("??????????????? SLA, ??????????????????"))
	}

	// ????????????????????????????????????, ????????????????????????????????????????????? SLA
	if req.Body.Approval.ToLower() == apistructs.AuthorizationAuto {
		var exAuto apistructs.SLAModel
		if err := dbclient.Sq().Model(&exAuto).Where(map[string]interface{}{
			"access_id": access.ID,
			"approval":  apistructs.AuthorizationAuto,
		}).Where("id != ?", req.URIParams.SLAID).
			Find(&exAuto).Error; err == nil {
			return apierrors.UpdateSLA.InvalidParameter(errors.Errorf("???????????????????????? SLA: %s, ??????????????????", exAuto.Name))
		}
	}

	tx := dbclient.Tx()
	defer tx.RollbackUnlessCommitted()

	// ??????????????????
	if req.Body.Name != nil {
		updates["name"] = *req.Body.Name
	}
	if req.Body.Desc != nil {
		updates["desc"] = *req.Body.Name
	}
	if req.Body.Approval != nil && req.Body.Approval.Valid() {
		updates["approval"] = *req.Body.Approval
	}
	if len(updates) > 0 {
		if err := tx.Model(new(apistructs.SLAModel)).
			Where(where).
			Updates(updates).
			Error; err != nil {
			logrus.Errorf("failed to Update SLAModel, err: %v", err)
			return apierrors.UpdateSLA.InternalError(errors.New("??????SLA??????"))
		}
	}

	// ?????? limit
	if err := tx.Delete(new(apistructs.SLALimitModel), map[string]interface{}{"sla_id": sla.ID}).Error; err != nil {
		logrus.Errorf("failed to Delete SLALimitModel, err: %v", err)
		return apierrors.DeleteSLA.InternalError(errors.New("??????????????????????????????"))
	}
	if err := tx.Create(&apistructs.SLALimitModel{
		BaseModel: apistructs.BaseModel{
			ID:        0,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
			CreatorID: req.Identity.UserID,
			UpdaterID: req.Identity.UserID,
		},
		SLAID: sla.ID,
		Limit: limit.Limit,
		Unit:  limit.Unit,
	}).Error; err != nil {
		logrus.Errorf("failed to Create limits, err: %v", err)
		return apierrors.UpdateSLA.InternalError(errors.New("????????????????????????"))
	}

	if len(updates) > 0 {
		switch approval := req.Body.Approval.ToLower(); {
		// ?????? SLA ?????????????????????????????????, ?????? access ????????? SLA ????????? SLA
		case sla.Approval.ToLower() == apistructs.AuthorizationManual && approval == apistructs.AuthorizationAuto:
			tx.Model(new(apistructs.APIAccessesModel)).
				Where(map[string]interface{}{"org_id": req.OrgID, "id": access.ID}).
				Updates(map[string]interface{}{"default_sla_id": sla.ID})

		// ??????????????? SLA ??? access ?????? SLA, ?????????????????????, ????????? access ????????? SLA ??????
		case access.DefaultSLAID != nil && *access.DefaultSLAID == sla.ID && approval == apistructs.AuthorizationManual:
			tx.Model(new(apistructs.APIAccessesModel)).
				Where(map[string]interface{}{"org_id": req.OrgID, "id": access.ID}).
				Updates(map[string]interface{}{"default_sla_id": nil})
		}
	}

	tx.Commit()

	// ?????? hepa ??????, ???????????????????????????
	go func() {
		// ?????????????????????????????????
		affectedClients, affectedContracts, err := svc.slaAffectClients(sla.ID)
		if err != nil {
			logrus.Warnf("failed to get slaAffectClients, err: %v", err)
			return
		}

		// ????????????
		sentContractIDs := make(map[uint64]bool)
		for _, affectedClient := range affectedClients {
			for _, affectedContract := range affectedContracts {
				if _, ok := sentContractIDs[affectedContract.ID]; ok {
					continue
				}
				svc.contractMsgToUser(req.OrgID, affectedContract.CreatorID, asset.AssetName, affectedClient, ManagerRewriteSLA(sla.Name))
				err = svc.createOrUpdateClientLimits(access.EndpointID, affectedClient.ClientID, affectedContract.ID)
				if err != nil {
					logrus.Errorf("createOrUpdateClientLimits failed, err:%+v", err)
				}
			}
		}
	}()

	return nil
}

// ?????? asset version
func (svc *Service) UpdateAssetVersion(req *apistructs.UpdateAssetVersionReq) (*apistructs.APIAssetVersionsModel, *errorresp.APIError) {
	if req == nil || req.URIParams == nil {
		return nil, apierrors.UpdateAssetVersion.InvalidParameter("???????????????")
	}
	if req.Body == nil {
		return nil, apierrors.UpdateAssetVersion.InvalidParameter("??????????????????")
	}

	// ??????
	if written := svc.writeAssetPermission(req.OrgID, req.Identity.UserID, req.URIParams.AssetID); !written {
		return nil, apierrors.UpdateAssetVersion.AccessDenied()
	}

	// ????????????
	var version apistructs.APIAssetVersionsModel
	if err := svc.FirstRecord(&version, map[string]interface{}{"id": req.URIParams.VersionID, "org_id": req.OrgID}); err != nil {
		logrus.Errorf("failed to FirstRecord version, err: %v", err)
		return nil, apierrors.UpdateAssetVersion.InternalError(errors.New("??????????????????"))
	}

	// ????????????
	updates := map[string]interface{}{
		"deprecated": req.Body.Deprecated,
		"updated_at": time.Now(),
		"updater_id": req.Identity.UserID,
	}
	if err := dbclient.Sq().Model(&version).Updates(updates).Error; err != nil {
		logrus.Errorf("failed to FirstRecord version, err: %v", err)
		return nil, apierrors.UpdateAssetVersion.InternalError(errors.New("????????????"))
	}

	// ????????????
	go svc.updateAssetVersionMsg(req, &version)

	return &version, nil
}

func (svc *Service) updateAssetVersionMsg(req *apistructs.UpdateAssetVersionReq, version *apistructs.APIAssetVersionsModel) {
	// ?????? minor latest version
	var latestVersion apistructs.APIAssetVersionsModel
	if err := dbclient.Sq().Where(map[string]interface{}{
		"org_id":   req.OrgID,
		"asset_id": req.URIParams.AssetID,
		"major":    version.Major,
		"minor":    version.Minor,
	}).Order("patch DESC").
		First(&latestVersion).
		Error; err != nil {
		logrus.Errorf("failed to first latestVersion, err: %v", err)
		return
	}
	// ????????????????????????????????????, ?????????????????????
	if latestVersion.ID != version.ID {
		logrus.Debug("needless to msg for version id")
		return
	}

	// ?????? access
	var access apistructs.APIAccessesModel
	if err := dbclient.Sq().Where(map[string]interface{}{
		"org_id":          req.OrgID,
		"asset_id":        req.URIParams.AssetID,
		"swagger_version": version.SwaggerVersion,
	}).First(&access).Error; err != nil {
		logrus.Warnf("failed to first access, %v", err)
		return
	}

	// ?????? access ????????? minor ?????????????????? version ??? minor ?????????, ???????????????????????????
	if access.Minor != version.Minor {
		return
	}

	// ????????????????????????
	var contracts []*apistructs.ContractModel
	if err := svc.ListRecords(&contracts, map[string]interface{}{
		"org_id":          req.OrgID,
		"asset_id":        req.URIParams.AssetID,
		"swagger_version": version.SwaggerVersion,
	}); err != nil {
		logrus.Warnf("failed to ListRecords contracts, err: %v", err)
		return
	}

	for _, contract := range contracts {
		var client apistructs.ClientModel
		if err := svc.FirstRecord(&client, map[string]interface{}{
			"org_id": req.OrgID,
			"id":     contract.ClientID,
		}); err != nil {
			logrus.Warnf("failed to FirstRecord, err: %v", err)
			continue
		}
		go svc.updateVersionMsgToUser(req.OrgID, contract.CreatorID, access.AssetName, version, &client)
	}
}
