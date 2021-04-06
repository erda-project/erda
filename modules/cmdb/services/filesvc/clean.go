package filesvc

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
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
