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

package release

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v2"

	pb "github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	imagedb "github.com/erda-project/erda/modules/dicehub/image/db"
	"github.com/erda-project/erda/modules/dicehub/registry"
	"github.com/erda-project/erda/modules/dicehub/release/db"
	"github.com/erda-project/erda/modules/dicehub/release/event"
	"github.com/erda-project/erda/modules/dicehub/service/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/template"
)

type releaseService struct {
	p       *provider
	db      *db.ReleaseConfigDB
	imageDB *imagedb.ImageConfigDB
	bdl     *bundle.Bundle
	Etcd    *clientv3.Client
	Config  *releaseConfig
}

// CreateRelease POST /api/releases release create release
func (s *releaseService) CreateRelease(ctx context.Context, req *pb.ReleaseCreateRequest) (*pb.ReleaseCreateResponseData, error) {
	_, err := getPermissionHeader(ctx)
	if err != nil {
		return nil, apierrors.ErrCreateRelease.NotLogin()
	}

	if req == nil {
		return nil, apierrors.ErrCreateRelease.MissingParameter("body")
	}

	logrus.Infof("creating release...request body: %v\n", req)

	if req.ReleaseName == "" {
		return nil, apierrors.ErrCreateRelease.MissingParameter("releaseName")
	}

	// create Release
	releaseID, err := s.Create(req)
	if err != nil {
		return nil, apierrors.ErrCreateRelease.InternalError(err)
	}

	respBody := &pb.ReleaseCreateResponseData{
		ReleaseID: releaseID,
	}

	return respBody, nil
}

func (s *releaseService) UpdateRelease(ctx context.Context, req *pb.ReleaseUpdateRequest) (*pb.ReleaseDataResponse, error) {
	orgID, err := getPermissionHeader(ctx)
	if err != nil {
		return nil, apierrors.ErrUpdateRelease.NotLogin()
	}

	// Check releaseId if exist in path or not
	releaseID := req.ReleaseID
	if releaseID == "" {
		return nil, apierrors.ErrUpdateRelease.MissingParameter("releaseId")
	}

	logrus.Infof("update release info: %+v", req)

	if err := s.Update(orgID, releaseID, req); err != nil {
		return nil, apierrors.ErrUpdateRelease.InternalError(err)
	}

	return &pb.ReleaseDataResponse{Data: "Update succ"}, nil
}

// UpdateRelease PUT /api/releases/<releaseId> update release
func (s *releaseService) UpdateReleaseReference(ctx context.Context, req *pb.ReleaseReferenceUpdateRequest) (*pb.ReleaseDataResponse, error) {
	orgID, err := getPermissionHeader(ctx)
	if err != nil {
		return nil, apierrors.ErrUpdateRelease.NotLogin()
	}

	// Check releaseId if exist in path or not
	releaseID := req.ReleaseID
	if releaseID == "" {
		return nil, apierrors.ErrUpdateRelease.MissingParameter("releaseId")
	}

	// Only update reference
	if req == nil {
		return nil, apierrors.ErrUpdateRelease.MissingParameter("body")
	}

	updateReferenceRequest := *req

	if err := s.UpdateReference(orgID, releaseID, &updateReferenceRequest); err != nil {
		return nil, apierrors.ErrUpdateRelease.InternalError(err)
	}

	return &pb.ReleaseDataResponse{Data: "Update succ"}, nil
}

