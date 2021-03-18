package apistructs

type BuildArtifact struct {
	ID           int64  `json:"id"`
	Sha256       string `json:"sha256"`
	IdentityText string `json:"identityText"`
	Type         string `json:"type"`
	Content      string `json:"content"`
	ClusterName  string `json:"clusterName"`
	PipelineID   uint64 `json:"pipelineID"`
}

type BuildArtifactType string

const (
	BuildArtifactOfNfsLink     BuildArtifactType = "NFS_LINK "
	BuildArtifactOfFileContent BuildArtifactType = "FILE_CONTENT "
)

// register

type BuildArtifactRegisterRequest struct {
	SHA          string `json:"sha"`
	IdentityText string `json:"identity_text"`
	Type         string `json:"type"`
	Content      string `json:"content"`
	ClusterName  string `json:"cluster_name"`
	PipelineID   uint64 `json:"pipelineID"`
}

type BuildArtifactRegisterResponse struct {
	Header
	Data *BuildArtifact `json:"data"`
}

// delete

type BuildArtifactDeleteByImagesRequest struct {
	Images []string `json:"images"`
}

// query

type BuildArtifactQueryResponse struct {
	Header
	Data *BuildArtifact `json:"data"`
}
