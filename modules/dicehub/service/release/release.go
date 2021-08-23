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
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dicehub/conf"
	"github.com/erda-project/erda/modules/dicehub/dbclient"
	"github.com/erda-project/erda/modules/dicehub/event"
	imagedb "github.com/erda-project/erda/modules/dicehub/image/db"
	"github.com/erda-project/erda/modules/dicehub/registry"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/template"
)

const (
	// AliYunRegistry 阿里云registry前缀
	AliYunRegistry = "registry.cn-hangzhou.aliyuncs.com"
)

// Release Release操作封装
type Release struct {
	db      *dbclient.DBClient
	bdl     *bundle.Bundle
	imageDB *imagedb.ImageConfigDB
}

// Option 定义 Release 对象的配置选项
type Option func(*Release)

// New 新建 Release 实例，操作 Release 资源
func New(options ...Option) *Release {
	app := &Release{}
	for _, op := range options {
		op(app)
	}
	return app
}

// WithDBClient 配置 db client
func WithDBClient(db *dbclient.DBClient) Option {
	return func(a *Release) {
		a.db = db
	}
}

// WithDBClient 配置 db client
func WithImageDBClient(db *imagedb.ImageConfigDB) Option {
	return func(a *Release) {
		a.imageDB = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(a *Release) {
		a.bdl = bdl
	}
}

// Create 创建 Release
func (r *Release) Create(req *apistructs.ReleaseCreateRequest) (string, error) {
	// 确保Version在应用层面唯一，若存在，则更新
	if req.Version != "" && req.ApplicationID > 0 {
		releases, err := r.db.GetReleasesByAppAndVersion(req.OrgID, req.ProjectID, req.ApplicationID, req.Version)
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
			if err := r.db.UpdateRelease(&releases[0]); err != nil {
				return "", err
			}
			return releases[0].ReleaseID, nil
		}
	}

	// 创建Release
	release, err := r.Convert(req)
	if err != nil {
		return "", err
	}
	err = r.db.CreateRelease(release)
	if err != nil {
		return "", err
	}

	// 创建Image
	images := r.GetImages(req)
	for _, v := range images {
		v.ReleaseID = release.ReleaseID
		if err := r.imageDB.CreateImage(v); err != nil {
			return "", err
		}
	}

	// Send release create event to eventbox
	event.SendReleaseEvent(event.ReleaseEventCreate, release)

	return release.ReleaseID, nil
}

// Update 更新 Release
func (r *Release) Update(orgID int64, releaseID string, req *apistructs.ReleaseUpdateRequestData) error {
	release, err := r.db.GetRelease(releaseID)
	if err != nil {
		return err
	}
	if orgID != 0 && release.OrgID != orgID {
		return errors.Errorf("release not found")
	}

	// 若version不为空时，确保Version在应用层面唯一
	if req.Version != "" && req.Version != release.Version {
		if req.ApplicationID > 0 {
			releases, err := r.db.GetReleasesByAppAndVersion(req.OrgID, req.ProjectID, req.ApplicationID, req.Version)
			if err != nil {
				return err
			}
			if len(releases) > 0 {
				return errors.Errorf("release version: %s already exist", req.Version)
			}
		}
	}

	// 更新描述
	if req.Desc != "" {
		release.Desc = req.Desc
	}
	// 更新Version
	if req.Version != "" {
		release.Version = req.Version
	}

	if err := r.db.UpdateRelease(release); err != nil {
		return err
	}

	// Send release update event to eventbox
	event.SendReleaseEvent(event.ReleaseEventUpdate, release)

	return nil
}

// UpdateReference 更新 Release 引用
func (r *Release) UpdateReference(orgID int64, releaseID string, req *apistructs.ReleaseReferenceUpdateRequest) error {
	release, err := r.db.GetRelease(releaseID)
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
	if err := r.db.UpdateRelease(release); err != nil {
		return err
	}

	return nil
}