// GetPlist GET /api/releases/<releaseId>/actions/get-plist Get the download plist configuration in the ios release type
func (s *releaseService) GetIosPlist(ctx context.Context, req *pb.GetIosPlistRequest) (*pb.GetIosPlistResponse, error) {
	orgID, err := getPermissionHeader(ctx)
	if err != nil {
		return nil, apierrors.ErrUpdateRelease.NotLogin()
	}
	releaseID := req.ReleaseID
	if releaseID == "" {
		return nil, apierrors.ErrGetIosPlist.MissingParameter("releaseId")
	}

	plist, err := s.GetIosPlistService(orgID, releaseID)
	if err != nil {
		return nil, apierrors.ErrGetIosPlist.NotFound()
	}
	fmt.Println(plist)
	return &pb.GetIosPlistResponse{Data: plist}, nil
}
func (s *releaseService) GetRelease(ctx context.Context, req *pb.GetIosPlistRequest) (*pb.ReleaseGetResponse, error) {
	orgID, err := getPermissionHeader(ctx)
	if err != nil {
		return nil, apierrors.ErrUpdateRelease.NotLogin()
	}

	releaseID := req.ReleaseID
	if releaseID == "" {
		return nil, apierrors.ErrGetRelease.MissingParameter("releaseId")
	}
	logrus.Infof("getting release...releaseId: %s\n", releaseID)

	resp, err := s.Get(orgID, releaseID)
	if err != nil {
		return nil, apierrors.ErrGetRelease.InternalError(err)
	}
	return &pb.ReleaseGetResponse{Data: resp}, nil
}

func (s *releaseService) DeleteRelease(ctx context.Context, req *pb.GetIosPlistRequest) (*pb.ReleaseDataResponse, error) {
	orgID, err := getPermissionHeader(ctx)
	if err != nil {
		return nil, apierrors.ErrUpdateRelease.NotLogin()
	}
	// Get releaseId
	releaseID := req.ReleaseID
	if releaseID == "" {
		return nil, apierrors.ErrDeleteRelease.MissingParameter("releaseId")
	}
	logrus.Infof("deleting release...releaseId: %s\n", releaseID)

	if err := s.Delete(orgID, releaseID); err != nil {
		return nil, apierrors.ErrDeleteRelease.InternalError(err)
	}

	return &pb.ReleaseDataResponse{Data: "Delete succ"}, nil
}

