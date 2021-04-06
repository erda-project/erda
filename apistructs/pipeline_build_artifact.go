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
