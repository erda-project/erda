package assetsvc

import (
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/apim/bdl"
)

func inSlice(s string, ss []string) bool {
	for _, v := range ss {
		if s == v {
			return true
		}
	}
	return false
}

// asset 创建者, 企业管理人员, 项目管理人员, 应用管理人员拥有写权限, 返回 true
func writePermission(rolesSet *bdl.RolesSet, assetModel *apistructs.APIAssetsModel) bool {
	if rolesSet.UserID() == assetModel.CreatorID {
		return true
	}

	var (
		orgs = rolesSet.RolesOrgs(bdl.OrgMRoles...)
		pros = rolesSet.RolesProjects(bdl.ProMRoles...)
		apps = rolesSet.RolesApps(bdl.AppMRoles...)
	)
	if inSlice(strconv.FormatUint(rolesSet.OrgID(), 10), orgs) {
		return true
	}
	if assetModel.ProjectID != nil && inSlice(strconv.FormatUint(*assetModel.ProjectID, 10), pros) {
		return true
	}
	if assetModel.AppID != nil && inSlice(strconv.FormatUint(*assetModel.AppID, 10), apps) {
		return true
	}

	return false
}

func (svc *Service) writeAssetPermission(orgID uint64, userID string, assetID string) (written bool) {
	// 查出 asset
	var asset apistructs.APIAssetsModel
	if err := svc.FirstRecord(&asset, map[string]interface{}{
		"org_id":   orgID,
		"asset_id": assetID,
	}); err != nil {
		logrus.Errorf("failed to FirstRecord asset, err: %v", err)
		return false
	}

	rolesSet := bdl.FetchRolesSet(orgID, userID)
	return writePermission(rolesSet, &asset)
}