func (s *releaseService) ListRelease(ctx context.Context, req *pb.ReleaseListRequest) (*pb.ReleaseListResponse, error) {
	identityInfo, err := getIdentityInfo(ctx)
	if err != nil {
		return nil, apierrors.ErrListRelease.NotLogin()
	}

	if req.EndTime == 0 {
		req.EndTime = time.Now().UnixNano() / 1000 / 1000 // milliseconds
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	params := req

	var orgID int64

	if !identityInfo.IsInternalClient() {
		orgID, err = getPermissionHeader(ctx)
		if err != nil {
			return nil, apierrors.ErrListRelease.NotLogin()
		}

		// Get User
		userID := apis.GetUserID(ctx)
		if userID == "" {
			return nil, errors.New("invalid user id")
		}

		// TODO：If there is no list release permission of the application, then judge whether there is enterprise permission, and after adding scope to the permission list, modify the authentication method
		var (
			req      apistructs.PermissionCheckRequest
			permResp *apistructs.PermissionCheckResponseData
			access   bool
		)

		if params.ApplicationID > 0 {
			req = apistructs.PermissionCheckRequest{
				UserID:   userID,
				Scope:    apistructs.AppScope,
				ScopeID:  uint64(params.ApplicationID),
				Resource: "release",
				Action:   apistructs.ListAction,
			}

			if permResp, err = s.bdl.CheckPermission(&req); err != nil {
				return nil, apierrors.ErrListRelease.AccessDenied()
			}

			access = permResp.Access
		}

		if !access {
			req = apistructs.PermissionCheckRequest{
				UserID:   userID,
				Scope:    apistructs.OrgScope,
				ScopeID:  uint64(orgID),
				Resource: "release",
				Action:   apistructs.ListAction,
			}

			if permResp, err = s.bdl.CheckPermission(&req); err != nil || !permResp.Access {
				return nil, apierrors.ErrListRelease.AccessDenied()
			}
		}
	}

	resp, err := s.List(orgID, params)
	if err != nil {
		return nil, apierrors.ErrListRelease.InternalError(err)
	}
	userIDs := make([]string, 0, len(resp.Releases))
	for _, v := range resp.Releases {
		userIDs = append(userIDs, v.UserID)
	}

	return &pb.ReleaseListResponse{
		Data:    resp,
		UserIDs: userIDs,
	}, nil
}
func (s *releaseService) ListReleaseName(ctx context.Context, req *pb.ListReleaseNameRequest) (*pb.ListReleaseNameResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, apierrors.ErrListRelease.NotLogin()
	}

	// get applicationID
	applicationIDStr := GetApplicationId(ctx)
	if applicationIDStr == "" {
		return nil, apierrors.ErrListRelease.MissingParameter("applicationId")
	}
	applicationID, err := strconv.ParseInt(applicationIDStr, 10, 64)
	if err != nil { // Prevent SQL injection
		return nil, apierrors.ErrListRelease.InvalidParameter("applicationId")
	}

	releaseNames, err := s.GetReleaseNamesByApp(orgID, applicationID)
	if err != nil {
		return nil, apierrors.ErrListRelease.InternalError(err)
	}

	return &pb.ListReleaseNameResponse{Data: releaseNames}, nil
}
func (s *releaseService) GetLatestReleases(ctx context.Context, req *pb.GetLatestReleasesRequest) (*pb.GetLatestReleasesResponse, error) {
	// Check the legitimacy of the projectID
	projectIDStr := GetProjectID(ctx)
	if projectIDStr == "" {
		return nil, apierrors.ErrListRelease.MissingParameter("projectId")
	}
	projectID, err := strutil.Atoi64(projectIDStr)
	if err != nil {
		return nil, apierrors.ErrListRelease.InvalidParameter(err)
	}

	// Check the legitimacy of the version
	version := GetVersion(ctx)
	if version == "" {
		return nil, apierrors.ErrListRelease.MissingParameter("version")
	}

	latests, err := s.GetLatestReleasesByProjectAndVersion(projectID, version)
	if err != nil {
		return nil, apierrors.ErrListRelease.InternalError(err)
	}
	resp, err := batchConvert(latests)
	if err != nil {
		return nil, apierrors.ErrListRelease.InternalError(err)
	}
	return &pb.GetLatestReleasesResponse{Data: resp}, nil
}
func (s *releaseService) ReleaseGC(ctx context.Context, req *pb.ReleaseGCRequest) (*pb.ReleaseDataResponse, error) {
	logrus.Infof("trigger release gc by api[ POST /gc ]!")
	go func() {
		if err := s.RemoveDeprecatedsReleases(time.Now()); err != nil {
			logrus.Warnf("remove deprecated release error: %v", err)
		}
	}()

	return &pb.ReleaseDataResponse{Data: "trigger release gc success"}, nil
}

// Create create Release
func (s *releaseService) Create(req *pb.ReleaseCreateRequest) (string, error) {
	// Ensure that the Version is unique at the application level, if it exists, update it
	if req.Version != "" && req.ApplicationID > 0 {
		releases, err := s.db.GetReleasesByAppAndVersion(req.OrgID, req.ProjectID, req.ApplicationID, req.Version)
		if err != nil {
			return "", err
		}
		if len(releases) > 0 {
			releases[0].Dice = req.Dice
			if len(req.Labels) > 0 {
				labelBytes, err := json.Marshal(req.Labels)
				if err != nil {
					return "", err
				}
				releases[0].Labels = string(labelBytes)
			}
			resourceBytes, err := json.Marshal(req.Resources)
			if err == nil {
				releases[0].Resources = string(resourceBytes)
			}
			if err := s.db.UpdateRelease(&releases[0]); err != nil {
				return "", err
			}
			return releases[0].ReleaseID, nil
		}
	}

	// create Release
	release, err := s.Convert(req)
	if err != nil {
		return "", err
	}
	err = s.db.CreateRelease(release)
	if err != nil {
		return "", err
	}

	// create Image
	images := s.GetImages(req)
	for _, v := range images {
		v.ReleaseID = release.ReleaseID
		if err := s.imageDB.CreateImage(v); err != nil {
			return "", err
		}
	}

	// Send release create event to eventbox
	event.SendReleaseEvent(event.ReleaseEventCreate, release)

	return release.ReleaseID, nil
}

