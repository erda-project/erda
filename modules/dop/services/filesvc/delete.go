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
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

func (svc *FileService) DeleteFile(file dao.File) error {
	// delete db record
	if err := svc.db.DeleteFile(uint64(file.ID)); err != nil {
		return apierrors.ErrDeleteFile.InternalError(err)
	}

	// delete file in storage
	storager := getStorage(file.StorageType)
	if err := storager.Delete(file.FullRelativePath); err != nil {
		return apierrors.ErrDeleteFile.InternalError(err)
	}

	return nil
}
