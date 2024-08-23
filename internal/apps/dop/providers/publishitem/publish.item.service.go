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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/dop/publishitem/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/publishitem/db"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type PublishItemService struct {
	p *provider

	db  *db.DBClient
	bdl *bundle.Bundle
}

func (s *PublishItemService) CreatePublishItemBlackList(ctx context.Context, req *pb.PublishItemUserlistRequest) (*pb.PublishItemAddBlacklistResponse, error) {
	req.UserID = apis.GetUserID(ctx)
	publishItem, err := s.addblackList(req)
	if err != nil {
		return nil, apierrors.ErrCreateBlacklist.InternalError(err)
	}
	return &pb.PublishItemAddBlacklistResponse{
		Data: publishItem.ToApiData(),
	}, nil
}

func (s *PublishItemService) DeletePublishItemBlackList(ctx context.Context, req *pb.DeletePublishItemBlackListRequest) (*pb.PublishItemDeleteBlacklistResponse, error) {
	publishItem, err := s.checkPublishItemExsit(int64(req.PublishItemId))
	if err != nil {
		return nil, apierrors.ErrDeleteBlacklist.InternalError(err)
	}
	blackList, err := s.db.GetBlacklistByID(req.BlacklistId)
	if err != nil {
		return nil, apierrors.ErrDeleteBlacklist.InternalError(err)
	}
	if blackList == nil {
		return &pb.PublishItemDeleteBlacklistResponse{
			Data: nil,
		}, nil
	}
	if err := s.db.DeleteBlacklist(blackList); err != nil {
		return nil, apierrors.ErrDeleteBlacklist.InternalError(err)
	}
	return &pb.PublishItemDeleteBlacklistResponse{
		Data: &pb.PublishItemUserListResponse{
			ID:              blackList.ID,
			UserID:          blackList.UserID,
			UserName:        blackList.UserName,
			DeviceNo:        blackList.DeviceNo,
			CreatedAt:       timestamppb.New(blackList.CreatedAt),
			PublishItemName: publishItem.Name,
		},
	}, nil
}

func (s *PublishItemService) GetPublishItemBlackList(ctx context.Context, req *pb.GetPublishItemBlackListRequest) (*pb.PublishItemUserListDataResponse, error) {
	queryReq := &pb.PublishItemUserlistRequest{
		PageSize:      req.PageSize,
		PageNo:        req.PageNo,
		PublishItemID: uint64(req.PublishItemId),
	}
	artifact, err := s.getBlackLists(queryReq)
	if err != nil {
		return nil, apierrors.ErrGetBlacklist.InternalError(err)
	}
	return &pb.PublishItemUserListDataResponse{
		Data: artifact,
	}, nil
}

func (s *PublishItemService) CreatePublishItemErase(ctx context.Context, req *pb.CreatePublishItemEraseRequest) (*pb.PublicItemAddEraseResponse, error) {
	req.Operator = apis.GetUserID(ctx)
	publishItem, err := s.addErase(&pb.PublishItemUserlistRequest{
		PublishItemID: req.PublishItemId,
		Operator:      req.Operator,
		PageNo:        req.PageNo,
		PageSize:      req.PageSize,
		UserID:        req.UserID,
		UserName:      req.UserName,
		DeviceNo:      req.DeviceNo,
	})
	if err != nil {
		return nil, apierrors.ErrCreateEraselist.InternalError(err)
	}
	return &pb.PublicItemAddEraseResponse{
		Data: &pb.PublicItemAddEraseData{
			Data:     publishItem.ToApiData(),
			DeviceNo: req.DeviceNo,
		},
	}, nil
}

