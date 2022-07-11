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

package horizontalpodscaler

import (
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
)

type DBService interface {
	CreateHPARule(req *dbclient.RuntimeHPA) error
	UpdateHPARule(req *dbclient.RuntimeHPA) error
	GetRuntimeHPARulesByServices(id spec.RuntimeUniqueId, services []string) ([]dbclient.RuntimeHPA, error)
	GetRuntimeHPARuleByRuleId(ruleId string) (dbclient.RuntimeHPA, error)
	GetRuntimeHPARulesByRuntimeId(runtimeID uint64) ([]dbclient.RuntimeHPA, error)
	DeleteRuntimeHPARulesByRuleId(ruleId string) error
	GetRuntime(id uint64) (*dbclient.Runtime, error)
	UpdateRuntime(runtime *dbclient.Runtime) error
	GetRuntimeByUniqueID(id spec.RuntimeUniqueId) (*dbclient.Runtime, error)
	FindRuntimesByIds(ids []uint64) ([]dbclient.Runtime, error)
	GetUnDeletableAttachMentsByRuntimeID(runtimeID uint64) (*[]dbclient.AddonAttachment, error)
	GetInstanceRouting(id string) (*dbclient.AddonInstanceRouting, error)
	UpdateAttachment(addonAttachment *dbclient.AddonAttachment) error
	GetPreDeployment(uniqueId spec.RuntimeUniqueId) (*dbclient.PreDeployment, error)
	UpdatePreDeployment(pre *dbclient.PreDeployment) error
	GetRuntimeHPAEventsByServices(runtimeId uint64, services []string) ([]dbclient.HPAEventInfo, error)
	DeleteRuntimeHPAEventsByRuleId(ruleId string) error
}
