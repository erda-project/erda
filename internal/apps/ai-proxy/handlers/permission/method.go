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
	promptpb "github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	serviceproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	templatepb "github.com/erda-project/erda-proto-go/apps/aiproxy/template/pb"
	usagepb "github.com/erda-project/erda-proto-go/apps/aiproxy/usage/token/pb"
)

var CheckClientPerm = CheckPermissions(
	&MethodPermission{Method: clientpb.ClientServiceServer.Create, OnlyAdmin: true},
	&MethodPermission{Method: clientpb.ClientServiceServer.Get, LoggedIn: true},
	&MethodPermission{Method: clientpb.ClientServiceServer.Update, OnlyAdmin: true},
	&MethodPermission{Method: clientpb.ClientServiceServer.Paging, LoggedIn: true},
	&MethodPermission{Method: clientpb.ClientServiceServer.Delete, OnlyAdmin: true},
)

var CheckServiceProviderPerm = CheckPermissions(
	&MethodPermission{Method: serviceproviderpb.ServiceProviderServiceServer.Create, AdminOrClient: true},
	&MethodPermission{Method: serviceproviderpb.ServiceProviderServiceServer.Get, LoggedIn: true},
	&MethodPermission{Method: serviceproviderpb.ServiceProviderServiceServer.Update, AdminOrClient: true},
	&MethodPermission{Method: serviceproviderpb.ServiceProviderServiceServer.Delete, AdminOrClient: true},
	&MethodPermission{Method: serviceproviderpb.ServiceProviderServiceServer.Paging, LoggedIn: true},
)

var CheckModelPerm = CheckPermissions(
	&MethodPermission{Method: modelpb.ModelServiceServer.Create, AdminOrClient: true},
	&MethodPermission{Method: modelpb.ModelServiceServer.Get, LoggedIn: true},
	&MethodPermission{Method: modelpb.ModelServiceServer.Update, AdminOrClient: true},
	&MethodPermission{Method: modelpb.ModelServiceServer.Delete, AdminOrClient: true},
	&MethodPermission{Method: modelpb.ModelServiceServer.Paging, LoggedIn: true},
	&MethodPermission{Method: modelpb.ModelServiceServer.UpdateModelAbilitiesInfo, AdminOrClient: true},
)

var CheckClientModelRelationPerm = CheckPermissions(
	&MethodPermission{Method: clientmodelrelationpb.ClientModelRelationServiceServer.ListClientModels, LoggedIn: true},
	&MethodPermission{Method: clientmodelrelationpb.ClientModelRelationServiceServer.Allocate, OnlyAdmin: true},
	&MethodPermission{Method: clientmodelrelationpb.ClientModelRelationServiceServer.UnAllocate, OnlyAdmin: true},
	&MethodPermission{Method: clientmodelrelationpb.ClientModelRelationServiceServer.Paging, OnlyAdmin: true},
)

var CheckPromptPerm = CheckPermissions(
	&MethodPermission{Method: promptpb.PromptServiceServer.Create, LoggedIn: true},
	&MethodPermission{Method: promptpb.PromptServiceServer.Get, LoggedIn: true},
	&MethodPermission{Method: promptpb.PromptServiceServer.Update, LoggedIn: true},
	&MethodPermission{Method: promptpb.PromptServiceServer.Delete, LoggedIn: true},
	&MethodPermission{Method: promptpb.PromptServiceServer.Paging, LoggedIn: true},
)

var CheckSessionPerm = CheckPermissions(
	&MethodPermission{Method: sessionpb.SessionServiceServer.Create, LoggedIn: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.Get, LoggedIn: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.Update, LoggedIn: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.Delete, LoggedIn: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.Paging, LoggedIn: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.Archive, LoggedIn: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.UnArchive, LoggedIn: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.Reset, LoggedIn: true},
	&MethodPermission{Method: sessionpb.SessionServiceServer.GetChatLogs, LoggedIn: true},
)

var CheckClientTokenPerm = CheckPermissions(
	&MethodPermission{Method: clienttokenpb.ClientTokenServiceServer.Create, AdminOrClient: true},
	&MethodPermission{Method: clienttokenpb.ClientTokenServiceServer.Get, LoggedIn: true},
	&MethodPermission{Method: clienttokenpb.ClientTokenServiceServer.Update, AdminOrClient: true},
	&MethodPermission{Method: clienttokenpb.ClientTokenServiceServer.Paging, AdminOrClient: true},
	&MethodPermission{Method: clienttokenpb.ClientTokenServiceServer.Delete, AdminOrClient: true},
)

var CheckRichClientPerm = CheckPermissions(
	&MethodPermission{Method: richclientpb.RichClientServiceServer.GetByAccessKeyId, AdminOrClient: true},
)

var CheckMCPPerm = CheckPermissions(
	&MethodPermission{Method: mcppb.MCPServerServiceServer.Get, AdminOrClient: true},
	&MethodPermission{Method: mcppb.MCPServerServiceServer.List, AdminOrClient: true},
	&MethodPermission{Method: mcppb.MCPServerServiceServer.Delete, OnlyAdmin: true},
	&MethodPermission{Method: mcppb.MCPServerServiceServer.Update, OnlyAdmin: true},
	&MethodPermission{Method: mcppb.MCPServerServiceServer.Register, AdminOrClient: true},
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
	&MethodPermission{Method: auditpb.AuditServiceServer.Paging, LoggedIn: true},
)

var CheckTokenUsagePerm = CheckPermissions(
	&MethodPermission{Method: usagepb.TokenUsageServiceServer.Paging, LoggedIn: true},
	&MethodPermission{Method: usagepb.TokenUsageServiceServer.Aggregate, LoggedIn: true},
)

var CheckTemplatePerm = CheckPermissions(
	&MethodPermission{Method: templatepb.TemplateServiceServer.ListServiceProviderTemplates, LoggedIn: true},
	&MethodPermission{Method: templatepb.TemplateServiceServer.ListModelTemplates, LoggedIn: true},
)
