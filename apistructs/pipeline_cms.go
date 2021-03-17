package apistructs

import "time"

const PipelineAppConfigNameSpacePreFix = "pipeline-secrets-app"

type PipelineCmsNs struct {
	PipelineSource PipelineSource `json:"pipelineSource"`
	NS             string         `json:"ns"`
	TimeCreated    *time.Time     `json:"timeCreated"`
	TimeUpdated    *time.Time     `json:"timeUpdated"`
}

// PipelineCmsCreateNsRequest 创建 pipeline 配置管理 namespace 请求体.
type PipelineCmsCreateNsRequest struct {
	PipelineSource PipelineSource `json:"pipelineSource"`
	NS             string         `json:"ns"`
}

// PipelineCmsCreateNsResponse 创建 pipeline 配置管理 namespace 返回体.
type PipelineCmsCreateNsResponse struct {
	Header
}

// PipelineCmsUpdateConfigsRequestV1 更新 pipeline 配置管理 配置 请求体. 不加密存储.
type PipelineCmsUpdateConfigsRequestV1 struct {
	KVs map[string]string `json:"kvs"`
}

// PipelineCmsUpdateConfigsRequest 更新 pipeline 配置管理 配置 请求体. 可选是否加密存储.
type PipelineCmsUpdateConfigsRequest struct {
	PipelineSource PipelineSource                    `json:"pipelineSource"`
	KVs            map[string]PipelineCmsConfigValue `json:"kvs"`
}

// PipelineCmsUpdateConfigsResponse 更新 pipeline 配置管理 配置 返回体.
type PipelineCmsUpdateConfigsResponse struct {
	Header
}

// PipelineCmsListNsRequest 列表查询 pipeline 配置管理 namespace 请求体.
type PipelineCmsListNsRequest struct {
	PipelineSource PipelineSource `json:"pipelineSource"`
	NsPrefix       string         `json:"nsPrefix"`
}
type PipelineCmsListNsResponse struct {
	Header
	Data []PipelineCmsNs `json:"data,omitempty"`
}

// PipelineCmsDeleteConfigsRequest 删除 pipeline 配置管理 配置 请求体.
type PipelineCmsDeleteConfigsRequest struct {
	PipelineSource PipelineSource
	DeleteNS       bool     `json:"deleteNS"`
	DeleteForce    bool     `json:"deleteForce"`
	DeleteKeys     []string `json:"deleteKeys"`
}

// PipelineCmsDeleteConfigsResponse 删除 pipeline 配置管理 配置 返回体.
type PipelineCmsDeleteConfigsResponse struct {
	Header
}

// PipelineCmsGetConfigsRequest 查询 pipeline 配置管理 配置 请求体.
type PipelineCmsGetConfigsRequest struct {
	PipelineSource PipelineSource         `json:"pipelineSource"`
	Keys           []PipelineCmsConfigKey `json:"keys"`          // 只获取指定的 key
	GlobalDecrypt  bool                   `json:"globalDecrypt"` // 全局指定是否解密，优先级低于 keys 中每个 key 的解密配置
}

type PipelineCmsConfigKey struct {
	Key                string `json:"key"`
	Decrypt            bool   `json:"decrypt"`
	ShowEncryptedValue bool   `json:"showEncryptedValue"` // 是否展示加密后的值
}

// PipelineCmsGetConfigsResponse 查询 pipeline 配置管理 配置 请求体.
type PipelineCmsGetConfigsResponse struct {
	Header
	Data []PipelineCmsConfig `json:"data"`
}

// PipelineCmsConfig 配置项
type PipelineCmsConfig struct {
	Key string `json:"key"`
	PipelineCmsConfigValue
}

// PipelineCmsConfigValue 配置项的值
type PipelineCmsConfigValue struct {
	// Value
	// 更新时，Value 为 realValue
	// 获取时，若 Decrypt=true，Value=decrypt(dbValue)；若 Decrypt=false，Value=dbValue
	Value string `json:"value"`
	// EncryptInDB 在数据库中是否加密存储
	EncryptInDB bool `json:"encryptInDB"`
	// Type
	// if not specified, default type is `kv`;
	// if type is `dice-file`, value is uuid of `dice-file`.
	Type PipelineCmsConfigType `json:"type"`
	// Operations 配置项操作，若为 nil，则使用默认配置: canDownload=false, canEdit=true, canDelete=true
	Operations *PipelineCmsConfigOperations `json:"operations"`
	// Comment
	Comment string `json:"comment"`
	// From 配置项来源，可为空。例如：证书管理同步
	From string `json:"from"`

	// 创建或更新时以下字段无需填写
	TimeCreated *time.Time `json:"timeCreated"`
	TimeUpdated *time.Time `json:"timeUpdated"`
}

var (
	PipelineCmsConfigDefaultOperationsForKV        = PipelineCmsConfigOperations{CanDownload: false, CanEdit: true, CanDelete: true}
	PipelineCmsConfigDefaultOperationsForDiceFiles = PipelineCmsConfigOperations{CanDownload: true, CanEdit: true, CanDelete: true}
)

// PipelineCmsConfigType 配置项类型
type PipelineCmsConfigType string

var (
	// kv 无需特殊处理
	PipelineCmsConfigTypeKV PipelineCmsConfigType = "kv"

	// dice-file 类型，config value 为 diceFileUUID
	// pipeline 上下文会特殊处理
	PipelineCmsConfigTypeDiceFile PipelineCmsConfigType = "dice-file"
)

func (t PipelineCmsConfigType) Valid() bool {
	return t == PipelineCmsConfigTypeKV || t == PipelineCmsConfigTypeDiceFile
}

// PipelineCmsConfigOperations 配置项操作
type PipelineCmsConfigOperations struct {
	CanDownload bool `json:"canDownload"`
	CanEdit     bool `json:"canEdit"`
	CanDelete   bool `json:"canDelete"` // CanDelete 仅在删除单个配置项时生效。若删除 ns，则所有配置项均会被删除
}
