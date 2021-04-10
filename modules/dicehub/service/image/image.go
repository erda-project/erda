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

package image

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dicehub/dbclient"
	"github.com/erda-project/erda/pkg/strutil"
)

// Image Image 操作封装
type Image struct {
	db  *dbclient.DBClient
	bdl *bundle.Bundle
}

// Option 定义 Image 对象的配置选项
type Option func(*Image)

// New 新建 Image 实例，操作 Image 资源
func New(options ...Option) *Image {
	app := &Image{}
	for _, op := range options {
		op(app)
	}
	return app
}

// WithDBClient 配置 db client
func WithDBClient(db *dbclient.DBClient) Option {
	return func(a *Image) {
		a.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(a *Image) {
		a.bdl = bdl
	}
}

// Create 创建镜像
func (i *Image) Create(req *apistructs.ImageCreateRequest) (int64, error) {
	image := dbclient.Image{
		ReleaseID: req.ReleaseID,
		ImageName: req.ImageName,
		ImageTag:  req.ImageTag,
		Image:     req.Image,
	}
	if err := i.db.CreateImage(&image); err != nil {
		return 0, err
	}

	return int64(image.ID), nil
}

// Update 更新镜像
func (i *Image) Update(ImageIDOrImage string, req *apistructs.ImageUpdateRequest) error {
	image, err := i.Get(ImageIDOrImage)
	if err != nil {
		return err
	}
	image.Image = req.Body.Image
	image.ImageName = req.Body.ImageName
	image.ImageTag = req.Body.ImageTag

	return i.db.UpdateImage(image)
}

// Get 获取镜像
func (i *Image) Get(imageIDOrImage string) (*dbclient.Image, error) {
	idOrImage, err := strutil.Atoi64(imageIDOrImage)
	if err == nil {
		image, err := i.db.GetImageByID(idOrImage)
		if err != nil {
			return nil, err
		}
		return image, nil
	}
	image, err := i.db.GetImageByImage(imageIDOrImage)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (i *Image) List(orgID, pageNo, pageSize int64) (*apistructs.ImageListResponseData, error) {
	total, images, err := i.db.ListImages(orgID, pageNo, pageSize)
	if err != nil {
		return nil, err
	}

	imgs := make([]apistructs.ImageGetResponseData, 0, len(images))
	for _, v := range images {
		imgs = append(imgs, *i.convert(&v))
	}

	return &apistructs.ImageListResponseData{
		Total: total,
		List:  imgs,
	}, nil
}

func (i *Image) convert(image *dbclient.Image) *apistructs.ImageGetResponseData {
	return &apistructs.ImageGetResponseData{
		ID:        int64(image.ID),
		ReleaseID: image.ReleaseID,
		ImageName: image.ImageName,
		ImageTag:  image.ImageTag,
		Image:     image.Image,
		CreatedAt: image.CreatedAt,
		UpdatedAt: image.UpdatedAt,
	}
}
