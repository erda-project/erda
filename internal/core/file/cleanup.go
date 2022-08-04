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

package file

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/pkg/safe"
	"github.com/erda-project/erda/internal/core/legacy/services/apierrors"
)

func (p *provider) asyncCleanupExpiredFiles() {
	// clean expired files
	safe.Go(func() {
		ticker := time.NewTicker(p.Cfg.Cleanup.ExpiredFilesInterval)
		for range ticker.C {
			_ = p.cleanExpiredFiles()
		}
	})
}

func (p *provider) cleanExpiredFiles(_expiredAt ...time.Time) error {
	// fetch expired time
	expiredAt := time.Unix(time.Now().Unix(), 0)
	if len(_expiredAt) > 0 {
		expiredAt = _expiredAt[0]
	}

	// fetch expired files
	files, err := p.db.ListExpiredFiles(expiredAt)
	if err != nil {
		logrus.Errorf("[alert] failed to list expired files, expiredBefore: %s, err: %v", expiredAt.Format(time.RFC3339), err)
		return apierrors.ErrCleanExpiredFile.InternalError(err)
	}

	// iterate delete files
	for _, file := range files {
		if err := p.fileService.DeleteFile(file); err != nil {
			logrus.Errorf("[alert] failed to clean expired file, fileUUID: %s, err: %v", file.UUID, err)
			continue
		}
	}

	return nil
}
