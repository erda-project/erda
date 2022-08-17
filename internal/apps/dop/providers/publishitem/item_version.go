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

package publishitem

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/dop/publishitem/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/publishitem/db"
	"github.com/erda-project/erda/pkg/i18n"
	"github.com/erda-project/erda/pkg/strutil"
)

// PublishItemVersion publish item version
func (i *PublishItemService) PublishItemVersion(req *pb.CreatePublishItemVersionRequest) (*pb.PublishItemVersion, error) {
	metaJson := ""
	logo := ""
	resource := ""
	if req.ReleaseID != "" {
		release, err := i.bdl.GetRelease(req.ReleaseID)
		if err != nil {
			return nil, err
		}

		meta := map[string]interface{}{}
		meta["appId"] = release.ApplicationID
		meta["appName"] = release.ApplicationName
		meta["projectId"] = release.ProjectID
		meta["projectName"] = release.ProjectName
		meta["orgId"] = release.OrgID
		meta["releaseId"] = release.ReleaseID
		metaBytes, err := json.Marshal(meta)

		if err == nil {
			metaJson = string(metaBytes)
		}
		resourceByte, _ := json.Marshal(release.Resources)
		resource = string(resourceByte)
		logo = getLogo(resource)
	} else if req.AppID != 0 {
		app, err := i.bdl.GetApp(req.AppID)
		if err != nil {
			return nil, err
		}
		meta := map[string]interface{}{}
		meta["appId"] = app.ID
		meta["appName"] = app.Name
		meta["projectId"] = app.ProjectID
		meta["projectName"] = app.ProjectName
		meta["orgId"] = app.OrgID
		metaBytes, err := json.Marshal(meta)

		if err == nil {
			metaJson = string(metaBytes)
		}

	}
	_, err := i.db.GetPublishItem(req.PublishItemID)
	if err != nil {
		return nil, err
	}

	versionInfo := &pb.VersionInfo{
		PackageName: req.PackageName,
		Version:     req.Version,
		BuildID:     req.BuildID,
	}

	version, err := i.db.GetPublishItemVersionByName(req.OrgID, req.PublishItemID, apistructs.ResourceType(req.MobileType), versionInfo)
	if err == nil {
		version.Public = req.Public
		version.BuildID = req.BuildID
		version.PackageName = req.PackageName
		version.Desc = req.Desc
		version.Resources = resource
		version.Meta = metaJson
		version.Creator = req.Creator
		version.IsDefault = req.IsDefault
		version.Spec = req.Spec
		version.Readme = req.Readme
		version.Logo = logo
		version.MobileType = string(req.MobileType)
		version.Swagger = req.Swagger
		err := i.db.UpdatePublishItemVersion(version)
		if err != nil {
			return nil, err
		}
		if version.IsDefault {
			i.db.SetPublishItemVersionDefault(version.PublishItemID, int64(version.ID))
		}
		return version.ToApiData(), nil
	}

	itemVersion := db.PublishItemVersion{
		Version:       req.Version,
		BuildID:       req.BuildID,
		PackageName:   req.PackageName,
		Public:        req.Public,
		OrgID:         req.OrgID,
		Desc:          req.Desc,
		Logo:          logo,
		PublishItemID: req.PublishItemID,
		MobileType:    string(req.MobileType),
		IsDefault:     req.IsDefault,
		Resources:     resource,
		Creator:       req.Creator,
		Meta:          metaJson,
		Readme:        req.Readme,
		Spec:          req.Spec,
		Swagger:       req.Swagger,
	}
	err = i.db.CreatePublishItemVersion(&itemVersion)
	if err != nil {
		return nil, err
	}
	if itemVersion.IsDefault {
		i.db.SetPublishItemVersionDefault(itemVersion.PublishItemID, int64(itemVersion.ID))
	}

	// 创建H5的目标版本关系
	if req.MobileType == string(apistructs.ResourceTypeH5) {
		h5version, err := i.db.GetPublishItemVersionByName(req.OrgID, req.PublishItemID, apistructs.ResourceType(req.MobileType), versionInfo)
		if err != nil {
			return nil, err
		}
		for k, v := range req.H5VersionInfo.TargetMobiles {
			mobiles := v.GetListValue()
			for _, v1 := range mobiles.Values {
				if err := i.db.CreateH5Targets(&db.PublishItemH5Targets{
					H5VersionID:      h5version.ID,
					TargetVersion:    v1.String(),
					TargetMobileType: k,
				}); err != nil {
					return nil, err
				}
			}
		}
	}

	return itemVersion.ToApiData(), nil
}

