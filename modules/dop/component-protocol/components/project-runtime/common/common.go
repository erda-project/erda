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

package common

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
)

const (
	ScenarioKey = "project-runtime"

	ProjectRuntimeFilter = "filter"

	DeleteOp = "delete"
	//DeployReleaseOp = "deploy-release"
	ReStartOp = "restart"

	FilterApp             = "app"
	FilterRuntimeStatus   = "runtimeStatus"
	FilterDeployStatus    = "deploymentStatus"
	FilterDeployOrderName = "deploymentOrderName"
	FilterDeployTime      = "deployTime"

	FilterInputCondition = "title"

	FrontedStatusSuccess    = "success"
	FrontedStatusDefault    = "default"
	FrontedStatusError      = "error"
	FrontedStatusWarning    = "warning"
	FrontedStatusProcessing = "processing"
)

var (
	PtrRequiredErr     = errors.New("b must be a pointer")
	NothingToBeDoneErr = errors.New("nothing to be done")
)

type Operation struct {
	JumpOut bool                   `json:"jumpOut"`
	Target  string                 `json:"target"`
	Query   map[string]interface{} `json:"query"`
	Params  map[string]interface{} `json:"params"`
}

// Transfer transfer a to b with json, kind of b must be pointer
func Transfer(a, b interface{}) error {
	if reflect.ValueOf(b).Kind() != reflect.Ptr {
		return PtrRequiredErr
	}
	if a == nil {
		return NothingToBeDoneErr
	}
	aBytes, err := json.Marshal(a)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(aBytes, b); err != nil {
		return err
	}
	return nil
}
