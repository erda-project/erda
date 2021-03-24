package appsvc

import (
	"github.com/erda-project/erda/apistructs"
)

func (s *AppSvc) GetApp(appID uint64) (*apistructs.ApplicationDTO, error) {
	return s.bdl.GetApp(appID)
}
