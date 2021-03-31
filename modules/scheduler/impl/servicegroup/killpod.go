package servicegroup

import (
	"context"

	"github.com/erda-project/erda/apistructs"
)

func (s ServiceGroupImpl) KillPod(ctx context.Context, namespace string, name string, containerID string) error {
	sg := apistructs.ServiceGroup{}
	if err := s.js.Get(context.Background(), mkServiceGroupKey(namespace, name), &sg); err != nil {
		return err
	}

	_, err := s.handleKillPod(ctx, &sg, containerID)
	if err != nil {
		return err
	}
	return nil
}
