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

package spec

import (
	"encoding/json"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

// PipelineExtra represents `pipeline_extras` table.
// `pipeline_extras` 与 `pipeline_bases` 一一对应
type PipelineExtra struct {
	PipelineID uint64 `json:"pipelineID,omitempty" xorm:"pk 'pipeline_id'"`

	// PipelineYml 流水线定义文件
	PipelineYml string `json:"pipelineYml"`

	// Extra 额外信息
	Extra PipelineExtraInfo `json:"extra" xorm:"json"`

	// NormalLabels 普通标签，仅展示，不可过滤
	NormalLabels map[string]string `json:"normalLabels" xorm:"json"`

	// Snapshot 运行时的快照
	Snapshot Snapshot `json:"snapshot" xorm:"json"`

	// CommitDetail 提交详情
	CommitDetail apistructs.CommitDetail `json:"commitDetail" xorm:"json"`

	// Progress 流水线整体执行进度，0-100
	// -1 表示未设置
	// progress 只存最终结果，若 >= 0，直接返回，无需再计算
	Progress int `json:"progress"`

	ExtraTimeCreated *time.Time `json:"timeCreated,omitempty" xorm:"created 'time_created'"`
	ExtraTimeUpdated *time.Time `json:"timeUpdated,omitempty" xorm:"updated 'time_updated'"`

	// 以下为冗余字段，因为使用 sql 迁移时，无法将 应用相关字段 迁移到 labels 中，所以要先做冗余
	// 新建的流水线，不会插入以下字段
	Commit  string `json:"commit"`
	OrgName string `json:"orgName"`

	Snippets []pipelineyml.SnippetPipelineYmlCache `json:"snippets" xorm:"snippets"`
}

func (*PipelineExtra) TableName() string {
	return "pipeline_extras"
}

type PipelineExtraInfo struct {
	Namespace         string                       `json:"namespace"`
	DiceWorkspace     apistructs.DiceWorkspace     `json:"diceWorkspace,omitempty"`
	PipelineYmlSource apistructs.PipelineYmlSource `json:"pipelineYmlSource,omitempty"`
	SubmitUser        *apistructs.PipelineUser     `json:"submitUser,omitempty"`
	RunUser           *apistructs.PipelineUser     `json:"runUser,omitempty"`
	CancelUser        *apistructs.PipelineUser     `json:"cancelUser,omitempty"`
	InternalClient    string                       `json:"internalClient,omitempty"`
	CronExpr          string                       `json:"cronExpr,omitempty"`
	CronTriggerTime   *time.Time                   `json:"cronTriggerTime,omitempty"` // 秒级精确，毫秒级误差请忽略，cron expr 精确度同样为秒级
	ShowMessage       *apistructs.ShowMessage      `json:"showMessage,omitempty"`
	Messages          []string                     `json:"errors,omitempty"` // TODO ShowMessage 和 Message
	// Deprecated
	ConfigManageNamespaceOfSecretsDefault string `json:"configManageNamespaceOfSecretsDefault,omitempty"`
	// Deprecated
	ConfigManageNamespaceOfSecrets string   `json:"configManageNamespaceOfSecrets,omitempty"`
	ConfigManageNamespaces         []string `json:"configManageNamespaces,omitempty"`

	CopyFromPipelineID *uint64            `json:"copyFromPipelineID,omitempty"` // 是否是从其他节点拷贝过来
	RerunFailedDetail  *RerunFailedDetail `json:"rerunFailedDetail,omitempty"`

	IsAutoRun      bool                     `json:"isAutoRun,omitempty"` // 创建后是否自动开始执行
	ShareVolumeID  string                   `json:"shareVolumeId,omitempty"`
	TaskWorkspaces []string                 `json:"taskWorkspaces,omitempty"` //工作目录,例如git
	StorageConfig  apistructs.StorageConfig `json:"storageConfig,omitempty"`  // 挂载设置

	CallbackURLs []string `json:"callbackURLs,omitempty"`

	Version string `json:"version,omitempty"` // 1.1, 1.0

	// 是否已经 完成 Reconciler GC
	CompleteReconcilerGC bool `json:"completeReconcilerGC"`

	// 是否已完成 Reconcile teardown
	CompleteReconcilerTeardown bool `json:"completeReconcilerTeardown"`

	// 用于保存自动转换前的 v1 pipelineYmlName（通过 V1 API 创建的流水线，通过该参数调用 gittar 获取内容）
	PipelineYmlNameV1 string `json:"pipelineYmlNameV1,omitempty"`

	// pipeline 运行时的输入参数
	RunPipelineParams []apistructs.PipelineRunParam `json:"runPipelineParams,omitempty"`

	// GC
	GC apistructs.PipelineGC `json:"gc,omitempty"`

	// OutputDefines
	DefinedOutputs []apistructs.PipelineOutput `json:"definedOutputs,omitempty"`

	SnippetChain []uint64 `json:"snippetChain,omitempty"`
}

type Snapshot struct {
	PipelineYml     string            `json:"pipeline_yml,omitempty"` // 对占位符进行渲染
	Secrets         map[string]string `json:"secrets,omitempty"`
	PlatformSecrets map[string]string `json:"platformSecrets,omitempty"`
	CmsDiceFiles    map[string]string `json:"cmsDiceFiles,omitempty"`
	Envs            map[string]string `json:"envs,omitempty"`

	AnalyzedCrossCluster *bool `json:"analyzedCrossCluster,omitempty"`

	RunPipelineParams apistructs.PipelineRunParamsWithValue `json:"runPipelineParams,omitempty"` // 流水线运行时参数

	// IdentityInfo 身份信息
	IdentityInfo apistructs.IdentityInfo `json:"identityInfo" xorm:"json"`

	// OutputValues output 定义和从 task 里采集上来的值
	OutputValues []apistructs.PipelineOutputWithValue `json:"outputValues,omitempty"`
}

// FromDB 兼容 Snapshot 老数据
func (s *Snapshot) FromDB(b []byte) error {

	// 先用 map[string]string 解析
	if err := json.Unmarshal(b, s); err == nil {
		return nil
	}

	// 用老数据结构 []apistructs.MetadataField 进行解析，并赋值回 s
	so := struct {
		PipelineYml string                     `json:"pipeline_yml,omitempty"`
		Secrets     []apistructs.MetadataField `json:"secrets,omitempty"`
	}{}
	if err := json.Unmarshal(b, &so); err == nil {
		s.PipelineYml = so.PipelineYml
		s.Secrets = make(map[string]string)
		for _, item := range so.Secrets {
			s.Secrets[item.Name] = item.Value
		}
	}

	// 反序列化失败，忽略错误，该字段可忽略

	return nil
}

func (s *Snapshot) ToDB() ([]byte, error) {
	return json.Marshal(s)
}

type RerunFailedDetail struct {
	RerunPipelineID uint64            `json:"rerunPipelineID,omitempty"`
	StageIndex      int               `json:"stageIndex,omitempty"`
	SuccessTasks    map[string]uint64 `json:"successTasks,omitempty"`
	FailedTasks     map[string]uint64 `json:"failedTasks,omitempty"`
	NotExecuteTasks map[string]uint64 `json:"notExecuteTasks,omitempty"`
}

func (extra *PipelineExtra) GetCommitID() string {
	if extra.CommitDetail.CommitID != "" {
		return extra.CommitDetail.CommitID
	}
	return extra.Commit
}

func (extra *PipelineExtra) GetOrgName() string {
	if extra.NormalLabels != nil {
		orgName, ok := extra.NormalLabels[apistructs.LabelOrgName]
		if ok {
			return orgName
		}
	}
	return extra.OrgName
}
