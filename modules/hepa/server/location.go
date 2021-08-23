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

package server

const (
	OPENAPI_PREFIX string = "/api/gateway/openapi"

	SERVICE_API_PREFIX = "/service-api-prefix"

	CLOUDAPI_INFO = "/cloudapi-info"

	METRICS = "/metrics/*subpath"

	SERVICE_RUNTIME = "/service-runtime"

	FEATURES = "/gateway-features/:clusterName"

	TENANT_DOMAIN          = "/tenant-domain"
	RUNTIME_DOMAIN         = "/runtimes/:runtimeId/domains"
	RUNTIME_SERVICE_DOMAIN = "/runtimes/:runtimeId/services/:serviceName/domains"

	PACKAGES            = "/packages"
	PACKAGE             = "/packages/:packageId"
	PACKAGEAPIS         = "/packages/:packageId/apis"
	PACKAGEROOTAPI      = "/packages/:packageId/root-api"
	PACKAGEAPI          = "/packages/:packageId/apis/:apiId"
	PACKAGELOAD         = "/packages/:packageId/loadserver"
	PACKAGEACL          = "/packages/:packageId/consumers"
	PACKAGEAPIACL       = "/packages/:packageId/apis/:apiId/authz"
	PACKAGE_ALIYUN_BIND = "/packages/:packageId/aliyun-bind"

	CONSUMERS                  = "/consumers"
	CONSUMER                   = "/consumers/:consumerId"
	CONSUMERACL                = "/consumers/:consumerId/packages"
	CONSUMERAUTH               = "/consumers/:consumerId/credentials"
	CONSUMER_ALIYUN_AUTH       = "/consumers/:consumerId/aliyun-credentials"
	CONSUMER_ALIYUN_AUTH_ASYNC = "/consumers/:consumerId/aliyun-credentials-async"

	CLIENTS    = "/clients"
	CLIENT     = "/clients/:clientId"
	CLIENTACL  = "/clients/:clientId/packages/:packageId"
	CLIENTAUTH = "/clients/:clientId/credentials"

	CLIENTLIMIT = "/clients/:clientId/packages/:packageId/limits"

	PACKAGESNAME  = "/packages-name"
	CONSUMERSNAME = "/consumers-name"

	LIMITS = "/limits"
	LIMIT  = "/limits/:ruleId"
)

const (
	API_GATEWAY_PREFIX string = "/api/gateway"

	DICE_HEALTH = "/_api/health"

	DOMAINS = "/domains"

	COMPONENT_INGRESS = "/component-ingress"

	TENANT_GROUP = "/tenant-group"

	PUB_AUTHN   = "/publications/:apiPublishId/authn"
	PUB_SWAGGER = "/publications/:apiPublishId/swagger"
	PUB_SUB     = "/publications/:apiPublishId/subscribe"

	REG       = "/registrations"
	REG_PUB   = "/registrations/:apiRegisterId/publish"
	REG_STS   = "/registrations/:apiRegisterId/status"
	API_CHECK = "/check-compatibility"

	RUNTIME_SERVICE        = "/runtime-services"
	RUNTIME_SERVICE_DELETE = "/runtime-services/:runtimeId"

	//租户管理
	TENANTS = "/tenants"
	TENANT  = "/tenant/:tenantId"

	//healthCheck
	HEALTH_CHECK = "/health/check"

	// api网关相关
	GATEWAY_UI_TYPE     = "/ui-type"
	GATEWAY_APP_LIST    = "/register-apps"
	GATEWAY_BIND_DOMAIN = "/domain"

	GATEWAY_GROUPS       = "/group"
	GATEWAY_GROUP_CREATE = "/group"

	GATEWAY_CONSUMER_CREATE       = "/consumer"
	GATEWAY_PROJECT_CONSUMER_INFO = "/consumer"
	GATEWAY_CONSUMER_API_EDIT     = "/consumer"
	GATEWAY_CONSUMER_DELETE       = "/consumer/:consumerId"
	GATEWAY_CONSUMER_INFO         = "/consumer/:consumerId"
	GATEWAY_CONSUMER_UPDATE       = "/consumer/:consumerId"

	GATEWAY_CONSUMER_API_INFO = "/consumer-api"
	GATEWAY_CONSUMER_LIST     = "/consumer-list"

	GATEWAY_GROUP_DELETE = "/group/:groupId"
	GATEWAY_GROUP_UPDATE = "/group/:groupId"

	API_GATEWAY_API         = "/api"
	API_GATEWAY_API_ID      = "/api/:apiId"
	API_GATEWAY_CATEGORY    = "/policies/:category"
	API_GATEWAY_CATEGORY_ID = "/policies/:category/:policyId"

	UPSTREAM_REGISTER       = "/register"
	UPSTREAM_REGISTER_ASYNC = "/register_async"

	UPSTREAM_TARGET_ONLINE  = "/target/online"
	UPSTREAM_TARGET_OFFLINE = "/target/offline"

	//mock
	API_MOCK_REGISTER = "/api/mock/register"
	API_MOCK_CALL     = "/api/mock/call"
	//业务网关注册
	API_TRANSFORM_REGISTER = "/api/rpc/register"
	API_GET_TRANS_CONFIG   = "/api/rpc/conf"
	API_DELETE_SERVICE     = "/api/rpc/delete"

	REQUEST_SERVICE  = "/rpc/:targetKey"
	REGISTER_SERVICE = "/rpc/register"
	DELETE_SERVICE   = "/rpc/delete"
)