// Update update Release
func (s *releaseService) Update(orgID int64, releaseID string, req *pb.ReleaseUpdateRequest) error {
	release, err := s.db.GetRelease(releaseID)
	if err != nil {
		return err
	}
	if orgID != 0 && release.OrgID != orgID {
		return errors.Errorf("release not found")
	}

	// If version is not empty, ensure that Version is unique at the application level
	if req.Version != "" && req.Version != release.Version {
		if req.ApplicationID > 0 {
			releases, err := s.db.GetReleasesByAppAndVersion(req.OrgID, req.ProjectID, req.ApplicationID, req.Version)
			if err != nil {
				return err
			}
			if len(releases) > 0 {
				return errors.Errorf("release version: %s already exist", req.Version)
			}
		}
	}

	if req.Desc != "" {
		release.Desc = req.Desc
	}

	if req.Version != "" {
		release.Version = req.Version
	}

	if err := s.db.UpdateRelease(release); err != nil {
		return err
	}

	// Send release update event to eventbox
	event.SendReleaseEvent(event.ReleaseEventUpdate, release)

	return nil
}

// UpdateReference update Release reference
func (s *releaseService) UpdateReference(orgID int64, releaseID string, req *pb.ReleaseReferenceUpdateRequest) error {
	release, err := s.db.GetRelease(releaseID)
	if err != nil {
		return err
	}
	if orgID != 0 && release.OrgID != orgID {
		return errors.Errorf("release not found")
	}

	if req.Increase {
		release.Reference++
	} else {
		release.Reference--
	}
	if err := s.db.UpdateRelease(release); err != nil {
		return err
	}

	return nil
}

// Delete delete Release
func (s *releaseService) Delete(orgID int64, releaseID string) error {
	release, err := s.db.GetRelease(releaseID)
	if err != nil {
		return err
	}
	if orgID != 0 && release.OrgID != orgID {
		return errors.Errorf("release not found")
	}

	// If release has been used，can not delete it
	if release.Reference > 0 {
		return errors.Errorf("reference > 0")
	}

	images, err := s.imageDB.GetImagesByRelease(releaseID)
	if err != nil {
		return err
	}

	// delete manifests
	if release.ClusterName != "" {
		var imgs []string
		for _, v := range images {
			imgs = append(imgs, v.Image)
		}
		if err := registry.DeleteManifests(s.bdl, release.ClusterName, imgs); err != nil {
			logrus.Errorf(err.Error())
		}
	}

	// delete images from db
	for _, v := range images {
		if err := s.imageDB.DeleteImage(int64(v.ID)); err != nil {
			logrus.Errorf("[alert] delete image: %s fail, err: %v", v.Image, err)
		}
		logrus.Infof("deleted image: %s", v.Image)
	}

	// delete release info
	if err := s.db.DeleteRelease(releaseID); err != nil {
		return err
	}

	// send release delete event to eventbox
	event.SendReleaseEvent(event.ReleaseEventDelete, release)

	return nil
}

// Get get Release
func (s *releaseService) Get(orgID int64, releaseID string) (*pb.ReleaseGetResponseData, error) {
	release, err := s.db.GetRelease(releaseID)
	if err != nil {
		return nil, err
	}
	if orgID != 0 && release.OrgID != orgID {
		return nil, errors.Errorf("release not found")
	}

	images, err := s.imageDB.GetImagesByRelease(releaseID)
	if err != nil {
		return nil, err
	}
	releaseInfoResponse := s.convertToReleaseResponse(release)
	for _, v := range images {
		releaseInfoResponse.Images = append(releaseInfoResponse.Images, v.Image)
	}

	return releaseInfoResponse, nil
}

