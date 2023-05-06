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

package common

// https://help.aliyun.com/document_detail/424813.htm
const (

	// 流量治理--Header 控制 header control
	AnnotationMSEHeaderControlRequestHeaderControlAdd     Annotation = "mse.ingress.kubernetes.io/request-header-control-add"     // Ingress  请求在转发给后端服务时，添加指定Header。若该Header存在，则其值拼接在原有值后面
	AnnotationMSEHeaderControlRequestHeaderControlUpdate  Annotation = "mse.ingress.kubernetes.io/request-header-control-update"  // Ingress  请求在转发给后端服务时，修改指定Header。若该Header存在，则其值覆盖原有值
	AnnotationMSEHeaderControlRequestHeaderControlRemove  Annotation = "mse.ingress.kubernetes.io/request-header-control-remove"  // Ingress  请求在转发给后端服务时，删除指定Header
	AnnotationMSEHeaderControlResponseHeaderControlAdd    Annotation = "mse.ingress.kubernetes.io/response-header-control-add"    // Ingress  请求收到后端服务响应后，在转发响应给客户端之前需要添加指定Header。若该Header存在，则其值拼接在原有值后面
	AnnotationMSEHeaderControlResponseHeaderControlUpdate Annotation = "mse.ingress.kubernetes.io/response-header-control-update" // Ingress  请求收到后端服务响应后，在转发响应给客户端之前需要修改指定Header。若该Header存在，则其值覆盖原有值
	AnnotationMSEHeaderControlResponseHeaderControlRemove Annotation = "mse.ingress.kubernetes.io/response-header-control-remove" // Ingress  请求收到后端服务响应后，在转发响应给客户端之前需要删除指定Header

	// 流量治理--超时 timeout
	AnnotationMSETimeOut Annotation = "mse.ingress.kubernetes.io/timeout" // Ingress  请求的超时时间，单位为秒。默认未配置超时时间 (说明:超时设置作用在应用层，非传输层TCP。)

	// 流量治理--单机限流 limit
	AnnotationMSELimitRouteLimitRPM             Annotation = "mse.ingress.kubernetes.io/route-limit-rpm"              // Ingress  该Ingress定义的路由在每个网关实例上每分钟最大请求次数。瞬时最大请求次数为该值乘以limit-burst-multiplier
	AnnotationMSELimitRouteLimitRPS             Annotation = "mse.ingress.kubernetes.io/route-limit-rps"              // Ingress  该Ingress定义的路由在每个网关实例上每秒最大请求次数。瞬时最大请求次数为该值乘以limit-burst-multiplier
	AnnotationMSELimitRouteLimitBurstMultiplier Annotation = "mse.ingress.kubernetes.io/route-limit-burst-multiplier" // Ingress  瞬时最大请求次数的因子，默认为5

	// 流量治理--服务预热 warmup
	AnnotationMSEServiceWarmUp Annotation = "mse.ingress.kubernetes.io/warmup" // Service   服务预热时间，单位为秒。默认不开启。

	// 流量治理--IP 访问控制 blacklist/whitelist
	AnnotationMSEBlackListSourceRange       Annotation = "mse.ingress.kubernetes.io/blacklist-source-range"        // Ingress  指定路由上的IP黑名单，支持IP地址或CIDR地址块，以英文逗号分隔
	AnnotationMSEDomainWhitelistSourceRange Annotation = "mse.ingress.kubernetes.io/domain-whitelist-source-range" // Ingress  指定域名上的IP白名单，域名优先级低于路由级别，支持IP地址或CIDR地址块，以英文逗号分隔
	AnnotationMSEDomainBlacklistSourceRange Annotation = "mse.ingress.kubernetes.io/domain-blacklist-source-range" // Ingress  指定域名上的IP黑名单，域名优先级低于路由级别，支持IP地址或CIDR地址块，以英文逗号分隔。

	// 安全防护--客户端与网关之间加密通信
	AnnotationMSETLSMinProtocolVersion Annotation = "mse.ingress.kubernetes.io/tls-min-protocol-version" // Domain  指定TLS的最小版本，默认值为TLSv1.0。合法值如下： TLSv1.0  TLSv1.1  TLSv1.2 TLSv1.3
	AnnotationMSETLSMaxProtocolVersion Annotation = "mse.ingress.kubernetes.io/tls-max-protocol-version" // Domain  指定TLS的最小版本，默认值为TLSv1.0。合法值如下： TLSv1.0  TLSv1.1  TLSv1.2 TLSv1.3
	AnnotationMSEAuthTLSSecret         Annotation = "mse.ingress.kubernetes.io/auth-tls-secret"          // Domain  网关使用的CA证书，用于验证MTLS握手期间，客户端提供的证书。该注解主要应用于网关需要验证客户端身份的场景。
)

type Annotation string

func (in Annotation) String() string {
	return string(in)
}
