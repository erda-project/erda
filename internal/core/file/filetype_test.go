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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/core/file/pb"
)

func TestGetFileTypeByExt(t *testing.T) {
	assert.Equal(t, pb.FileType_picture, GetFileTypeByExt(".png"))
	assert.Equal(t, pb.FileType_picture, GetFileTypeByExt(".jpg"))
	assert.Equal(t, pb.FileType_picture, GetFileTypeByExt(".jpeg"))
	assert.Equal(t, pb.FileType_picture, GetFileTypeByExt(".bmp"))
	assert.Equal(t, pb.FileType_picture, GetFileTypeByExt(".gif"))
	assert.Equal(t, pb.FileType_other, GetFileTypeByExt(".html"))
}