// List Search based on search parameters
func (s *releaseService) List(orgID int64, req *pb.ReleaseListRequest) (*pb.ReleaseListResponseData, error) {
	startTime := time.Unix(req.StartTime/1000, 0)
	endTime := time.Unix(req.EndTime/1000, 0)
	total, releases, err := s.db.GetReleasesByParams(
		orgID, req.ProjectID, req.ApplicationID,
		req.Query, req.ReleaseName, req.Branch,
		req.Cluster, req.CrossCluster, req.IsVersion,
		req.CrossClusterOrSpecifyCluster,
		startTime, endTime, req.PageNum, req.PageSize)
	if err != nil {
		return nil, err
	}

	releaseList := make([]*pb.ReleaseGetResponseData, 0, len(releases))
	for _, v := range releases {
		releaseList = append(releaseList, s.convertToReleaseResponse(&v))
	}

	return &pb.ReleaseListResponseData{
		Total:    total,
		Releases: releaseList,
	}, nil
}

// GetDiceYAML get dice.yml context
func (s *releaseService) GetDiceYAML(orgID int64, releaseID string) (string, error) {
	release, err := s.db.GetRelease(releaseID)
	if err != nil {
		return "", err
	}
	if orgID != 0 && release.OrgID != orgID { // when calling internal，orgID is 0
		return "", errors.Errorf("release not found")
	}

	return release.Dice, nil
}

