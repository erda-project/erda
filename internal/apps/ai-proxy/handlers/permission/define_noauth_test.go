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
	"testing"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	templatepb "github.com/erda-project/erda-proto-go/apps/aiproxy/template/pb"
	"github.com/erda-project/erda/internal/pkg/audit"
)

func TestIsNoNeedAuthMethodName(t *testing.T) {
	modelTplMethod := audit.GetMethodName(templatepb.TemplateServiceServer.ListModelTemplates)
	if !IsNoNeedAuthMethodName(modelTplMethod) {
		t.Fatalf("expected template list-model method to be no-auth")
	}

	spTplMethod := audit.GetMethodName(templatepb.TemplateServiceServer.ListServiceProviderTemplates)
	if !IsNoNeedAuthMethodName(spTplMethod) {
		t.Fatalf("expected template list-service-provider method to be no-auth")
	}

	clientGetMethod := audit.GetMethodName(clientpb.ClientServiceServer.Get)
	if IsNoNeedAuthMethodName(clientGetMethod) {
		t.Fatalf("expected client get method not to be no-auth")
	}
}
