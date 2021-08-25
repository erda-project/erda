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
