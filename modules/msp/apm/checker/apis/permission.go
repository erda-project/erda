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
