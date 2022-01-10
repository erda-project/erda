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
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
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
	if err := limitLabelsLength(req); err != nil {
		return "", err
	}

	// 确保Version唯一
	if req.IsProjectRelease {
		releases, err := r.db.GetReleasesByProjectAndVersion(req.OrgID, req.ProjectID, req.Version)
		if err != nil {
			return "", err
		}
		if len(releases) > 0 {
			return "", errors.Errorf("release version: %s already exist in target project", req.Version)
		}
	} else if req.Version != "" && req.ApplicationID > 0 {
		releases, err := r.db.GetReleasesByAppAndVersion(req.OrgID, req.ProjectID, req.ApplicationID, req.Version)
		if err != nil {
			return "", err
		}
		if len(releases) > 0 {
			return "", errors.Errorf("release version: %s already exist in target application", req.Version)
		}
	}

	var (
		appReleases []dbclient.Release
		err         error
	)
	if req.IsProjectRelease {
		list := strutil.DedupSlice(req.ApplicationReleaseList)
		sort.Strings(list)
		appReleases, err = r.db.GetReleases(list)
		if err != nil {
			return "", err
		}
	}

	// 创建Release
	release, err := r.Convert(req, appReleases)
	if err != nil {
		return "", err
	}

	var dices []string
	if req.IsProjectRelease {
		for i := range appReleases {
			dices = append(dices, appReleases[i].Dice)
		}
		if err = r.createProjectReleaseAndUpdateReference(release, appReleases); err != nil {
			return "", err
		}
	} else {
		dices = append(dices, req.Dice)
		err = r.db.CreateRelease(release)
		if err != nil {
			return "", err
		}
	}

	// 创建Image
	images := r.GetImages(dices)
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

func (r *Release) createProjectReleaseAndUpdateReference(release *dbclient.Release, appReleases []dbclient.Release) (err error) {
	tx := r.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Create(release).Error; err != nil {
		return
	}

	for i := range appReleases {
		appReleases[i].Reference++
		if err = tx.Save(appReleases[i]).Error; err != nil {
			return
		}
	}
	return tx.Commit().Error
}

func limitLabelsLength(req *apistructs.ReleaseCreateRequest) error {
	if len(req.Labels) == 0 {
		return nil
	}
	labelBytes, err := json.Marshal(req.Labels)
	if err != nil {
		return err
	}
	if len([]rune(string(labelBytes))) <= 1000 {
		return nil
	}

	for k, v := range req.Labels {
		runes := []rune(v)
		if len(runes) > 100 {
			req.Labels[k] = string(runes[:100]) + "..."
		}
	}
	return nil
}