// Delete 删除 Release
func (r *Release) Delete(orgID int64, releaseID string) error {
	release, err := r.db.GetRelease(releaseID)
	if err != nil {
		return err
	}
	if orgID != 0 && release.OrgID != orgID {
		return errors.Errorf("release not found")
	}

	// Release被使用时，不可删除
	if release.Reference > 0 {
		return errors.Errorf("reference > 0")
	}

	images, err := r.imageDB.GetImagesByRelease(releaseID)
	if err != nil {
		return err
	}

	// delete manifests
	if release.ClusterName != "" {
		var imgs []string
		for _, v := range images {
			imgs = append(imgs, v.Image)
		}
		if err := registry.DeleteManifests(r.bdl, release.ClusterName, imgs); err != nil {
			logrus.Errorf(err.Error())
		}
	}

	// delete images from db
	for _, v := range images {
		if err := r.imageDB.DeleteImage(int64(v.ID)); err != nil {
			logrus.Errorf("[alert] delete image: %s fail, err: %v", v.Image, err)
		}
		logrus.Infof("deleted image: %s", v.Image)
	}

	// delete release info
	if err := r.db.DeleteRelease(releaseID); err != nil {
		return err
	}

	// send release delete event to eventbox
	event.SendReleaseEvent(event.ReleaseEventDelete, release)

	return nil
}

// Get 获取 Release 详情
func (r *Release) Get(orgID int64, releaseID string) (*apistructs.ReleaseGetResponseData, error) {
	release, err := r.db.GetRelease(releaseID)
	if err != nil {
		return nil, err
	}
	if orgID != 0 && release.OrgID != orgID {
		return nil, errors.Errorf("release not found")
	}

	images, err := r.imageDB.GetImagesByRelease(releaseID)
	if err != nil {
		return nil, err
	}
	releaseInfoResponse := r.convertToReleaseResponse(release)
	for _, v := range images {
		releaseInfoResponse.Images = append(releaseInfoResponse.Images, v.Image)
	}

	return releaseInfoResponse, nil
}

// List 根据搜索条件进行搜索
func (r *Release) List(orgID int64, req *apistructs.ReleaseListRequest) (*apistructs.ReleaseListResponseData, error) {
	startTime := time.Unix(req.StartTime/1000, 0)
	endTime := time.Unix(req.EndTime/1000, 0)
	total, releases, err := r.db.GetReleasesByParams(
		orgID, req.ProjectID, req.ApplicationID,
		req.Query, req.ReleaseName, req.Branch,
		req.Cluster, req.CrossCluster, req.IsVersion,
		req.CrossClusterOrSpecifyCluster,
		startTime, endTime, req.PageNum, req.PageSize)
	if err != nil {
		return nil, err
	}

	releaseList := make([]apistructs.ReleaseGetResponseData, 0, len(releases))
	for _, v := range releases {
		releaseList = append(releaseList, *r.convertToReleaseResponse(&v))
	}

	return &apistructs.ReleaseListResponseData{
		Total:    total,
		Releases: releaseList,
	}, nil
}

// GetDiceYAML 获取dice.yml内容
func (r *Release) GetDiceYAML(orgID int64, releaseID string) (string, error) {
	release, err := r.db.GetRelease(releaseID)
	if err != nil {
		return "", err
	}
	if orgID != 0 && release.OrgID != orgID { // when calling internally，orgID is 0
		return "", errors.Errorf("release not found")
	}

	return release.Dice, nil
}

