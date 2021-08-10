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

package http

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/eventbox/dispatcher/errors"
	"github.com/erda-project/erda/modules/eventbox/server/types"
)

func genResponse(dispatchErrs *errors.DispatchError) types.HTTPResponse {
	if len(dispatchErrs.BackendErrs) > 0 {
		logrus.Errorf("dispatcher backenderr: %v", dispatchErrs.BackendErrs)
		return types.HTTPResponse{Status: http.StatusBadRequest, Content: dispatchErrs.BackendErrs}
	}
	if dispatchErrs.FilterErr != nil {
		logrus.Errorf("dispatcher filterErr: %v", dispatchErrs.FilterErr)
		return types.HTTPResponse{Status: http.StatusBadRequest, Content: dispatchErrs.FilterErr.Error()}
	}
	return types.HTTPResponse{Status: http.StatusOK, Content: ""}
}