func (s *PublishItemService) GetPublishItemErase(ctx context.Context, req *pb.GetPublishItemEraseRequest) (*pb.PublishItemUserListDataResponse, error) {
	_, err := s.checkPublishItemExsit(int64(req.PublishItemId))
	if err != nil {
		return nil, apierrors.ErrGetPublishItem.InternalError(err)
	}
	total, blackList, err := s.db.GetErases(req.PageNo, req.PageSize, req.PublishItemId)
	if err != nil {
		return nil, apierrors.ErrGetPublishItem.InternalError(err)
	}
	resultData := &pb.PublishItemUserListData{}
	if len(*blackList) == 0 {
		return &pb.PublishItemUserListDataResponse{
			Data: resultData,
		}, nil
	}
	resultData.Total = total
	for _, v := range *blackList {
		resultData.List = append(resultData.List, &pb.PublishItemUserListResponse{
			EraseStatus: v.EraseStatus,
			DeviceNo:    v.DeviceNo,
			CreatedAt:   timestamppb.New(v.UpdatedAt),
		})
	}
	return &pb.PublishItemUserListDataResponse{
		Data: resultData,
	}, nil
}
func (s *PublishItemService) ListPublishItemMonitorKeys(ctx context.Context, req *pb.ListPublishItemMonitorKeysRequest) (*pb.ListPublishItemMonitorKeysResponse, error) {
	publishItem, err := s.db.GetPublishItem(int64(req.PublishItemId))
	if err != nil {
		return nil, apierrors.ErrGetMonitorKeys.InternalError(err)
	}
	mks := []*pb.MonitorKeys{{
		AK:    publishItem.AK,
		AI:    publishItem.AI,
		Env:   "OFFLINE",
		AppID: 0,
	}}
	resp, err := s.bdl.QueryAppPublishItemRelations(&apistructs.QueryAppPublishItemRelationRequest{
		PublishItemID: int64(req.PublishItemId),
	})
	if err != nil {
		return nil, apierrors.ErrGetMonitorKeys.InternalError(err)
	}

	for _, relation := range resp.Data {
		if relation.AK == "" || relation.AI == "" {
			continue
		}
		mks = append(mks, &pb.MonitorKeys{
			AK:    relation.AK,
			AI:    relation.AI,
			Env:   relation.Env,
			AppID: relation.AppID,
		})
	}
	mksMap := make(map[string]*pb.MonitorKeys, 0)
	for _, relation := range mks {
		mksMap[relation.Env+"-"+relation.AI] = relation
	}
	return &pb.ListPublishItemMonitorKeysResponse{
		Data: mksMap,
	}, nil
}

func (s *PublishItemService) CreatePublishItem(ctx context.Context, req *pb.CreatePublishItemRequest) (*pb.CreatePublishItemResponse, error) {
	orgID, err := s.getOrgIDFromContext(ctx)
	if err != nil {
		return nil, apierrors.ErrCreatePublishItem.NotLogin()
	}
	req.OrgID = orgID
	req.Creator = apis.GetUserID(ctx)

	if req.Name == "" {
		return nil, apierrors.ErrCreatePublishItem.InvalidParameter("name is null")
	}
	if len(req.Name) > 50 {
		return nil, apierrors.ErrCreatePublishItem.InvalidParameter("name too long,limit 50")
	}
	queryNameResult, err := s.db.QueryPublishItem(&pb.QueryPublishItemRequest{
		PageNo:      0,
		PageSize:    0,
		PublisherId: req.PublisherID,
		Name:        req.Name,
		OrgID:       req.OrgID,
	})
	if err != nil {
		return nil, apierrors.ErrCreatePublishItem.InternalError(err)
	}
	if queryNameResult.Total > 0 {
		return nil, apierrors.ErrCreatePublishItem.InvalidParameter("name already exist")
	}
	publishItem := db.PublishItem{
		Name:             req.Name,
		PublisherID:      req.PublisherID,
		Type:             req.Type,
		Logo:             req.Logo,
		Public:           req.Public,
		DisplayName:      req.DisplayName,
		OrgID:            req.OrgID,
		Desc:             req.Desc,
		Creator:          req.Creator,
		AK:               s.db.GeneratePublishItemKey(),
		AI:               req.Name,
		NoJailbreak:      req.NoJailbreak,
		GeofenceLon:      req.GeofenceLon,
		GeofenceLat:      req.GeofenceLat,
		GeofenceRadius:   req.GeofenceRadius,
		GrayLevelPercent: int(req.GrayLevelPercent),
		PreviewImages:    strings.Join(req.PreviewImages, ","),
		BackgroundImage:  req.BackgroundImage,
	}
	err = s.db.CreatePublishItem(&publishItem)
	if err != nil {
		return nil, err
	}
	return &pb.CreatePublishItemResponse{
		Data: publishItem.ToApiData(),
	}, nil
}

