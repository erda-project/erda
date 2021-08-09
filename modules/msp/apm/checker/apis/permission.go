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

package apis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/modules/msp/apm/checker/storage/db"
	"github.com/erda-project/erda/pkg/common/errors"
	perm "github.com/erda-project/erda/pkg/common/permission"
)

func (s *checkerV1Service) getProjectFromMetricID() func(ctx context.Context, req interface{}) (string, error) {
	return getProjectFromMetricID(s.metricDB, s.projectDB)
}

func getProjectFromMetricID(metricDB *db.MetricDB, projectDB *db.ProjectDB) func(ctx context.Context, req interface{}) (string, error) {
	getter := perm.FieldValue("Id")
	return func(ctx context.Context, req interface{}) (string, error) {
		mid, err := getter(ctx, req)
		if err != nil {
			return "", err
		}
		metricID, err := strconv.ParseInt(mid, 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid metricID")
		}
		m, err := metricDB.GetByID(metricID)
		if err != nil {
			return "", errors.NewDatabaseError(err)
		}
		if m == nil {
			return "", fmt.Errorf("not found id for permission")
		}
		if m.Extra != "" {
			extra, err := strconv.ParseInt(m.Extra, 10, 64)
			if err != nil {
				return "", errors.NewInternalServerError(err)
			}
			if extra == m.ProjectID {
				proj, err := projectDB.GetByProjectID(m.ProjectID)
				if err != nil {
					return "", errors.NewDatabaseError(err)
				}
				if proj != nil {
					return strconv.FormatInt(proj.ProjectID, 10), nil
				}
			}
		}
		proj, err := projectDB.GetByID(m.ProjectID)
		if err != nil {
			return "", errors.NewDatabaseError(err)
		}
		if proj == nil {
			return "", fmt.Errorf("not found id for permission")
		}
		return strconv.FormatInt(proj.ProjectID, 10), nil
	}
}
