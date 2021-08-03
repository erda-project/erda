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

package release

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/modules/dicehub/release/db"
	"github.com/erda-project/erda/modules/dicehub/service/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
)

type releaseGetDiceService struct {
	p  *provider
	db *db.ReleaseConfigDB
}

func (s *releaseGetDiceService) PullDiceYAML(ctx context.Context, req *pb.ReleaseGetDiceYmlRequest) (*pb.ReleaseGetDiceYmlResponse, error) {
	return s.GetDiceYAML(ctx, req)
}

func (s *releaseGetDiceService) GetDiceYAML(ctx context.Context, req *pb.ReleaseGetDiceYmlRequest) (*pb.ReleaseGetDiceYmlResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, apierrors.ErrGetYAML.NotLogin()
	}

	releaseID := req.ReleaseID
	if releaseID == "" {
		logrus.Warn("Param [ReleaseID] is missing when get release info.")
		return nil, apierrors.ErrGetYAML.MissingParameter("releaseID")
	}

	logrus.Infof("getting dice.yml...releaseId: %s\n", releaseID)

	diceYAML, err := s.GetDiceYAMLData(orgID, releaseID)
	if err != nil {
		return nil, apierrors.ErrGetYAML.InvalidState("release not found")
	}
	return &pb.ReleaseGetDiceYmlResponse{Data: diceYAML}, nil
}

// GetDiceYAML Get dice.yml context
func (r *releaseGetDiceService) GetDiceYAMLData(orgID int64, releaseID string) (string, error) {
	release, err := r.db.GetRelease(releaseID)
	if err != nil {
		return "", err
	}
	if orgID != 0 && release.OrgID != orgID { // When calling internally，orgID is 0
		return "", errors.Errorf("release not found")
	}

	return release.Dice, nil
}