func (s *PublishItemService) DeletePublishItem(ctx context.Context, req *pb.DeletePublishItemRequest) (*pb.DeletePublishItemResponse, error) {
	item, err := s.db.GetPublishItem(int64(req.PublishItemId))
	if err != nil {
		return nil, apierrors.ErrDeletePublishItem.InternalError(err)
	}

	if err := s.bdl.RemoveAppPublishItemRelations(int64(item.ID)); err != nil {
		return nil, apierrors.ErrDeletePublishItem.InternalError(err)
	}
	if err := s.db.DeletePublishItemVersionsByItemID(int64(item.ID)); err != nil {
		return nil, apierrors.ErrDeletePublishItem.InternalError(err)
	}
	if err := s.db.Delete(item).Error; err != nil {
		return nil, apierrors.ErrDeletePublishItem.InternalError(err)
	}
	return &pb.DeletePublishItemResponse{
		Data: item.ToApiData(),
	}, nil
}

func (s *PublishItemService) GetPublishItem(ctx context.Context, req *pb.GetPublishItemRequest) (*pb.GetPublishItemResponse, error) {
	item, err := s.GetPublishItemImpl(req.Id)
	if err != nil {
		return nil, apierrors.ErrGetPublishItem.InternalError(err)
	}
	return &pb.GetPublishItemResponse{
		Data: item,
	}, nil
}

func (s *PublishItemService) GetPublishItemImpl(publishItemID int64) (*pb.PublishItem, error) {
	publishItem, err := s.db.GetPublishItem(publishItemID)
	if err != nil {
		return nil, err
	}
	result := publishItem.ToApiData()
	result.DownloadUrl = fmt.Sprintf("%s/download/%d", s.p.Cfg.SiteUrl, result.ID)
	return result, nil
}

func (s *PublishItemService) GetPublishItemH5PackageName(ctx context.Context, req *pb.GetPublishItemH5PackageNameRequest) (*pb.GetPublishItemH5PackageNameResponse, error) {
	h5Versions, err := s.db.GetH5VersionByItemID(req.PublishItemId)
	if err != nil {
		return nil, apierrors.ErrGetPublishItem.InternalError(err)
	}
	result := make([]string, 0, len(h5Versions))
	for _, v := range h5Versions {
		result = append(result, v.PackageName)
	}
	return &pb.GetPublishItemH5PackageNameResponse{
		Data: result,
	}, nil
}

func (s *PublishItemService) QueryPublishItem(ctx context.Context, req *pb.QueryPublishItemRequest) (*pb.QueryPublishItemResponse, error) {
	orgID, err := s.getOrgIDFromContext(ctx)
	if err != nil {
		return nil, apierrors.ErrQueryPublishItem.NotLogin()
	}
	internalClient := apis.GetInternalClient(ctx)
	req.OrgID = orgID
	queryPublishItemResult, err := s.db.QueryPublishItem(req)
	if err != nil {
		return nil, err
	}
	for _, item := range queryPublishItemResult.List {
		item.DownloadUrl = fmt.Sprintf("%s/download/%d", s.p.Cfg.SiteUrl, item.ID)
		if item.Type == string(apistructs.ApplicationModeLibrary) {
			versions, err := s.QueryPublishItemVersions(ctx, &pb.QueryPublishItemVersionRequest{
				Public:   "true",
				ItemID:   item.ID,
				OrgID:    item.OrgID,
				PageSize: 1,
			})
			if err == nil && len(versions.List) > 0 {
				item.LatestVersion = versions.List[0].Version
			}
		}
		// if it is not an internal call, desensitize the sensitive information
		if internalClient == "" {
			item.AK = ""
			item.AI = ""
		}
	}
	return &pb.QueryPublishItemResponse{Data: queryPublishItemResult}, nil
}