// GetIosPlist 读取ios类型release中下载地址plist
func (r *Release) GetIosPlist(orgID int64, releaseID string) (string, error) {
	release, err := r.db.GetRelease(releaseID)
	if err != nil {
		return "", err
	}
	if orgID != 0 && release.OrgID != orgID { // when calling internally，orgID is 0
		return "", errors.Errorf("release not found")
	}

	releaseData := r.convertToReleaseResponse(release)
	for _, resource := range releaseData.Resources {
		if resource.Type == apistructs.ResourceTypeIOS {
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
			bundleId := resource.Meta["bundleId"].(string)
			version := resource.Meta["version"].(string)
			displayName := resource.Meta["displayName"].(string)
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

// GetReleaseNamesByApp 根据 appID 获取 releaseName 列表
func (r *Release) GetReleaseNamesByApp(orgID, appID int64) ([]string, error) {
	// releaseNames := make([]string, 0)
	// for _, item := range releases {
	// 	releaseNames = append(releaseNames, item.ReleaseName)
	// }
	return r.db.GetReleaseNamesByApp(orgID, appID)
}

// GetLatestReleasesByProjectAndVersion 获取给定项目 & version情况下各应用最新 Release
func (r *Release) GetLatestReleasesByProjectAndVersion(projectID int64, version string) (*[]dbclient.Release, error) {
	appIDs, err := r.db.GetAppIDsByProjectAndVersion(projectID, version)
	if err != nil {
		return nil, err
	}
	latests := make([]dbclient.Release, 0, len(appIDs))
	for _, v := range appIDs {
		release, err := r.db.GetLatestReleaseByAppAndVersion(v, version)
		if err != nil {
			logrus.Warnf("failed to get latest release, (%v)", err)
			continue
		}
		latests = append(latests, *release)
	}

	return &latests, nil
}

// RemoveDeprecatedsReleases 回收过期release具体逻辑
func (r *Release) RemoveDeprecatedsReleases(now time.Time) error {
	d, err := time.ParseDuration(strutil.Concat("-", conf.MaxTimeReserved(), "h")) // one month before, eg: -720h
	if err != nil {
		return err
	}
	before := now.Add(d)

	releases, err := r.db.GetUnReferedReleasesBefore(before)
	if err != nil {
		return err
	}
	for i := range releases {
		release := releases[i]
		if release.Version != "" {
			logrus.Debugf("release %s have been tagged, can't be recycled", release.ReleaseID)
			continue
		}

		images, err := r.imageDB.GetImagesByRelease(release.ReleaseID)
		if err != nil {
			logrus.Warnf(err.Error())
			continue
		}

		deletable := true // 若release下的image manifest删除失败，release不可删除
		for _, image := range images {
			// 若有其他release引用此镜像，镜像manifest不可删，只删除DB元信息(多次构建，存在镜像相同的情况)
			count, err := r.imageDB.GetImageCount(release.ReleaseID, image.Image)
			if err != nil {
				logrus.Errorf(err.Error())
				continue
			}
			if count == 0 && release.ClusterName != "" && !strings.HasPrefix(image.Image, AliYunRegistry) {
				if err := registry.DeleteManifests(r.bdl, release.ClusterName, []string{image.Image}); err != nil {
					deletable = false
					logrus.Errorf(err.Error())
					continue
				}
			}

			// Delete image info
			if err := r.imageDB.DeleteImage(int64(image.ID)); err != nil {
				logrus.Errorf("[alert] delete image: %s fail, err: %v", image.Image, err)
			}
			logrus.Infof("deleted image: %s", image.Image)
		}

		if deletable {
			// Delete release info
			if err := r.db.DeleteRelease(release.ReleaseID); err != nil {
				logrus.Errorf("[alert] delete release: %s fail, err: %v", release.ReleaseID, err)
			}
			logrus.Infof("deleted release: %s", release.ReleaseID)

			// Send release delete event to eventbox
			event.SendReleaseEvent(event.ReleaseEventDelete, &release)
		}
	}
	return nil
}

// Convert 从ReleaseRequest中提取Release元信息
func (r *Release) Convert(releaseRequest *apistructs.ReleaseCreateRequest) (*dbclient.Release, error) {
	release := dbclient.Release{
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

// release数据库结构转换为API返回所需结构
func (r *Release) convertToReleaseResponse(release *dbclient.Release) *apistructs.ReleaseGetResponseData {
	var labels map[string]string
	err := json.Unmarshal([]byte(release.Labels), &labels)
	if err != nil {
		labels = make(map[string]string)
	}

	var resources []apistructs.ReleaseResource
	err = json.Unmarshal([]byte(release.Resources), &resources)
	if err != nil {
		resources = make([]apistructs.ReleaseResource, 0)
	}

	respData := &apistructs.ReleaseGetResponseData{
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
		CreatedAt:       release.CreatedAt,
		UpdatedAt:       release.UpdatedAt,
		ProjectName:     release.ProjectName,
		ApplicationName: release.ApplicationName,
		UserID:          release.UserID,
		CrossCluster:    release.CrossCluster,
	}
	return respData
}

// GetImages 从ReleaseRequest中提取Image信息
func (r *Release) GetImages(req *apistructs.ReleaseCreateRequest) []*imagedb.Image {
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
