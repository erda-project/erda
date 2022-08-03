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

package types

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type RuntimeHPARuleDTO struct {
	RuleID          string `json:"rule_id"`
	OrgID           uint64 `json:"org_id"`
	OrgName         string `json:"org_name"`
	ProjectID       uint64 `json:"project_id"`
	ProjectName     string `json:"org_name"`
	ApplicationID   uint64 `json:"application_id"`
	ApplicationName string `json:"application_name"`
	RuntimeID       uint64 `json:"runtime_id"`
	RuntimeName     string `json:"runtime_name"`
	ClusterName     string `json:"cluster_name"` // 部署目标所在 k8s 集群名称
	Workspace       string `json:"workspace"`
	UserId          string `json:"user_id"`   // 操作人ID
	UserName        string `json:"user_name"` // 操作人名称
	NickName        string `json:"nick_name"` // 操作人昵称
	ServiceName     string `json:"service_name"`
	Rules           string `json:"rules" gorm:"type:text"`
	// TODO: 是否必要待确认
	IsApplied string `json:"service_name"` // 表示规则是否已经应用，‘Y’ 表示已经应用，‘N’表示规则存在但未应用
}

type ErdaHPAObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

const (
	RuntimePARuleApplied  string = "Y"
	RuntimePARuleCanceled string = "N"

	// Labels for set sg Lables. if set, scaler is an autoscaler (HPA or VPA)
	ErdaPALabelKey           string = "erdaPALabel"
	ErdaHPALabelValue        string = "erdaHPA"
	ErdaHPALabelValueCreate  string = "erdaHPACreate"
	ErdaHPALabelValueApply   string = "erdaHPAApply"
	ErdaHPALabelValueCancel  string = "erdaHPACancel"
	ErdaHPALabelValueReApply string = "erdaHPAReApply"

	ErdaPARuleActionApply  string = "apply"
	ErdaPARuleActionCancel string = "cancel"
	// ErdaHPADefaultMaxReplicaCount define max replicas for hpa rule
	ErdaHPADefaultMaxReplicaCount int32 = 100

	ErdaHPATriggerCPU                     string = "cpu"
	ErdaHPATriggerCPUMetaType             string = "Utilization"
	ErdaHPATriggerMemory                  string = "memory"
	ErdaHPATriggerMemoryMetaType          string = "Utilization"
	ErdaHPATriggerCron                    string = "cron"
	ErdaHPATriggerCronMetaStart           string = "start"
	ErdaHPATriggerCronMetaEnd             string = "end"
	ErdaHPATriggerCronMetaDesiredReplicas string = "desiredReplicas"
	ErdaHPATriggerCronMetaTimeZone        string = "timezone"
	ErdaHPATriggerExternal                string = "external"

	ErdaPAObjectRuntimeIDLabel          string = "erdaRuntimeId"
	ErdaPAObjectRuleIDLabel             string = "erdaPARuleId"
	ErdaPAObjectOrgIDLabel              string = "erdaPAOrgId"
	ErdaPAObjectRuntimeServiceNameLabel string = "erdaRuntimeServiceName"

	// ErdaHPARecentlyEventsMaxToListForRuntimeDefault define max hpa events for recently to list for runtime
	ErdaHPARecentlyEventsMaxToListForRuntimeDefault int32 = 100
	// ErdaHPARecentlyEventsMaxToListForServiceDefault define max hpa events for recently to list for service
	ErdaHPARecentlyEventsMaxToListForServiceDefault int32 = 20

	ErdaHPAPrefix        string = "erdaHPA_"
	ErdaVPAPrefix        string = "erdaVPA_"
	ErdaHPAServiceSepStr string = "HPA_"
	ErdaVPAServiceSepStr string = "VPA_"

	ErdaVPARecommendationsHistory   int     = 8
	ErdaVPALabelValue               string  = "erdaVPA"
	ErdaVPALabelValueCreate         string  = "erdaVPACreate"
	ErdaVPALabelValueApply          string  = "erdaVPAApply"
	ErdaVPALabelValueCancel         string  = "erdaVPACancel"
	ErdaVPALabelValueReApply        string  = "erdaVPAReApply"
	ErdaVPAMaxResourceCPU           float64 = 32
	ErdaVPAMinResourceCPU           float64 = 0.005
	ErdaVPAMaxResourceMemory        int64   = 65536
	ErdaVPAMinResourceMemory        int64   = 32
	ErdaVPADefaultMaxCPU            float64 = 2
	ErdaVPADefaultMaxMemory         int64   = 4096
	ErdaVPADefaultResourceMaxFactor         = 20
	ErdaVPADefaultResourceMinFactor         = 4
	ErdaVPAUpdaterModeAuto          string  = "Auto"
	ErdaVPAUpdaterModeRecreate      string  = "Recreate"
	ErdaVPAUpdaterModeInitial       string  = "Initial"
	ErdaVPAUpdaterModeOff           string  = "Off"
	ErdaVPAPodEvictEventReason      string  = "EvictedByVPA"
)

// HPA event details
type EventDetail struct {
	LastTimestamp metav1.Time `json:"lastTimestamp,omitempty"`
	Type          string      `json:"type,omitempty"`
	Reason        string      `json:"reason,omitempty"`
	Message       string      `json:"message,omitempty"`
}
