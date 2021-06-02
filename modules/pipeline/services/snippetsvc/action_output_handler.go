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

package snippetsvc

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/encoding/jsonpath"
)

func handlerActionOutputsWithJq(action *apistructs.PipelineYmlAction, jq string) ([]string, error) {
	if action.Params == nil || len(jq) <= 0 {
		return nil, nil
	}

	paramsJson, err := json.Marshal(action.Params)
	if err != nil {
		return nil, err
	}

	filterValue, err := jsonpath.JQ(string(paramsJson), jq)
	if err != nil {
		return nil, err
	}

	filterJson, err := json.Marshal(filterValue)
	if err != nil {
		return nil, err
	}

	var outputs []string
	err = json.Unmarshal(filterJson, &outputs)
	if err != nil {
		logrus.Warnf("action jq filter outputs error: %v not []string", err)
		return nil, nil
	}

	return outputs, nil
}