func (s *PublishItemService) UpdatePublishItem(ctx context.Context, req *pb.UpdatePublishItemRequest) (*pb.UpdatePublishItemResponse, error) {
	req.ID = req.PublishItemId
	if err := s.updatePublishItemImpl(req); err != nil {
		return nil, apierrors.ErrUpdatePublishItem.InternalError(err)
	}
	item, err := s.GetPublishItem(ctx, &pb.GetPublishItemRequest{
		Id: req.PublishItemId,
	})
	if err != nil {
		return nil, apierrors.ErrUpdatePublishItem.InternalError(err)
	}
	return &pb.UpdatePublishItemResponse{Data: item.Data}, nil
}

func (s *PublishItemService) CreatePublishItemVersion(ctx context.Context, req *pb.CreatePublishItemVersionRequest) (*pb.CreatePublishItemVersionResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, apierrors.ErrCreatePublishItemVersion.NotLogin()
	}
	userID := apis.GetUserID(ctx)
	req.OrgID = orgID
	req.Creator = userID

	result, err := s.PublishItemVersion(req)
	if err != nil {
		return nil, apierrors.ErrCreatePublishItemVersion.InternalError(err)
	}

	item, err := s.GetPublishItemImpl(req.PublishItemID)
	if err != nil {
		return nil, apierrors.ErrCreatePublishItemVersion.InternalError(err)
	}

	return &pb.CreatePublishItemVersionResponse{
		Data: &pb.CreatePublishItemVersionData{
			PublishItem: item,
			Data:        result,
		},
	}, nil
}

func (s *PublishItemService) GetPublicPublishItemVersion(ctx context.Context, req *pb.GetPublicPublishItemVersionRequest) (*pb.QueryPublishItemVersionResponse, error) {
	results, err := s.GetPublicPublishItemVersionImpl(req.PublishItemId, req.MobileType, req.PackageName)
	if err != nil {
		return nil, apierrors.ErrGetPublishItem.InternalError(err)
	}
	return &pb.QueryPublishItemVersionResponse{
		Data: results,
	}, nil
}

func (s *PublishItemService) QueryPublishItemVersion(ctx context.Context, req *pb.QueryPublishItemVersionRequest) (*pb.QueryPublishItemVersionResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, apierrors.ErrQueryPublishItemVersion.NotLogin()
	}
	req.OrgID = orgID
	req.ItemID = req.PublishItemId
	result, err := s.QueryPublishItemVersions(ctx, req)
	if err != nil {
		return nil, apierrors.ErrQueryPublishItemVersion.InternalError(err)
	}
	return &pb.QueryPublishItemVersionResponse{
		Data: result,
	}, nil
}

