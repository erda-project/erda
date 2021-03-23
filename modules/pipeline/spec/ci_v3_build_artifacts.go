package spec

import (
	"time"

	"github.com/erda-project/erda/apistructs"
)

type CIV3BuildArtifact struct {
	ID           int64                        `json:"id" xorm:"pk autoincr BIGINT(20)"`
	CreatedAt    time.Time                    `json:"createdAt" xorm:"created"`
	UpdatedAt    time.Time                    `json:"updatedAt" xorm:"updated"`
	Sha256       string                       `json:"sha256" xorm:"sha_256"` // 唯一标识
	IdentityText string                       `json:"identityText"`          // 便于记忆的字段，用来生成唯一标识
	Type         apistructs.BuildArtifactType `json:"type"`                  // 类型，存的是文件在 NFS 上的地址，或者直接是文件内容
	Content      string                       `json:"content"`               // 内容，根据 type 进行解析
	ClusterName  string                       `json:"clusterName"`           // 集群 name
	PipelineID   uint64                       `json:"pipelineID"`            // 关联的构建 ID
}

func (*CIV3BuildArtifact) TableName() string {
	return "ci_v3_build_artifacts"
}

func (artifact *CIV3BuildArtifact) Convert2DTO() *apistructs.BuildArtifact {
	if artifact == nil {
		return nil
	}
	return &apistructs.BuildArtifact{
		ID:           artifact.ID,
		Sha256:       artifact.Sha256,
		IdentityText: artifact.IdentityText,
		Type:         string(artifact.Type),
		Content:      artifact.Content,
		ClusterName:  artifact.ClusterName,
		PipelineID:   artifact.PipelineID,
	}
}
