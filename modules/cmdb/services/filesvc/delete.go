package filesvc

import (
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
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
