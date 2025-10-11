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
	auditpb "github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	richclientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/rich_client/pb"
	clientmodelrelationpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	i18npb "github.com/erda-project/erda-proto-go/apps/aiproxy/i18n/pb"
	mcppb "github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	promptpb "github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	usagepb "github.com/erda-project/erda-proto-go/apps/aiproxy/usage/token/pb"
)

var CheckClientPerm = CheckPermissions(
	&MethodPermission{Method: clientpb.ClientServiceServer.Create, OnlyAdmin: true},
	&MethodPermission{Method: clientpb.ClientServiceServer.Get, AdminOrAk: true},
	&MethodPermission{Method: clientpb.ClientServiceServer.Update, OnlyAdmin: true},
	&MethodPermission{Method: clientpb.ClientServiceServer.Paging, OnlyAdmin: true},
	&MethodPermission{Method: clientpb.ClientServiceServer.Delete, OnlyAdmin: true},
)

var CheckModelProviderPerm = CheckPermissions(
	&MethodPermission{Method: modelproviderpb.ModelProviderServiceServer.Create, OnlyAdmin: true},
	&MethodPermission{Method: modelproviderpb.ModelProviderServiceServer.Get, OnlyAdmin: true},
	&MethodPermission{Method: modelproviderpb.ModelProviderServiceServer.Update, OnlyAdmin: true},
	&MethodPermission{Method: modelproviderpb.ModelProviderServiceServer.Delete, OnlyAdmin: true},
	&MethodPermission{Method: modelproviderpb.ModelProviderServiceServer.Paging, OnlyAdmin: true},
)

var CheckModelPerm = CheckPermissions(
	&MethodPermission{Method: modelpb.ModelServiceServer.Create, OnlyAdmin: true},
	&MethodPermission{Method: modelpb.ModelServiceServer.Get, AdminOrAk: true},
	&MethodPermission{Method: modelpb.ModelServiceServer.Update, OnlyAdmin: true},
	&MethodPermission{Method: modelpb.ModelServiceServer.Delete, OnlyAdmin: true},
	&MethodPermission{Method: modelpb.ModelServiceServer.Paging, OnlyAdmin: true},
	&MethodPermission{Method: modelpb.ModelServiceServer.UpdateModelAbilitiesInfo, OnlyAdmin: true},
)

var CheckClientModelRelationPerm = CheckPermissions(
	&MethodPermission{Method: clientmodelrelationpb.ClientModelRelationServiceServer.ListClientModels, AdminOrAk: true},
	&MethodPermission{Method: clientmodelrelationpb.ClientModelRelationServiceServer.Allocate, OnlyAdmin: true},
	&MethodPermission{Method: clientmodelrelationpb.ClientModelRelationServiceServer.UnAllocate, OnlyAdmin: true},
	&MethodPermission{Method: clientmodelrelationpb.ClientModelRelationServiceServer.Paging, OnlyAdmin: true},
)

var CheckPromptPerm = CheckPermissions(
	&MethodPermission{Method: promptpb.PromptServiceServer.Create, AdminOrAk: true},
	&MethodPermission{Method: promptpb.PromptServiceServer.Get, AdminOrAk: true},
	&MethodPermission{Method: promptpb.PromptServiceServer.Update, AdminOrAk: true, CheckButNotSetClientId: false},
	&MethodPermission{Method: promptpb.PromptServiceServer.Delete, AdminOrAk: true},
	&MethodPermission{Method: promptpb.PromptServiceServer.Paging, AdminOrAk: true},
)

var CheckSessionPerm = CheckPermissions(
	&MethodPermission{Method: sessionpb.SessionServiceServer.Create, AdminOrAk: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.Get, AdminOrAk: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.Update, AdminOrAk: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.Delete, AdminOrAk: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.Paging, AdminOrAk: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.Archive, AdminOrAk: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.UnArchive, AdminOrAk: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.Reset, AdminOrAk: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.GetChatLogs, AdminOrAk: true},
)

var CheckClientTokenPerm = CheckPermissions(
	&MethodPermission{Method: clienttokenpb.ClientTokenServiceServer.Create, AdminOrAk: true},
	&MethodPermission{Method: clienttokenpb.ClientTokenServiceServer.Get, AdminOrAk: true},
	&MethodPermission{Method: clienttokenpb.ClientTokenServiceServer.Update, AdminOrAk: true},
	&MethodPermission{Method: clienttokenpb.ClientTokenServiceServer.Paging, AdminOrAk: true},
	&MethodPermission{Method: clienttokenpb.ClientTokenServiceServer.Delete, AdminOrAk: true},
)

var CheckRichClientPerm = CheckPermissions(
	&MethodPermission{Method: richclientpb.RichClientServiceServer.GetByAccessKeyId, AdminOrAk: true},
)

var CheckMCPPerm = CheckPermissions(
	&MethodPermission{Method: mcppb.MCPServerServiceServer.Get, AdminOrAk: true},
	&MethodPermission{Method: mcppb.MCPServerServiceServer.List, AdminOrAk: true},
	&MethodPermission{Method: mcppb.MCPServerServiceServer.Delete, OnlyAdmin: true},
	&MethodPermission{Method: mcppb.MCPServerServiceServer.Update, OnlyAdmin: true},
	&MethodPermission{Method: mcppb.MCPServerServiceServer.Register, AdminOrAk: true},
	&MethodPermission{Method: mcppb.MCPServerServiceServer.Publish, OnlyAdmin: true},
)

var CheckI18nPerm = CheckPermissions(
	&MethodPermission{Method: i18npb.I18NServiceServer.Create, OnlyAdmin: true},
	&MethodPermission{Method: i18npb.I18NServiceServer.Get, OnlyAdmin: true},
	&MethodPermission{Method: i18npb.I18NServiceServer.Update, OnlyAdmin: true},
	&MethodPermission{Method: i18npb.I18NServiceServer.Delete, OnlyAdmin: true},
	&MethodPermission{Method: i18npb.I18NServiceServer.Paging, OnlyAdmin: true},
	&MethodPermission{Method: i18npb.I18NServiceServer.BatchCreate, OnlyAdmin: true},
	&MethodPermission{Method: i18npb.I18NServiceServer.GetByConfig, OnlyAdmin: true},
)

var CheckAuditPerm = CheckPermissions(
	&MethodPermission{Method: auditpb.AuditServiceServer.Paging, AdminOrAk: true},
)

var CheckTokenUsagePerm = CheckPermissions(
	&MethodPermission{Method: usagepb.TokenUsageServiceServer.Paging, AdminOrAk: true},
	&MethodPermission{Method: usagepb.TokenUsageServiceServer.Aggregate, AdminOrAk: true},
)
