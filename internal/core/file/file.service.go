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

package file

import (
	"io"

	"github.com/erda-project/erda-proto-go/core/file/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/file/db"
	"github.com/erda-project/erda/internal/core/file/types"
)

// ServiceInterface define service methods like a grpc service.
// For stream reason, we defined a non-methods file service in proto.
type ServiceInterface interface {
	UploadFile(req types.FileUploadRequest) (*pb.File, error)
	DownloadFile(w io.Writer, file db.File) (headers map[string]string, err error)
	DeleteFile(file db.File) error
}

type fileService struct {
	p *provider

	db  *db.Client
	bdl *bundle.Bundle
}