func (s *PublishItemService) SetPublishItemVersionStatus(ctx context.Context, req *pb.SetPublishItemVersionStatusRequest) (*emptypb.Empty, error) {
	var err error
	if req.Action == "public" {
		err = s.SetPublishItemVersionPublic(req.VersionID, req.PublishItemId)
	} else if req.Action == "unpublic" {
		err = s.SetPublishItemVersionUnPublic(req.VersionID, req.PublishItemId)
	} else if req.Action == "default" {
		err = s.SetPublishItemVersionDefault(req.VersionID, req.PublishItemId)
	} else {
		return nil, apierrors.ErrSetPublishItemVersionStatus.InvalidParameter("action")
	}

	if err != nil {
		return nil, apierrors.ErrSetPublishItemVersionStatus.InternalError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PublishItemService) UpdatePublishItemVersion(ctx context.Context, req *pb.UpdatePublishItemVersionStatesRequset) (*emptypb.Empty, error) {
	if req.Action == "publish" {
		req.Public = true
	} else if req.Action == "unpublish" {
		req.Public = false
	} else {
		return nil, apierrors.ErrUpdatePublishItemVersion.InvalidParameter("action")
	}
	if err := s.PublicPublishItemVersion(req, s.bdl.GetLocale(apis.GetLang(ctx))); err != nil {
		return nil, apierrors.ErrUpdatePublishItemVersion.InternalError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PublishItemService) QueryMyPublishItem(ctx context.Context, req *pb.QueryPublishItemRequest) (*pb.QueryPublishItemResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, apierrors.ErrQueryPublishItem.NotLogin()
	}
	publishers, err := s.bdl.GetUserRelationPublisher(apis.GetUserID(ctx), apis.GetOrgID(ctx))
	if err != nil {
		return nil, apierrors.ErrQueryPublishItem.InternalError(err)
	}
	if publishers.Total == 0 || len(publishers.List) == 0 {
		return nil, apierrors.ErrQueryPublishItem.InternalError(errors.New("no publisher"))
	}
	req.OrgID = orgID
	req.PublisherId = int64(publishers.List[0].ID)
	res, err := s.QueryPublishItem(apis.WithInternalClientContext(ctx, discover.DOP()), req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *PublishItemService) QueryPublishItemVersions(ctx context.Context, req *pb.QueryPublishItemVersionRequest) (*pb.QueryPublishItemVersionData, error) {
	itemversions, err := s.db.QueryPublishItemVersions(req)
	if err != nil {
		return nil, err
	}
	return itemversions, nil
}

func (s *PublishItemService) getBlackLists(req *pb.PublishItemUserlistRequest) (*pb.PublishItemUserListData, error) {
	_, err := s.checkPublishItemExsit(int64(req.PublishItemID))
	if err != nil {
		return nil, err
	}
	total, blackLists, err := s.db.GetBlacklists(req.PageNo, req.PageSize, req.PublishItemID)
	if err != nil {
		return nil, err
	}
	resultData := &pb.PublishItemUserListData{}
	if len(*blackLists) == 0 {
		return resultData, nil
	}
	resultData.Total = total
	for _, v := range *blackLists {
		resultData.List = append(resultData.List, &pb.PublishItemUserListResponse{
			ID:        v.ID,
			UserID:    v.UserID,
			UserName:  v.UserName,
			DeviceNo:  v.DeviceNo,
			CreatedAt: timestamppb.New(v.CreatedAt),
		})
	}
	return resultData, nil
}

func (s *PublishItemService) addblackList(req *pb.PublishItemUserlistRequest) (*db.PublishItem, error) {
	artifact, err := s.checkPublishItemExsit(int64(req.PublishItemID))
	if err != nil {
		return nil, err
	}
	if req.UserID != "" {
		blackList, err := s.db.GetBlacklistByUserID(req.UserID, req.PublishItemID)
		if err != nil {
			return nil, err
		}
		if blackList != nil {
			for _, v := range blackList {
				if v.DeviceNo == req.DeviceNo {
					return nil, errors.New("请勿重复添加")
				}
			}
		}
	}
	if req.DeviceNo != "" {
		blackList, err := s.db.GetBlacklistByDeviceNo(req.PublishItemID, req.DeviceNo)
		if err != nil {
			return nil, err
		}
		if blackList != nil {
			for _, v := range blackList {
				if v.UserID == req.UserID {
					return nil, errors.New("请勿重复添加")
				}
			}
		}
	}
	if err := s.db.CreateBlacklist(&db.PublishItemBlackList{
		PublishItemID:  req.PublishItemID,
		PublishItemKey: artifact.AK,
		DeviceNo:       req.DeviceNo,
		UserID:         req.UserID,
		UserName:       req.UserName,
		Operator:       req.Operator,
	}); err != nil {
		return nil, err
	}
	return artifact, nil
}

func (s *PublishItemService) addErase(req *pb.PublishItemUserlistRequest) (*db.PublishItem, error) {
	artifact, err := s.checkPublishItemExsit(int64(req.PublishItemID))
	if err != nil {
		return nil, err
	}
	erase, err := s.db.GetEraseByDeviceNo(req.PublishItemID, req.DeviceNo)
	if err != nil {
		return nil, err
	}
	if erase == nil {
		if err := s.db.CreateErase(&db.PublishItemErase{
			PublishItemID:  req.PublishItemID,
			PublishItemKey: artifact.AK,
			DeviceNo:       req.DeviceNo,
			EraseStatus:    apistructs.Erasing,
			Operator:       req.Operator,
		}); err != nil {
			return nil, err
		}
		return artifact, nil
	}
	if erase.EraseStatus == apistructs.Erasing {
		return nil, fmt.Errorf("do not add repeatedly")
	}
	erase.EraseStatus = apistructs.Erasing
	if err := s.db.UpdateErase(erase); err != nil {
		return nil, err
	}
	return artifact, nil
}

type PKGInfo struct {
	PackageName string
	Version     string
	BuildID     string
	DisplayName string
	Logo        string
}

func (s *PublishItemService) CreateOffLineVersion(param apistructs.CreateOffLinePublishItemVersionRequest) (string, error) {
	fileHeader := param.FileHeader
	// 校验publishItem是否存在
	_, err := s.db.GetPublishItem(param.PublishItemID)
	if err != nil {
		return "", err
	}

	var (
		resourceType apistructs.ResourceType
		meta         map[string]interface{}
		pkgInfo      *PKGInfo
		logoURL      string
	)

	// 校验文件后缀
	ext := filepath.Ext(fileHeader.Filename)
	if ext == ".apk" {
		resourceType = apistructs.ResourceTypeAndroid
	} else if ext == ".ipa" {
		resourceType = apistructs.ResourceTypeIOS
	} else {
		return "", fmt.Errorf("unknow file type: %s", ext)
	}

	// 上传移动应用文件
	mobileFileUploadResult, err := s.UploadFileFromReader(fileHeader)
	if err != nil {
		return "", err
	}

	switch resourceType {
	case apistructs.ResourceTypeAndroid:
		info, err := GetAndoridInfo(fileHeader)
		if err != nil {
			return "", err
		}
		if info.Icon != nil {
			logoTmpPath := GenrateTmpImagePath(time.Now().String())
			if err := SaveImageToFile(info.Icon, logoTmpPath); err != nil {
				return "", fmt.Errorf("error encode jpeg icon %s %v", logoTmpPath, err)
			}
			defer func() {
				if err := os.Remove(logoTmpPath); err != nil {
					logrus.Errorf("remove logoFile %s err: %v", logoTmpPath, err)
				}
			}()
			logoUploadResult, err := s.UploadFileFromFile(logoTmpPath)
			if err != nil {
				return "", err
			}
			logoURL = logoUploadResult.DownloadURL
		}
		versionCodeStr := strconv.FormatInt(int64(info.VersionCode), 10)
		pkgInfo = &PKGInfo{
			PackageName: info.PackageName,
			Version:     info.Version,
			BuildID:     versionCodeStr,
			DisplayName: info.Version,
			Logo:        logoURL,
		}
		meta = map[string]interface{}{"packageName": info.PackageName, "version": info.Version, "buildID": pkgInfo.BuildID,
			"displayName": info.Version, "logo": logoURL}
	case apistructs.ResourceTypeIOS:
		info, err := GetIosInfo(fileHeader)
		if err != nil {
			return "", err
		}
		if info.Icon != nil {
			logoTmpPath := GenrateTmpImagePath(time.Now().String())
			err = SaveImageToFile(info.Icon, logoTmpPath)
			if err != nil {
				return "", fmt.Errorf("error encode jpeg icon %s %v", logoTmpPath, err)
			}
			defer func() {
				if err := os.Remove(logoTmpPath); err != nil {
					logrus.Errorf("remove logoFile %s err: %v", logoTmpPath, err)
				}
			}()
			logoUploadResult, err := s.UploadFileFromFile(logoTmpPath)
			if err != nil {
				return "", err
			}
			logoURL = logoUploadResult.DownloadURL
		}
		installPlistContent := GenerateInstallPlist(info, mobileFileUploadResult.DownloadURL)
		t := time.Now().Format("20060102150405")
		if err := os.Mkdir("/tmp/"+t, os.ModePerm); err != nil {
			return "", err
		}
		plistFile := "/tmp/" + t + "/install.plist"
		if err = os.WriteFile(plistFile, []byte(installPlistContent), os.ModePerm); err != nil {
			return "", err
		}
		defer func() {
			if err := os.RemoveAll("/tmp/" + t); err != nil {
				logrus.Errorf("remove plistFile %s err: %v", plistFile, err)
			}
		}()
		plistFileUpploadResult, err := s.UploadFileFromFile(plistFile)
		if err != nil {
			return "", err
		}
		pkgInfo = &PKGInfo{
			PackageName: info.Name,
			Version:     info.Version,
			BuildID:     info.Build,
			DisplayName: info.Name,
			Logo:        logoURL,
		}
		meta = map[string]interface{}{"packageName": info.Name, "displayName": info.Name, "version": info.Version,
			"buildID": info.Build, "bundleId": info.BundleId, "installPlist": plistFileUpploadResult.DownloadURL,
			"log": logoURL, "build": info.Build, "appStoreURL": getAppStoreURL(info.BundleId)}
	}

	meta["byteSize"] = mobileFileUploadResult.ByteSize
	meta["fileId"] = mobileFileUploadResult.ID
	releaseResources := []apistructs.ReleaseResource{{
		Type: resourceType,
		Name: mobileFileUploadResult.DisplayName,
		URL:  mobileFileUploadResult.DownloadURL,
		Meta: meta,
	}}
	resourceBytes, err := json.Marshal(releaseResources)
	if err != nil {
		return "", err
	}
	resource := string(resourceBytes)

	versionInfo := &pb.VersionInfo{
		PackageName: pkgInfo.PackageName,
		Version:     pkgInfo.Version,
		BuildID:     pkgInfo.BuildID,
	}

	// 离线包没有项目和应用名，展示需要，拿这个顶一下
	itemVersionMeta := map[string]string{"appName": pkgInfo.PackageName, "projectName": "OFFLINE"}
	itemVersionMetaBytes, err := json.Marshal(itemVersionMeta)
	if err != nil {
		return "", err
	}
	itemVersionMetaStr := string(itemVersionMetaBytes)

	version, err := s.db.GetPublishItemVersionByName(param.OrgID, param.PublishItemID, resourceType, versionInfo)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return "", err
	}
	// 相同版本不会直接覆盖
	if version != nil {
		return "", fmt.Errorf("The version already existed, please check the version, versionCode or BundleVersion: %s-%s",
			version.Version, version.BuildID)
	}

	itemVersion := db.PublishItemVersion{
		Version:       pkgInfo.Version,
		BuildID:       pkgInfo.BuildID,
		PackageName:   pkgInfo.PackageName,
		Public:        false,
		OrgID:         param.OrgID,
		Desc:          param.Desc,
		Logo:          pkgInfo.Logo,
		PublishItemID: param.PublishItemID,
		MobileType:    string(resourceType),
		IsDefault:     false,
		Resources:     resource,
		Creator:       param.IdentityInfo.UserID,
		Meta:          itemVersionMetaStr,
		Readme:        "",
		Spec:          "",
		Swagger:       "",
	}
	if err = s.db.CreatePublishItemVersion(&itemVersion); err != nil {
		return "", err
	}

	return string(resourceType), nil
}

// checkPublishItemExsit check artifact is exist
func (s *PublishItemService) checkPublishItemExsit(publishItemID int64) (*db.PublishItem, error) {
	publishItem, err := s.db.GetPublishItem(publishItemID)
	if err != nil {
		return nil, err
	}
	if publishItem == nil {
		return nil, fmt.Errorf("non-existing publishItem information，publishItem id: %d", publishItemID)
	}
	return publishItem, nil
}

func (s *PublishItemService) updatePublishItemImpl(req *pb.UpdatePublishItemRequest) error {
	item, err := s.db.GetPublishItem(req.ID)
	if err != nil {
		return err
	}
	item.Desc = req.Desc
	item.Public = req.Public
	item.Logo = req.Logo
	item.DisplayName = req.DisplayName
	item.GeofenceLat = req.GeofenceLat
	item.GeofenceLon = req.GeofenceLon
	item.GeofenceRadius = req.GeofenceRadius
	item.NoJailbreak = req.NoJailbreak
	item.GrayLevelPercent = int(req.GrayLevelPercent)
	item.PreviewImages = strings.Join(req.PreviewImages, ",")
	item.BackgroundImage = req.BackgroundImage
	return s.db.UpdatePublishItem(item)
}

func (s *PublishItemService) getOrgIDFromContext(ctx context.Context) (int64, error) {
	orgID := apis.GetOrgID(ctx)
	if orgID == "" {
		return 0, nil
	}
	return strconv.ParseInt(orgID, 10, 64)
}

func getPublishItemId(vars map[string]string) (int64, error) {
	itemIDStr := vars["publishItemId"]
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		return 0, errors.New("publishItem id parse failed")
	}
	return itemID, nil
}

func getPermissionHeader(r *http.Request) (int64, error) {
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return 0, nil
	}
	return strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
}