func (r *Release) parseReleaseFile(req apistructs.ReleaseUploadRequest, file io.ReadCloser) (*dbclient.Release, []dbclient.Release, error) {
	var metadata apistructs.ReleaseMetadata
	dices := make(map[string]string)
	reader := tar.NewReader(file)
	hasMetadata := false
	for {
		hdr, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		buf := bytes.Buffer{}
		if _, err = io.Copy(&buf, reader); err != nil {
			return nil, nil, err
		}

		splits := strings.Split(hdr.Name, "/")
		if !strings.HasSuffix(splits[0], ".tar") {
			return nil, nil, errors.New("only support .tar file")
		}
		if len(splits) == 2 && splits[1] == "metadata.yml" {
			hasMetadata = true
			if err := yaml.Unmarshal(buf.Bytes(), &metadata); err != nil {
				return nil, nil, err
			}
		} else if len(splits) == 4 && splits[3] == "dice.yml" {
			appName := splits[2]
			dices[appName] = buf.String()
		}
	}

	if !hasMetadata {
		return nil, nil, errors.New("invalid file, metadata.yml not found")
	}
	if len(dices) == 0 {
		return nil, nil, errors.Errorf("invalid file, dice.yml not found")
	}

	projectReleaseID := uuid.UUID()

	appName2ID := make(map[string]uint64)
	apps, err := r.bdl.GetAppsByProject(uint64(req.ProjectID), uint64(req.OrgID), req.UserID)
	if err != nil {
		return nil, nil, errors.Errorf("failed to list apps, %v", err)
	}
	for i := range apps.List {
		appName2ID[apps.List[i].Name] = apps.List[i].ID
	}

	now := time.Now()

	var appReleaseList []string
	var appReleases []dbclient.Release
	for appName, dice := range dices {
		if _, ok := appName2ID[appName]; !ok {
			return nil, nil, errors.Errorf("app %s not existed", appName)
		}
		id := uuid.UUID()
		md := metadata.AppList[appName]
		existedReleases, err := r.db.GetReleasesByAppAndVersion(req.OrgID, req.ProjectID, int64(appName2ID[appName]), md.Version)
		if err != nil {
			return nil, nil, errors.Errorf("failed to get releases by app and version, %v", err)
		}
		if len(existedReleases) > 0 {
			oldDice, err := diceyml.New([]byte(existedReleases[0].Dice), true)
			if err != nil {
				return nil, nil, errors.Errorf("dice yml for release %s is invalid, %v", existedReleases[0].ReleaseID, err)
			}
			newDice, err := diceyml.New([]byte(dice), true)
			if err != nil {
				return nil, nil, errors.Errorf("dice yml for app %s release is invalid, %v", appName, err)
			}
			if !reflect.DeepEqual(oldDice.Obj(), newDice.Obj()) {
				return nil, nil, errors.Errorf("app release %s was already existed but has different dice yml", md.Version)
			}
			appReleaseList = append(appReleaseList, existedReleases[0].ReleaseID)
			continue
		}

		labels := map[string]string{
			"gitBranch":        md.GitBranch,
			"gitCommitId":      md.GitCommitID,
			"gitCommitMessage": md.GitCommitMessage,
			"gitRepo":          md.GitRepo,
		}
		data, err := json.Marshal(labels)
		if err != nil {
			return nil, nil, errors.Errorf("failed to marshal labels, %v", err)
		}

		appReleaseList = append(appReleaseList, id)
		appReleases = append(appReleases, dbclient.Release{
			ReleaseID:        id,
			ReleaseName:      md.GitBranch,
			Desc:             fmt.Sprintf("referenced by project release %s", projectReleaseID),
			Dice:             dice,
			Changelog:        md.ChangeLog,
			IsStable:         true,
			IsFormal:         false,
			IsProjectRelease: false,
			Labels:           string(data),
			Version:          md.Version,
			OrgID:            req.OrgID,
			ProjectID:        req.ProjectID,
			ProjectName:      req.ProjectName,
			UserID:           req.UserID,
			ClusterName:      req.ClusterName,
			Reference:        1,
			CreatedAt:        now,
			UpdatedAt:        now,
		})
	}

	listData, err := json.Marshal(appReleaseList)
	if err != nil {
		return nil, nil, errors.Errorf("faield to marshal application release list, %v", err)
	}
	projectRelease := &dbclient.Release{
		ReleaseID:              projectReleaseID,
		Desc:                   metadata.Desc,
		Changelog:              metadata.ChangeLog,
		IsStable:               true,
		IsFormal:               false,
		IsProjectRelease:       true,
		ApplicationReleaseList: string(listData),
		Version:                metadata.Version,
		OrgID:                  req.OrgID,
		ProjectID:              req.ProjectID,
		ProjectName:            req.ProjectName,
		UserID:                 req.UserID,
		ClusterName:            req.ClusterName,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	return projectRelease, appReleases, nil
}

func (r *Release) createReleases(releases []dbclient.Release) (err error) {
	tx := r.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for i := range releases {
		if err = tx.Create(releases[i]).Error; err != nil {
			return
		}
	}
	return tx.Commit().Error
}

// CreateByFile 用文件创建项目制品
func (r *Release) CreateByFile(req apistructs.ReleaseUploadRequest, file io.ReadCloser) (string, string, error) {
	projectRelease, appReleases, err := r.parseReleaseFile(req, file)
	if err != nil {
		return "", "", err
	}

	releases, err := r.db.GetReleasesByProjectAndVersion(req.OrgID, req.ProjectID, projectRelease.Version)
	if err != nil {
		return "", "", err
	}
	if len(releases) > 0 {
		return "", "", errors.Errorf("release version: %s already exist", projectRelease.Version)
	}

	err = r.createReleases(append(appReleases, *projectRelease))
	if err != nil {
		return "", "", err
	}
	return projectRelease.Version, projectRelease.ReleaseID, nil
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

	if release.IsFormal {
		return errors.New("formal release can not be updated")
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

	// update changelog
	if req.Changelog != "" {
		release.Changelog = req.Changelog
	}

	if release.IsProjectRelease {
		if len(req.ApplicationReleaseList) == 0 {
			return errors.New("application release list can not be null for project release")
		}
		newList := strutil.DedupSlice(req.ApplicationReleaseList)
		sort.Strings(newList)
		newAppReleases, err := r.db.GetReleases(req.ApplicationReleaseList)
		if err != nil {
			return errors.Errorf("failed to get application releases: %v", err)
		}

		// update application_release_list
		oldListStr := release.ApplicationReleaseList
		newListData, err := json.Marshal(req.ApplicationReleaseList)
		if err != nil {
			return errors.Errorf("failed to marshal release list, %v", err)
		}
		release.ApplicationReleaseList = string(newListData)

		// update reference count
		var oldList []string
		oldList, err = unmarshalApplicationReleaseList(oldListStr)
		if err != nil {
			return errors.Errorf("failed to json unmarshal release list, %v", err)
		}

		if isSliceEqual(newList, oldList) {
			if err := r.db.UpdateRelease(release); err != nil {
				return err
			}
		} else {
			oldAppReleases, err := r.db.GetReleases(oldList)
			if err != nil {
				return err
			}
			if err = r.updateProjectReleaseAndReference(release, oldAppReleases, newAppReleases); err != nil {
				return err
			}
		}
	} else {
		if err := r.db.UpdateRelease(release); err != nil {
			return err
		}
	}

	// Send release update event to eventbox
	event.SendReleaseEvent(event.ReleaseEventUpdate, release)

	return nil
}

func (r *Release) updateProjectReleaseAndReference(release *dbclient.Release, oldAppReleases, newAppReleases []dbclient.Release) (err error) {
	type pair = struct {
		release  *dbclient.Release
		deltaRef int
	}
	idToRelease := make(map[string]pair)
	for i := range oldAppReleases {
		idToRelease[oldAppReleases[i].ReleaseID] = pair{&oldAppReleases[i], -1}
	}
	for i := range newAppReleases {
		p := idToRelease[newAppReleases[i].ReleaseID]
		idToRelease[newAppReleases[i].ReleaseID] = pair{&newAppReleases[i], p.deltaRef + 1}
	}

	tx := r.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for _, p := range idToRelease {
		if p.deltaRef == 0 {
			continue
		}
		release := p.release
		if p.deltaRef < 0 {
			release.Reference--
		} else {
			release.Reference++
		}
		if err = tx.Save(release).Error; err != nil {
			return
		}
	}

	if err = tx.Save(release).Error; err != nil {
		return
	}
	return tx.Commit().Error
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
func (r *Release) Delete(orgID int64, releaseIDs ...string) error {
	var failed []string
	for _, releaseID := range releaseIDs {
		release, err := r.db.GetRelease(releaseID)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s(%s)", releaseID, err.Error()))
			continue
		}
		if orgID != 0 && release.OrgID != orgID {
			failed = append(failed, fmt.Sprintf("%s(%s)", releaseID, "not fount"))
			continue
		}

		if release.IsFormal {
			failed = append(failed, fmt.Sprintf("%s(%s)", releaseID, "formal release can not be deleted"))
		}
		// Release被使用时，不可删除
		if release.Reference > 0 {
			failed = append(failed, fmt.Sprintf("%s(%s)", releaseID, "release was referenced"))
			logrus.Errorf("failed to delete release %s, which was referenced", releaseID)
			continue
		}

		images, err := r.imageDB.GetImagesByRelease(releaseID)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s(%s)", releaseID, "failed to get images"+err.Error()))
			continue
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
		if release.IsProjectRelease && release.ApplicationReleaseList != "" {
			list, err := unmarshalApplicationReleaseList(release.ApplicationReleaseList)
			if err != nil {
				failed = append(failed, fmt.Sprintf("%s(%s)", releaseID, "failed to unmarshal release list, "+err.Error()))
				continue
			}
			appReleases, err := r.db.GetReleases(list)
			if err != nil {
				failed = append(failed, fmt.Sprintf("%s(%s)", releaseID, "failed to get app releases, "+err.Error()))
				continue
			}
			if err = r.deleteReleaseAndUpdateReference(release, appReleases); err != nil {
				failed = append(failed, fmt.Sprintf("%s(%s)", releaseID, err.Error()))
				continue
			}
		} else {
			if err := r.db.DeleteRelease(releaseID); err != nil {
				failed = append(failed, fmt.Sprintf("%s(%s)", releaseID, err.Error()))
				logrus.Errorf("failed to delete release %s, %v", releaseID, err)
				continue
			}
		}

		// send release delete event to eventbox
		event.SendReleaseEvent(event.ReleaseEventDelete, release)
	}
	if len(failed) != 0 {
		return errors.New(strings.Join(failed, ", "))
	}
	return nil
}

func (r *Release) deleteReleaseAndUpdateReference(release *dbclient.Release, appReleases []dbclient.Release) (err error) {
	tx := r.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for i := range appReleases {
		appReleases[i].Reference--
		if err = tx.Save(appReleases[i]).Error; err != nil {
			return
		}
	}

	if err = tx.Delete(release).Error; err != nil {
		return
	}
	return tx.Commit().Error
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

	return r.convertToGetReleaseResponse(release)
}

// List 根据搜索条件进行搜索
func (r *Release) List(orgID int64, req *apistructs.ReleaseListRequest) (*apistructs.ReleaseListResponseData, error) {
	total, releases, err := r.db.GetReleasesByParams(
		orgID, req.ProjectID, req.ApplicationID,
		req.Query, req.ReleaseName, req.Branch,
		req.IsStable, req.IsFormal, req.IsProjectRelease,
		req.UserID, req.Version, req.CommitID, req.Tags,
		req.Cluster, req.CrossCluster, req.IsVersion,
		req.CrossClusterOrSpecifyCluster,
		req.StartTime, req.EndTime, req.PageNum, req.PageSize,
		req.OrderBy, req.Order)
	if err != nil {
		return nil, err
	}

	releaseList := make([]apistructs.ReleaseData, 0, len(releases))
	for _, v := range releases {
		release, err := r.convertToListReleaseResponse(&v)
		if err != nil {
			logrus.WithField("func", "*ReleaseList").Errorln("failed to convertToListReleaseResponse")
			continue
		}
		releaseList = append(releaseList, *release)
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

	releaseData, err := r.convertToListReleaseResponse(release)
	if err != nil {
		return "", err
	}
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

// Convert 从ReleaseRequest中提取Release元信息, 若为应用级制品, appReleases填nil
func (r *Release) Convert(releaseRequest *apistructs.ReleaseCreateRequest, appReleases []dbclient.Release) (*dbclient.Release, error) {
	release := dbclient.Release{
		ReleaseID:        uuid.UUID(),
		ReleaseName:      releaseRequest.ReleaseName,
		Desc:             releaseRequest.Desc,
		Dice:             releaseRequest.Dice,
		Addon:            releaseRequest.Addon,
		Changelog:        releaseRequest.Changelog,
		IsStable:         releaseRequest.IsStable,
		IsFormal:         releaseRequest.IsFormal,
		IsProjectRelease: releaseRequest.IsProjectRelease,
		Version:          releaseRequest.Version,
		OrgID:            releaseRequest.OrgID,
		ProjectID:        releaseRequest.ProjectID,
		ApplicationID:    releaseRequest.ApplicationID,
		ProjectName:      releaseRequest.ProjectName,
		ApplicationName:  releaseRequest.ApplicationName,
		UserID:           releaseRequest.UserID,
		ClusterName:      releaseRequest.ClusterName,
		CrossCluster:     releaseRequest.CrossCluster,
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
	}

	if release.ProjectName == "" {
		project, err := r.bdl.GetProject(uint64(release.ProjectID))
		if err != nil {
			return nil, err
		}
		release.ProjectName = project.Name
	}

	if len(releaseRequest.Labels) > 0 {
		labelBytes, err := json.Marshal(releaseRequest.Labels)
		if err != nil {
			return nil, err
		}
		release.Labels = string(labelBytes)
	}

	if len(releaseRequest.Tags) > 0 {
		tagBytes, err := json.Marshal(releaseRequest.Tags)
		if err != nil {
			return nil, err
		}
		release.Tags = string(tagBytes)
	}

	if len(releaseRequest.Resources) > 0 {
		resourceBytes, err := json.Marshal(releaseRequest.Resources)
		if err != nil {
			return nil, err
		}
		release.Resources = string(resourceBytes)
	}

	if releaseRequest.IsProjectRelease {
		if len(appReleases) == 0 {
			return nil, errors.New("application release list can not be null for project release when dice yaml is empty")
		}

		selectedApp := make(map[int64]struct{})
		for i := range appReleases {
			if _, ok := selectedApp[appReleases[i].ApplicationID]; ok {
				return nil, errors.New("one application can only be selected once")
			}
			selectedApp[appReleases[i].ApplicationID] = struct{}{}
		}
		release.ApplicationID = 0
		release.ApplicationName = ""
		release.Dice = ""

		listData, err := json.Marshal(releaseRequest.ApplicationReleaseList)
		if err != nil {
			return nil, errors.Errorf("failed to marshal release list, %v", err)
		}
		release.ApplicationReleaseList = string(listData)
	}
	return &release, nil
}

func (r *Release) ToFormal(releaseIDs []string) error {
	releases, err := r.db.GetReleases(releaseIDs)
	if err != nil {
		return err
	}
	var failed []string
	for i := range releases {
		if !releases[i].IsStable {
			failed = append(failed, fmt.Sprintf("%s(%s)", releases[i].ReleaseID, "stable release can not be formaled"))
		}
		if err := r.formalRelease(&releases[i]); err != nil {
			failed = append(failed, fmt.Sprintf("%s(%s)", releases[i].ReleaseID, err.Error()))
		}
	}
	if len(failed) == 0 {
		return nil
	}
	return errors.Errorf("failed to formal releases: %s", strings.Join(failed, ","))
}

func (r *Release) formalRelease(release *dbclient.Release) (err error) {
	tx := r.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if release.IsProjectRelease {
		var list []string
		if err = json.Unmarshal([]byte(release.ApplicationReleaseList), &list); err != nil {
			return err
		}

		for _, id := range list {
			if err = tx.Model(&dbclient.Release{}).Where("release_id = ?", id).Updates(map[string]interface{}{
				"is_stable": true,
				"is_formal": true,
			}).Error; err != nil {
				return
			}
		}
	}

	if err = tx.Model(&dbclient.Release{}).Where("release_id = ?", release.ReleaseID).Update("is_formal", true).Error; err != nil {
		return
	}
	return tx.Commit().Error
}

// release数据库结构转换为API返回所需结构
func (r *Release) convertToGetReleaseResponse(release *dbclient.Release) (*apistructs.ReleaseGetResponseData, error) {
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

	var summary []*apistructs.ApplicationReleaseSummary
	if release.IsProjectRelease {
		var appReleaseIDs []string
		if err = json.Unmarshal([]byte(release.ApplicationReleaseList), &appReleaseIDs); err != nil {
			return nil, errors.Errorf("failed to Unmarshal appReleaseIDs")
		}

		releases, err := r.db.GetReleases(appReleaseIDs)
		if err != nil {
			logrus.Errorf("failed to get app releases for release %s", release.ReleaseID)
		}
		for i := range releases {
			summary = append(summary, &apistructs.ApplicationReleaseSummary{
				ReleaseID:       releases[i].ReleaseID,
				ReleaseName:     releases[i].ReleaseName,
				Version:         releases[i].Version,
				ApplicationID:   releases[i].ApplicationID,
				ApplicationName: releases[i].ApplicationName,
				CreatedAt:       releases[i].CreatedAt.Format("2006/01/02 15:04:05"),
				DiceYml:         releases[i].Dice,
			})
		}
	}

	respData := &apistructs.ReleaseGetResponseData{
		ReleaseID:              release.ReleaseID,
		ReleaseName:            release.ReleaseName,
		Diceyml:                release.Dice,
		Desc:                   release.Desc,
		Addon:                  release.Addon,
		Changelog:              release.Changelog,
		IsStable:               release.IsStable,
		IsFormal:               release.IsFormal,
		IsProjectRelease:       release.IsProjectRelease,
		ApplicationReleaseList: summary,
		Resources:              resources,
		Labels:                 labels,
		Tags:                   release.Tags,
		Version:                release.Version,
		CrossCluster:           release.CrossCluster,
		Reference:              release.Reference,
		OrgID:                  release.OrgID,
		ProjectID:              release.ProjectID,
		ApplicationID:          release.ApplicationID,
		ProjectName:            release.ProjectName,
		ApplicationName:        release.ApplicationName,
		UserID:                 release.UserID,
		ClusterName:            release.ClusterName,
		CreatedAt:              release.CreatedAt,
		UpdatedAt:              release.UpdatedAt,
	}
	if err = respData.ReLoadImages(); err != nil {
		logrus.WithError(err).Errorln("failed to ReLoadImages")
		return nil, err
	}

	return respData, nil
}

// release数据库结构转换为API返回列表所需结构
func (r *Release) convertToListReleaseResponse(release *dbclient.Release) (*apistructs.ReleaseData, error) {
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

	respData := &apistructs.ReleaseData{
		ReleaseID:              release.ReleaseID,
		ReleaseName:            release.ReleaseName,
		Diceyml:                release.Dice,
		Desc:                   release.Desc,
		Addon:                  release.Addon,
		Changelog:              release.Changelog,
		IsStable:               release.IsStable,
		IsFormal:               release.IsFormal,
		IsProjectRelease:       release.IsProjectRelease,
		ApplicationReleaseList: release.ApplicationReleaseList,
		Resources:              resources,
		Labels:                 labels,
		Tags:                   release.Tags,
		Version:                release.Version,
		CrossCluster:           release.CrossCluster,
		Reference:              release.Reference,
		OrgID:                  release.OrgID,
		ProjectID:              release.ProjectID,
		ApplicationID:          release.ApplicationID,
		ProjectName:            release.ProjectName,
		ApplicationName:        release.ApplicationName,
		UserID:                 release.UserID,
		ClusterName:            release.ClusterName,
		CreatedAt:              release.CreatedAt,
		UpdatedAt:              release.UpdatedAt,
	}
	return respData, nil
}

// GetImages 从ReleaseRequest中提取Image信息
func (r *Release) GetImages(dices []string) []*imagedb.Image {
	existed := make(map[string]struct{})
	images := make([]*imagedb.Image, 0)
	for _, yml := range dices {
		var dice diceyml.Object
		err := yaml.Unmarshal([]byte(yml), &dice)
		if err != nil {
			continue
		}

		// Get images from dice.yml
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
			if _, ok := existed[image.Image]; !ok {
				images = append(images, image)
				existed[image.Image] = struct{}{}
			}
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
			if _, ok := existed[image.Image]; !ok {
				images = append(images, image)
				existed[image.Image] = struct{}{}
			}
		}
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

func unmarshalApplicationReleaseList(str string) ([]string, error) {
	var list []string
	if err := json.Unmarshal([]byte(str), &list); err != nil {
		return nil, err
	}
	return list, nil
}

func isSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