// GetIosPlist Read the download address plist in the ios type release
func (s *releaseService) GetIosPlistService(orgID int64, releaseID string) (string, error) {
	release, err := s.db.GetRelease(releaseID)
	if err != nil {
		return "", err
	}
	if orgID != 0 && release.OrgID != orgID { // 内部调用时，orgID为0
		return "", errors.Errorf("release not found")
	}

	releaseData := s.convertToReleaseResponse(release)
	for _, resource := range releaseData.Resources {
		if ResourceType(resource.Type) == ResourceTypeIOS {
			plistTemplate := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
   <key>items</key>
   <array>
       <dict>
           <key>assets</key>
           <array>
               <dict>
                   <key>kind</key>
                   <string>software-package</string>
                   <key>url</key>
                   <string>{{appUrl}}</string>
               </dict>
           </array>
           <key>metadata</key>
           <dict>
               <key>bundle-identifier</key>
               <string>{{bundleId}}</string>
               <key>bundle-version</key>
               <string>{{version}}</string>
               <key>kind</key>
               <string>software</string>
               <key>subtitle</key>
               <string>{{displayName}}</string>
               <key>title</key>
               <string>{{displayName}}</string>
           </dict>
       </dict>
   </array>
</dict>
</plist>`
			bundleId := resource.Meta["bundleId"].GetStringValue()
			version := resource.Meta["version"].GetStringValue()
			displayName := resource.Meta["displayName"].GetStringValue()
			appUrl := resource.URL
			plistContent := template.Render(plistTemplate, map[string]string{
				"bundleId":    bundleId,
				"version":     version,
				"displayName": displayName,
				"appUrl":      appUrl,
			})
			return plistContent, nil
		}
	}
	return "", errors.New("not ios release")
}

// GetReleaseNamesByApp get releaseName list by appID
func (s *releaseService) GetReleaseNamesByApp(orgID, appID int64) ([]string, error) {
	// releaseNames := make([]string, 0)
	// for _, item := range releases {
	// 	releaseNames = append(releaseNames, item.ReleaseName)
	// }
	return s.db.GetReleaseNamesByApp(orgID, appID)
}

// GetLatestReleasesByProjectAndVersion get latelest Release by projectID & version
func (s *releaseService) GetLatestReleasesByProjectAndVersion(projectID int64, version string) (*[]db.Release, error) {
	appIDs, err := s.db.GetAppIDsByProjectAndVersion(projectID, version)
	if err != nil {
		return nil, err
	}
	latests := make([]db.Release, 0, len(appIDs))
	for _, v := range appIDs {
		release, err := s.db.GetLatestReleaseByAppAndVersion(v, version)
		if err != nil {
			logrus.Warnf("failed to get latest release, (%v)", err)
			continue
		}
		latests = append(latests, *release)
	}

	return &latests, nil
}

// RemoveDeprecatedsReleases Recycling expired release
func (s *releaseService) RemoveDeprecatedsReleases(now time.Time) error {
	d, err := time.ParseDuration(strutil.Concat("-", s.Config.MaxTimeReserved, "h")) // one month before, eg: -720h
	if err != nil {
		return err
	}
	before := now.Add(d)

	releases, err := s.db.GetUnReferedReleasesBefore(before)
	if err != nil {
		return err
	}
	for i := range releases {
		release := releases[i]
		if release.Version != "" {
			logrus.Debugf("release %s have been tagged, can't be recycled", release.ReleaseID)
			continue
		}

		images, err := s.imageDB.GetImagesByRelease(release.ReleaseID)
		if err != nil {
			logrus.Warnf(err.Error())
			continue
		}

		deletable := true // if image manifest delete false，release can not be deleted
		for _, image := range images {
			// If there are other releases that refer to this mirror, the mirror manifest cannot be deleted, only the DB meta information is deleted (multiple builds, the same mirror exists)
			count, err := s.imageDB.GetImageCount(release.ReleaseID, image.Image)
			if err != nil {
				logrus.Errorf(err.Error())
				continue
			}
			if count == 0 && release.ClusterName != "" && !strings.HasPrefix(image.Image, AliYunRegistry) {
				if err := registry.DeleteManifests(s.bdl, release.ClusterName, []string{image.Image}); err != nil {
					deletable = false
					logrus.Errorf(err.Error())
					continue
				}
			}

			// Delete image info
			if err := s.imageDB.DeleteImage(int64(image.ID)); err != nil {
				logrus.Errorf("[alert] delete image: %s fail, err: %v", image.Image, err)
			}
			logrus.Infof("deleted image: %s", image.Image)
		}

		if deletable {
			// Delete release info
			if err := s.db.DeleteRelease(release.ReleaseID); err != nil {
				logrus.Errorf("[alert] delete release: %s fail, err: %v", release.ReleaseID, err)
			}
			logrus.Infof("deleted release: %s", release.ReleaseID)

			// Send release delete event to eventbox
			event.SendReleaseEvent(event.ReleaseEventDelete, &release)
		}
	}
	return nil
}

func (s *releaseService) Convert(releaseRequest *pb.ReleaseCreateRequest) (*db.Release, error) {
	release := db.Release{
		ReleaseID:       uuid.UUID(),
		ReleaseName:     releaseRequest.ReleaseName,
		Desc:            releaseRequest.Desc,
		Dice:            releaseRequest.Dice,
		Addon:           releaseRequest.Addon,
		Version:         releaseRequest.Version,
		OrgID:           releaseRequest.OrgID,
		ProjectID:       releaseRequest.ProjectID,
		ApplicationID:   releaseRequest.ApplicationID,
		UserID:          releaseRequest.UserID,
		ClusterName:     releaseRequest.ClusterName,
		ProjectName:     releaseRequest.ProjectName,
		ApplicationName: releaseRequest.ApplicationName,
		CrossCluster:    releaseRequest.CrossCluster,
	}

	if len(releaseRequest.Labels) > 0 {
		labelBytes, err := json.Marshal(releaseRequest.Labels)
		if err != nil {
			return nil, err
		}
		release.Labels = string(labelBytes)
	}

	if len(releaseRequest.Resources) > 0 {
		resourceBytes, err := json.Marshal(releaseRequest.Resources)
		if err != nil {
			return nil, err
		}
		release.Resources = string(resourceBytes)
	}

	return &release, nil
}

func (s *releaseService) convertToReleaseResponse(release *db.Release) *pb.ReleaseGetResponseData {
	var labels map[string]string
	err := json.Unmarshal([]byte(release.Labels), &labels)
	if err != nil {
		labels = make(map[string]string)
	}

	var resources []*pb.ReleaseResource
	err = json.Unmarshal([]byte(release.Resources), &resources)
	if err != nil {
		resources = make([]*pb.ReleaseResource, 0)
	}

	respData := &pb.ReleaseGetResponseData{
		ReleaseID:       release.ReleaseID,
		ReleaseName:     release.ReleaseName,
		Addon:           release.Addon,
		Diceyml:         release.Dice,
		Resources:       resources,
		Desc:            release.Desc,
		Labels:          labels,
		Version:         release.Version,
		Reference:       release.Reference,
		OrgID:           release.OrgID,
		ProjectID:       release.ProjectID,
		ApplicationID:   release.ApplicationID,
		ClusterName:     release.ClusterName,
		CreatedAt:       timestamppb.New(release.CreatedAt),
		UpdatedAt:       timestamppb.New(release.UpdatedAt),
		ProjectName:     release.ProjectName,
		ApplicationName: release.ApplicationName,
		UserID:          release.UserID,
		CrossCluster:    release.CrossCluster,
	}
	return respData
}

// GetImages get image by ReleaseRequest
func (s *releaseService) GetImages(req *pb.ReleaseCreateRequest) []*imagedb.Image {
	var dice diceyml.Object
	err := yaml.Unmarshal([]byte(req.Dice), &dice)
	if err != nil {
		return make([]*imagedb.Image, 0)
	}

	// Get images from dice.yml
	images := make([]*imagedb.Image, 0)
	for key, service := range dice.Services {
		// Check service if contain any image
		if service.Image == "" {
			logrus.Errorf("service %s doesn't contain any image", key)
			continue
		}
		repoName, tag := parseImage(service.Image)
		image := &imagedb.Image{
			Image:     service.Image,
			ImageName: repoName,
			ImageTag:  tag,
		}
		images = append(images, image)
	}
	for key, job := range dice.Jobs {
		// Check service if contain any image
		if job.Image == "" {
			logrus.Errorf("job %s doesn't contain any image", key)
			continue
		}
		repoName, tag := parseImage(job.Image)
		image := &imagedb.Image{
			Image:     job.Image,
			ImageName: repoName,
			ImageTag:  tag,
		}
		images = append(images, image)
	}
	return images
}

// image format: docker-registry.registry.marathon.mesos:5000/pampas-blog/blog-service:v0.2
func parseImage(image string) (repoName, tag string) {
	ss := strings.SplitN(image, "/", 2)
	if len(ss) == 2 {
		repo := strings.Split(ss[1], ":")[0]
		var repoTag string
		if strings.Contains(ss[1], ":") {
			repoTag = strings.Split(ss[1], ":")[1]
		} else {
			repoTag = "latest"
		}
		return repo, repoTag
	}
	return "", ""
}

func batchConvert(releases *[]db.Release) ([]*pb.GetLatestReleasesResponseData, error) {
	var data []*pb.GetLatestReleasesResponseData
	for _, v := range *releases {
		data = append(data, &pb.GetLatestReleasesResponseData{
			ReleaseID:       v.ReleaseID,
			ReleaseName:     v.ReleaseName,
			Dice:            v.Dice,
			Desc:            v.Desc,
			Addon:           v.Addon,
			Resources:       v.Resources,
			Labels:          v.Labels,
			Version:         v.Version,
			CrossCluster:    v.CrossCluster,
			Reference:       v.Reference,
			OrgID:           v.OrgID,
			ProjectID:       v.ProjectID,
			ApplicationID:   v.ApplicationID,
			ProjectName:     v.ProjectName,
			ApplicationName: v.ApplicationName,
			UserID:          v.UserID,
			ClusterName:     v.ClusterName,
			CreatedAt:       timestamppb.New(v.CreatedAt),
			UpdatedAt:       timestamppb.New(v.UpdatedAt),
		})
	}
	return data, nil
}

func getPermissionHeader(ctx context.Context) (int64, error) {
	orgIDStr := apis.GetOrgID(ctx)
	if orgIDStr == "" {
		return 0, nil
	}
	return strconv.ParseInt(orgIDStr, 10, 64)
}

func getIdentityInfo(ctx context.Context) (apistructs.IdentityInfo, error) {
	userID := apis.GetUserID(ctx)
	internalClient := apis.GetHeader(ctx, httputil.InternalHeader)

	// not login
	if userID == "" && internalClient == "" {
		return apistructs.IdentityInfo{}, errors.Errorf("invalid identity info")
	}

	identity := apistructs.IdentityInfo{
		UserID:         userID,
		InternalClient: internalClient,
	}

	return identity, nil
}
