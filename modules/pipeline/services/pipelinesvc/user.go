package pipelinesvc

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

// tryGetUser try to get user info from cmdb. If failed, return a basic user just with id.
// TODO later add cache here if need.
func (s *PipelineSvc) tryGetUser(userID string) *apistructs.PipelineUser {
	user, err := s.bdl.GetCurrentUser(userID)
	if err != nil {
		logrus.Warnf("failed to get user info, userID: %s, err: %v", userID, err)
		// return basic user just with id
		return &apistructs.PipelineUser{ID: userID}
	}
	if user == nil {
		logrus.Warnf("failed to get user info, userID: %s, err: %v", userID, fmt.Errorf("get empty user info"))
		// return basic user just with id
		return &apistructs.PipelineUser{ID: userID}
	}
	// return queried user
	return user.ConvertToPipelineUser()
}
