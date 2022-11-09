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

package extension

import (
	"context"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/extension/pb"
	"github.com/erda-project/erda/pkg/i18n"
)

type Interface interface {
	GetExtension(name string, version string, yamlFormat bool) (*pb.ExtensionVersion, error)
	MenuExtWithLocale(extensions []*pb.Extension, locale *i18n.LocaleResource, all bool) (map[string][]pb.ExtensionMenu, error)
	CreateExtensionVersionByRequest(req *pb.ExtensionVersionCreateRequest) (*pb.ExtensionVersionCreateResponse, error)
	QueryExtensionList(all bool, typ string, labels string) ([]*pb.Extension, error)
	QueryExtensionVersions(ctx context.Context, req *pb.ExtensionVersionQueryRequest) (*pb.ExtensionVersionQueryResponse, error)
	DeleteExtensionVersion(name, version string) error
	ToProtoValue(i interface{}) (*structpb.Value, error)
}
