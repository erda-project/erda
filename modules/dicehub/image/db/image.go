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

package db

import "github.com/jinzhu/gorm"

// ImageConfig .
type ImageConfigDB struct {
	*gorm.DB
}

// CreateImage
func (client *ImageConfigDB) CreateImage(image *Image) error {
	return client.Create(image).Error
}

// UpdateImage
func (client *ImageConfigDB) UpdateImage(image *Image) error {
	return client.Save(image).Error
}

// DeleteImage
func (client *ImageConfigDB) DeleteImage(imageID int64) error {
	return client.Where("id = ?", imageID).Delete(&Image{}).Error
}

// GetImagesByRelease Get image list by releaseID
func (client *ImageConfigDB) GetImagesByRelease(releaseID string) ([]Image, error) {
	var images []Image
	if err := client.Where("release_id = ?", releaseID).Find(&images).Error; err != nil {
		return nil, err
	}
	return images, nil
}

// GetImagesByID Get image by imageID
func (client *ImageConfigDB) GetImageByID(imageID int64) (*Image, error) {
	var image Image
	if err := client.Where("id = ?", imageID).Find(&image).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImageByImage Get image by image string
func (client *ImageConfigDB) GetImageByImage(image string) (*Image, error) {
	var img Image
	if err := client.Where("image = ?", image).First(&img).Error; err != nil {
		return nil, err
	}
	return &img, nil
}

// ListImages Get image list
func (client *ImageConfigDB) ListImages(orgID, pageNo, pageSize int64) (int64, []Image, error) {
	var (
		total  int64
		images []Image
	)
	if orgID == 0 { // verification method Token
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

// GetImageCount Get the quantity not equal releaseID
func (client *ImageConfigDB) GetImageCount(releaseID, image string) (int64, error) {
	var count int64
	if err := client.Where("image = ?", image).Not("release_id", releaseID).
		Model(&Image{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
