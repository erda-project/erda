package servicegroup

import (
	"context"
	"fmt"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/impl/clusterinfo"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (s ServiceGroupImpl) Update(req apistructs.ServiceGroupUpdateV2Request) (apistructs.ServiceGroup, error) {
	sg, err := convertServiceGroupUpdateV2Request(req, s.clusterinfo)
	if err != nil {
		return apistructs.ServiceGroup{}, err
	}

	oldSg := apistructs.ServiceGroup{}
	if err := s.js.Get(context.Background(), mkServiceGroupKey(sg.Type, sg.ID), &oldSg); err != nil {
		return apistructs.ServiceGroup{}, fmt.Errorf("Cannot get servicegroup(%s/%s) from etcd, err: %v", sg.Type, sg.ID, err)
	}
	diffAndPatchRuntime(&sg, &oldSg)

	if err := s.js.Put(context.Background(), mkServiceGroupKey(sg.Type, sg.ID), &oldSg); err != nil {
		return apistructs.ServiceGroup{}, err
	}

	oldSg.Labels = appendServiceTags(oldSg.Labels, oldSg.Executor)
	if _, err := s.handleServiceGroup(context.Background(), &oldSg, task.TaskUpdate); err != nil {
		return apistructs.ServiceGroup{}, err
	}
	return oldSg, nil
}

func convertServiceGroupUpdateV2Request(req apistructs.ServiceGroupUpdateV2Request, clusterinfo clusterinfo.ClusterInfo) (apistructs.ServiceGroup, error) {
	return convertServiceGroup(apistructs.ServiceGroupCreateV2Request(req), clusterinfo)
}

func diffAndPatchRuntime(newsg *apistructs.ServiceGroup, oldsg *apistructs.ServiceGroup) {
	// generate LastModifiedTime according to current time
	oldsg.LastModifiedTime = time.Now().Unix()

	oldsg.Labels = newsg.Labels
	oldsg.ServiceDiscoveryKind = newsg.ServiceDiscoveryKind

	// TODO: refactor it, separate data and status into different etcd key
	// 全量更新
	oldsg.Services = newsg.Services
}
