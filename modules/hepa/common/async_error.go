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

package common

import (
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/hepa/bundle"
)

func AsyncRuntimeError(runtimeId, humanLog, primevalLog string) {
	err := bundle.Bundle.CreateErrorLog(&apistructs.ErrorLogCreateRequest{
		ErrorLog: apistructs.ErrorLog{
			ResourceType:   apistructs.RuntimeError,
			ResourceID:     runtimeId,
			HumanLog:       humanLog,
			PrimevalLog:    primevalLog,
			OccurrenceTime: strconv.FormatInt(time.Now().Unix(), 10),
		},
	})
	if err != nil {
		log.Errorf("async runtime error failed, err:%+v", errors.WithStack(err))
	}
}
