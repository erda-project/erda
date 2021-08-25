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

package filesvc

import (
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
)

func (svc *FileService) DeleteFile(file dao.File) error {
	// delete db record
	if err := svc.db.DeleteFile(uint64(file.ID)); err != nil {
		return apierrors.ErrDeleteFile.InternalError(err)
	}

	// delete file in storage
	storager := svc.GetStorage(file.StorageType)
	if err := storager.Delete(file.FullRelativePath); err != nil {
		return apierrors.ErrDeleteFile.InternalError(err)
	}

	return nil
}
