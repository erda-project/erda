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

package filesvc

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

func (svc *FileService) CleanExpiredFiles(_expiredAt ...time.Time) error {
	// 获取过期时间
	expiredAt := time.Unix(time.Now().Unix(), 0)
	if len(_expiredAt) > 0 {
		expiredAt = _expiredAt[0]
	}

	// 获取过期文件列表
	files, err := svc.db.ListExpiredFiles(expiredAt)
	if err != nil {
		logrus.Errorf("[alert] failed to list expired files, expiredBefore: %s, err: %v", expiredAt.Format(time.RFC3339), err)
		return apierrors.ErrCleanExpiredFile.InternalError(err)
	}

	// 遍历删除文件
	for _, file := range files {
		if err := svc.DeleteFile(file); err != nil {
			logrus.Errorf("[alert] failed to clean expired file, fileUUID: %s, err: %v", file.UUID, err)
			continue
		}
	}

	return nil
}