func (s *PublishItemService) GetPublishItemDistribution(id int64, mobileType apistructs.ResourceType, packageName string,
	w http.ResponseWriter, r *http.Request) (*pb.PublishItemDistributionData, error) {
	publishItem, err := s.db.GetPublishItem(id)
	if err != nil {
		return nil, err
	}
	result := &pb.PublishItemDistributionData{
		Name:            publishItem.Name,
		DisplayName:     publishItem.DisplayName,
		Desc:            publishItem.Desc,
		Logo:            publishItem.Logo,
		CreatedAt:       timestamppb.New(publishItem.CreatedAt),
		PreviewImages:   strutil.SplitIfEmptyString(publishItem.PreviewImages, ","), // 预览图
		BackgroundImage: publishItem.BackgroundImage,                                // 背景图
	}
	if mobileType == "" {
		result.Versions = &pb.QueryPublishItemVersionData{List: []*pb.PublishItemVersion{}, Total: 0}
		return result, nil
	}
	versions, err := s.db.QueryPublishItemVersions(&pb.QueryPublishItemVersionRequest{
		Public:      "true",
		PageNo:      1,
		PageSize:    10,
		ItemID:      int64(publishItem.ID),
		MobileType:  string(mobileType),
		PackageName: packageName,
	})
	if err != nil {
		return nil, err
	}
	result.Versions = versions

	if publishItem.Type == apistructs.PublishItemTypeMobile {
		// 移动应用灰度分发
		err = s.GrayDistribution(w, r, *publishItem, result, mobileType, packageName)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// GetPublicPublishItemLaststVersion 获取线上最新版本的包
func (i *PublishItemService) GetPublicPublishItemLaststVersion(rw http.ResponseWriter, r *http.Request, req pb.GetPublishItemLatestVersionRequest) (*pb.GetPublishItemLatestVersionData, error) {
	resp, err := i.bdl.QueryAppPublishItemRelations(&apistructs.QueryAppPublishItemRelationRequest{
		AK: req.Ak,
		AI: req.Ai,
	})
	if err != nil {
		return nil, fmt.Errorf("get app publishItem relations fail: %v", err)
	}

	if len(resp.Data) != 1 {
		return nil, fmt.Errorf("invalid ak, ai: %s, %s", req.Ak, req.Ai)
	}

	publishItem, err := i.db.GetPublishItem(resp.Data[0].PublishItemID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, fmt.Errorf("invalid publishItemID: %v", resp.Data[0].PublishItemID)
		}
		return nil, err
	}

	// 获取当前的app版本
	currentAppVersion, err := i.db.GetPublishItemVersionByName(publishItem.OrgID, int64(publishItem.ID), apistructs.ResourceType(req.MobileType), req.CurrentAppInfo)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("can't find the app version: %v-%v-%v-%v", req.MobileType,
				req.CurrentAppInfo.PackageName, req.CurrentAppInfo.Version, req.CurrentAppInfo.BuildID)
		}
		return nil, err
	}

	// 获取灰度出来的最新app版本
	apps := []*db.PublishItemVersion{currentAppVersion}
	publishItemDistribution, err := i.GetPublishItemDistribution(int64(publishItem.ID), apistructs.ResourceType(req.MobileType), "",
		rw, r)
	if err != nil {
		return nil, err
	}

	if publishItemDistribution.Default != nil {
		distributionVersion, err := i.db.GetPublishItemVersion(int64(publishItemDistribution.Default.ID))
		if err != nil {
			return nil, err
		}
		apps = append(apps, distributionVersion)
	}
	newestAppVersion := getNewestVersion(apps...)
	// 服务端校验版本是否需要更新，如果当前版本已经是最新了，返回空版本即可
	if req.Check && newestAppVersion.ID == currentAppVersion.ID {
		// 校验app版本
		newestAppVersion = nil
	}

	newestH5Versions := make(map[string]*pb.PublishItemVersion, 0)
	for _, v := range req.CurrentH5Info {
		// 当前H5版本
		currentH5Version, err := i.db.GetPublishItemVersionByName(publishItem.OrgID, int64(publishItem.ID), apistructs.ResourceTypeH5, v)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("can't find the h5 version: %v-%v-%v-%v", req.MobileType, v.PackageName, v.Version, v.BuildID)
			}
			return nil, err
		}

		// 强制下载预览版H5包
		if req.ForceBetaH5 {
			versions, err := i.GetPublicPublishItemVersionImpl(int64(publishItem.ID), string(apistructs.ResourceTypeH5), v.PackageName)
			if err != nil {
				return nil, err
			}
			if len(versions.List) < 2 {
				goto LABEL1
			}
			if req.Check && versions.List[1].ID == currentH5Version.ID {
				// 校验h5版本
				newestH5Versions[v.PackageName] = nil
			} else {
				newestH5Versions[v.PackageName] = versions.List[1]
			}
			continue
		}

		// 正常最新版本逻辑
	LABEL1:
		// 获取灰度出来的最新H5版本
		h5s := []*db.PublishItemVersion{currentH5Version}
		publishItemDistributionH5, err := i.GetPublishItemDistribution(int64(publishItem.ID), apistructs.ResourceTypeH5, v.PackageName,
			rw.(http.ResponseWriter), r)
		if err != nil {
			return nil, err
		}
		if publishItemDistributionH5.Default != nil {
			distributionVersionH5, err := i.db.GetPublishItemVersion(int64(publishItemDistributionH5.Default.ID))
			if err != nil {
				return nil, err
			}
			h5s = append(h5s, distributionVersionH5)
		}
		newestH5Version := getNewestVersion(h5s...)
		// 服务端校验版本是否需要更新，如果当前版本已经是最新了，返回空版本即可
		if req.Check && newestH5Version.ID == currentH5Version.ID {
			// 校验h5版本
			newestH5Version = nil
		}
		newestH5Versions[v.PackageName] = newestH5Version
	}

	return &pb.GetPublishItemLatestVersionData{
		AppVersion: newestAppVersion,
		H5Versions: newestH5Versions,
	}, nil
}

