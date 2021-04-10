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

package dbclient

import (
	"github.com/erda-project/erda/pkg/dbengine"
)

type Image struct {
	dbengine.BaseModel
	ReleaseID string `json:"releaseId" gorm:"index:idx_release_id"`       // 关联release
	ImageName string `json:"imageName" gorm:"type:varchar(128);not null"` // 镜像名称
	ImageTag  string `json:"imageTag" gorm:"type:varchar(64)"`            // 镜像Tag
	Image     string `json:"image" gorm:"not null"`                       // 镜像地址
}

// Set table name
func (Image) TableName() string {
	return "ps_images"
}

// CreateImage 创建镜像
func (client *DBClient) CreateImage(image *Image) error {
	return client.Create(image).Error
}

// UpdateImage 更新镜像
func (client *DBClient) UpdateImage(image *Image) error {
	return client.Save(image).Error
}

// DeleteImage 删除镜像
func (client *DBClient) DeleteImage(imageID int64) error {
	return client.Where("id = ?", imageID).Delete(&Image{}).Error
}

// GetImagesByRelease 根据 releaseID 获取镜像列表
func (client *DBClient) GetImagesByRelease(releaseID string) ([]Image, error) {
	var images []Image
	if err := client.Where("release_id = ?", releaseID).Find(&images).Error; err != nil {
		return nil, err
	}
	return images, nil
}

// GetImagesByID 根据 imageID 获取镜像
func (client *DBClient) GetImageByID(imageID int64) (*Image, error) {
	var image Image
	if err := client.Where("id = ?", imageID).Find(&image).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImageByImage 根据 image 获取镜像
func (client *DBClient) GetImageByImage(image string) (*Image, error) {
	var img Image
	if err := client.Where("image = ?", image).First(&img).Error; err != nil {
		return nil, err
	}
	return &img, nil
}

// ListImages 获取镜像列表
func (client *DBClient) ListImages(orgID, pageNo, pageSize int64) (int64, []Image, error) {
	var (
		total  int64
		images []Image
	)
	if orgID == 0 { // Token认证方式
		if err := client.Order("updated_at DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).
			Find(&images).Error; err != nil {
			return 0, nil, err
		}
		if err := client.Model(&Image{}).Count(&total).Error; err != nil {
			return 0, nil, err
		}
	} else {
		if err := client.Where("org_id = ?", orgID).Order("updated_at DESC").
			Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&images).Error; err != nil {
			return 0, nil, err
		}
		if err := client.Where("org_id = ?", orgID).Model(&Image{}).Count(&total).Error; err != nil {
			return 0, nil, err
		}
	}
	return total, images, nil
}

// GetImageCount 获取不在给定 releaseID 下的image数目
func (client *DBClient) GetImageCount(releaseID, image string) (int64, error) {
	var count int64
	if err := client.Where("image = ?", image).Not("release_id", releaseID).
		Model(&Image{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
