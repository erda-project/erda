package models

import (
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/gittar/pkg/gitmodule"
)

func (svc *Service) TriggerEvent(repository *gitmodule.Repository, eventName string, contentData interface{}) {
	err := svc.bundle.CreateEvent(&apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			ApplicationID: strconv.FormatInt(repository.ApplicationId, 10),
			ProjectID:     strconv.FormatInt(repository.ProjectId, 10),
			OrgID:         strconv.FormatInt(repository.OrgId, 10),
			Event:         eventName,
			Action:        eventName,
		},
		Sender:  "gittar",
		Content: contentData,
	})
	if err != nil {
		logrus.Errorf("failed to trigger event %s err:%s", eventName, err)
	}
}
