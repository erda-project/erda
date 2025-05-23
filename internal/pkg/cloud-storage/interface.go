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

package cloud_storage

import (
	"context"
	"io"

	"github.com/erda-project/erda/internal/pkg/cloud-storage/types"
)

type Interface interface {
	WhoIAm() string
	PutObject(ctx context.Context, localFilepath string, filename string) (*types.PutObjectResult, error)
	PutObjectWithReader(ctx context.Context, reader io.Reader, filename string) (*types.PutObjectResult, error)
	ListObjects(ctx context.Context) ([]types.FileObject, error)
	DeleteObject(ctx context.Context, key string) error
	DeleteObjects(ctx context.Context, keys []string) error
}
