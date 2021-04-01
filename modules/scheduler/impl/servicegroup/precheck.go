package servicegroup

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (s ServiceGroupImpl) Precheck(req apistructs.ServiceGroupPrecheckRequest) (apistructs.ServiceGroupPrecheckData, error) {
	sg, err := convertServiceGroupCreateV2Request(apistructs.ServiceGroupCreateV2Request(req), s.clusterinfo)
	if err != nil {
		return apistructs.ServiceGroupPrecheckData{}, err
	}
	t, err := s.handleServiceGroup(context.Background(), &sg, task.TaskPrecheck)
	if err != nil {
		return apistructs.ServiceGroupPrecheckData{}, err
	}

	return t.Extra.(apistructs.ServiceGroupPrecheckData), nil
}