func (s *PublishItemService) GetPublicPublishItemVersionImpl(itemID int64, mobileType, packageName string) (*pb.QueryPublishItemVersionData, error) {
	total, tmpVersions, err := s.db.GetPublicVersion(itemID, apistructs.ResourceType(mobileType), packageName)
	if err != nil {
		return nil, err
	}

	versions, err := discriminateReleaseAndBeta(total, tmpVersions)
	if err != nil {
		return nil, err
	}

	var result []*pb.PublishItemVersion
	for _, v := range versions {
		result = append(result, v.ToApiData())
	}

	return &pb.QueryPublishItemVersionData{
		Total: int64(total),
		List:  result,
	}, nil
}

// SetPublishItemVersionDefault 设置发布版本默认状态
func (s *PublishItemService) SetPublishItemVersionDefault(itemVersionID, itemID int64) error {
	return s.db.SetPublishItemVersionDefault(itemID, itemVersionID)
}

// SetPublishItemVersionPublic 设置发布版本为公开
func (s *PublishItemService) SetPublishItemVersionPublic(id, itemID int64) error {
	return s.db.SetPublishItemVersionPublic(id, itemID)
}

// SetPublishItemVersionUnPublic 设置发布版本为公开
func (s *PublishItemService) SetPublishItemVersionUnPublic(id, itemID int64) error {
	return s.db.SetPublishItemVersionUnPublic(id, itemID)
}

func (s *PublishItemService) PublicPublishItemVersion(req *pb.UpdatePublishItemVersionStatesRequset,
	local *i18n.LocaleResource) error {
	v, err := s.db.GetPublishItemVersion(req.PublishItemVersionID)
	if err != nil {
		return err
	}

	// 在发布版本的时候，ios和android暂时不区分包名，即认为一个发布内容下的ios和android版本没有区别
	if v.MobileType == "ios" || v.MobileType == "android" {
		req.PackageName = ""
	}

	total, tmpVersions, err := s.db.GetPublicVersion(req.PublishItemID, apistructs.ResourceType(v.MobileType), req.PackageName)
	if err != nil {
		return err
	}

	versions, err := discriminateReleaseAndBeta(total, tmpVersions)
	if err != nil {
		return err
	}

	switch req.VersionStates {
	case string(apistructs.PublishItemReleaseVersion):
		if req.Public {
			// 上架正式版
			err = s.publicReleaseVersion(total, versions, req, local)
		} else {
			// 下架正式版
			err = s.unPublicReleaseVersion(total, versions, req, local)
		}
	case string(apistructs.PublishItemBetaVersion):
		if req.Public {
			// 上架beta版
			err = s.publicBetaVersion(total, versions, req, local)
		} else {
			err = s.unPublicBetaVersion(total, versions, req, local)
		}
	default:
		err = fmt.Errorf("unknow version state: %v", req.VersionStates)
	}

	if err != nil {
		return err
	}

	return nil
}

func getNewestVersion(versions ...*db.PublishItemVersion) *pb.PublishItemVersion {
	if len(versions) == 1 {
		return versions[0].ToApiData()
	}
	tmpV := versions[0]
	for _, v := range versions[1:] {
		if !tmpV.IsLater(v) {
			tmpV = v
		}
	}

	return tmpV.ToApiData()
}

func getLogo(resourcedStr string) string {
	var resources []apistructs.ReleaseResource
	err := json.Unmarshal([]byte(resourcedStr), &resources)
	if err != nil {
		return ""
	}
	for _, resource := range resources {
		if resource.Type == apistructs.ResourceTypeAndroid || resource.Type == apistructs.ResourceTypeIOS {
			logo, ok := resource.Meta["logo"]
			if ok {
				return logo.(string)
			}
		}
	}
	return ""
}
