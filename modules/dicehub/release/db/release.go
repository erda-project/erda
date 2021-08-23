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

package db

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/strutil"
)

// ReleaseConfig .
type ReleaseConfigDB struct {
	*gorm.DB
}

// CreateRelease Create Release
func (client *ReleaseConfigDB) CreateRelease(release *Release) error {
	return client.Create(release).Error
}

// UpdateRelease Update Release
func (client *ReleaseConfigDB) UpdateRelease(release *Release) error {
	return client.Save(release).Error
}

// DeleteRelease Delete Release
func (client *ReleaseConfigDB) DeleteRelease(releaseID string) error {
	return client.Where("release_id = ?", releaseID).Delete(&Release{}).Error
}

// GetRelease Get Release
func (client *ReleaseConfigDB) GetRelease(releaseID string) (*Release, error) {
	var release Release
	if err := client.Where("release_id = ?", releaseID).Find(&release).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, dbengine.ErrNotFound
		}
		return nil, err
	}
	return &release, nil
}

// GetReleasesByParams filter Releases by parameters
func (client *ReleaseConfigDB) GetReleasesByParams(
	orgID, projectID, applicationID int64,
	keyword, releaseName, branch,
	cluster string, crossCluster string, isVersion bool, crossClusterOrSpecifyCluster string,
	startTime, endTime time.Time, pageNum, pageSize int64) (int64, []Release, error) {

	var releases []Release
	db := client.DB.Debug()
	if orgID > 0 {
		db = db.Where("org_id = ?", orgID)
	}
	if applicationID > 0 {
		db = db.Where("application_id = ?", applicationID)
	}

	if projectID > 0 {
		db = db.Where("project_id = ?", projectID)
	}
	if keyword != "" {
		db = db.Where("release_id LIKE ? or release_name LIKE ? or version LIKE ?", "%"+keyword+"%",
			"%"+keyword+"%", "%"+keyword+"%")
	} else if releaseName != "" {
		db = db.Where("release_name = ?", releaseName)
	}
	if isVersion {
		db = db.Not("version", "")
	}

	if cluster != "" {
		db = db.Where("cluster_name = ?", cluster)
	}
	if crossCluster != "" {
		db = db.Where("cross_cluster = ?", crossCluster)
	}
	if crossClusterOrSpecifyCluster != "" {
		db = db.Where("(cluster_name = ? AND cross_cluster = 0) OR cross_cluster = 1", crossClusterOrSpecifyCluster)
	}
	if branch != "" {
		db = db.Where("labels LIKE ?", "%"+fmt.Sprintf("\"gitBranch\":\"%s\"", branch)+"%")
	}

	if !startTime.IsZero() {
		db = db.Where("created_at > ?", startTime)
	}

	if err := db.Where("created_at <= ?", endTime).Order("created_at DESC").Offset((pageNum - 1) * pageSize).
		Limit(pageSize).Find(&releases).Error; err != nil {
		return 0, nil, err
	}

	// Get total
	var total int64
	if err := db.Where("created_at <= ?", endTime).Model(&Release{}).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, releases, nil
}

// GetReleasesByAppAndVersion Get Release list by appID & version
func (client *ReleaseConfigDB) GetReleasesByAppAndVersion(orgID, projectID, appID int64, version string) ([]Release, error) {
	var releases []Release
	if err := client.Where("org_id = ?", orgID).
		Where("project_id = ?", projectID).
		Where("application_id = ?", appID).
		Where("version = ?", version).
		Find(&releases).Error; err != nil {
		return nil, err
	}
	return releases, nil
}

// GetReleaseNamesByApp Get releaseName list by appID
func (client *ReleaseConfigDB) GetReleaseNamesByApp(orgID, appID int64) ([]string, error) {
	var releaseNames []string
	if orgID == 0 {
		if err := client.Select("release_name").
			Where("application_id = ?", appID).
			Group("release_name").Find(&releaseNames).Error; err != nil {
			return nil, err
		}
	} else {
		if err := client.Select("release_name").
			Where("org_id = ?", orgID).
			Where("application_id = ?", appID).
			Group("release_name").Find(&releaseNames).Error; err != nil {
			return nil, err
		}
	}

	return releaseNames, nil
}

// GetAppIDsByProjectAndVersion Get appIDList by projectID & version
func (client *ReleaseConfigDB) GetAppIDsByProjectAndVersion(projectID int64, version string) ([]int64, error) {
	var appIDs []int64
	if err := client.Select([]string{"application_id"}).
		Where("project_id = ?", projectID).
		Where("version LIKE ?", strutil.Concat(version, "%")).
		Group("application_id").Find(&appIDs).Error; err != nil {
		return nil, err
	}
	return appIDs, nil
}

// GetLatestReleaseByAppAndVersion Get the latest release under the app
func (client *ReleaseConfigDB) GetLatestReleaseByAppAndVersion(appID int64, version string) (*Release, error) {
	var release Release
	if err := client.Where("application_id = ?", appID).
		Where("version LIKE ?", strutil.Concat(version, "%")).
		Order("created_at DESC").
		Limit(1).Find(&release).Error; err != nil {
		return nil, err
	}
	return &release, nil
}

// GetUnReferedReleasesBefore Get the Release that has not been referenced before a given point in time
func (client *ReleaseConfigDB) GetUnReferedReleasesBefore(before time.Time) ([]Release, error) {
	var releases []Release
	if err := client.Where("reference <= ?", 0).Where("updated_at < ?", before).
		Order("updated_at").Find(&releases).Error; err != nil {
		return nil, err
	}
	return releases, nil
}
