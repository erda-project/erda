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

package permission

import (
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
)

type Interface interface {
	CheckInternalClient(identityInfo *commonpb.IdentityInfo) error
	CheckApp(identityInfo *commonpb.IdentityInfo, appID uint64, action string) error
	CheckBranch(identityInfo *commonpb.IdentityInfo, appIDStr, branch, action string) error
}
