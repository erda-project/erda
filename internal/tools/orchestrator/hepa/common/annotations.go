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

// https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/?spm=a2c4g.11186623.0.0.21d017325l6vNp#annotations
const (
	AnnotationAppRoot                              Annotation = "nginx.ingress.kubernetes.io/app-root"                                 // Ingress   修改应用根路径，对于访问/的请求将会被重定向为设置的新路径    值类型: string
	AnnotationCookieAffinity                       Annotation = "nginx.ingress.kubernetes.io/affinity"                                 // Service   亲和性种类，目前只支持Cookie，默认为cookie    值类型: cookie
	AnnotationCookieAffinityMode                   Annotation = "nginx.ingress.kubernetes.io/affinity-mode"                            // Service   亲和性模式，MSE Ingress目前只支持Balanced模式，默认为balanced模式。  值类型: "balanced" or "persistent"
	AnnotationAffinityCanaryBehavior               Annotation = "nginx.ingress.kubernetes.io/affinity-canary-behavior"                 // Ingress   值类型:   "sticky" or "legacy"
	AnnotationAuthRealm                            Annotation = "nginx.ingress.kubernetes.io/auth-realm"                               // Ingress   保护域。相同的保护域共享用户名和密码。   值类型: string
	AnnotationAuthSecret                           Annotation = "nginx.ingress.kubernetes.io/auth-secret"                              // Ingress   Secret名字，格式支持<namespace>/<name>，包含被授予能够访问该Ingress上定义的路由的访问权限的用户名和密码。   值类型: string
	AnnotationAuthSecretType                       Annotation = "nginx.ingress.kubernetes.io/auth-secret-type"                         // Ingress   Secret内容格式。 auth-file：Data的Key为auth，Value为用户名和密码，多账号回车分隔。 auth-map：Data的Key为用户名，Value为密码。   值类型: string
	AnnotationAuthType                             Annotation = "nginx.ingress.kubernetes.io/auth-type"                                // Ingress   认证类型. mse部分兼容，暂只支持Basic。     值类型: basic 或 digest
	AnnotationAuthTLSSecret                        Annotation = "nginx.ingress.kubernetes.io/auth-tls-secret"                          // Ingress   值类型: string
	AnnotationAuthTLSVerifyDepth                   Annotation = "nginx.ingress.kubernetes.io/auth-tls-verify-depth"                    // Ingress   值类型: number
	AnnotationAuthTLSVerifyClient                  Annotation = "nginx.ingress.kubernetes.io/auth-tls-verify-client"                   // Ingress   值类型: string
	AnnotationAuthTLSErrorPage                     Annotation = "nginx.ingress.kubernetes.io/auth-tls-error-page"                      // Ingress   值类型: string
	AnnotationAuthTLSPassCertificateToUpstream     Annotation = "nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream"    // Ingress   值类型: "true" or "false"
	AnnotationAuthTLSMatchCN                       Annotation = "nginx.ingress.kubernetes.io/auth-tls-match-cn"                        // Ingress   值类型: string
	AnnotationAuthURL                              Annotation = "nginx.ingress.kubernetes.io/auth-url"                                 // Ingress   值类型: string
	AnnotationAuthCacheKey                         Annotation = "nginx.ingress.kubernetes.io/auth-cache-key"                           // Ingress   值类型: string
	AnnotationAuthCacheDuration                    Annotation = "nginx.ingress.kubernetes.io/auth-cache-duration"                      // Ingress   值类型: string
	AnnotationAuthKeepAlive                        Annotation = "nginx.ingress.kubernetes.io/auth-keepalive"                           // Ingress   值类型: number
	AnnotationAuthKeepaliveRequests                Annotation = "nginx.ingress.kubernetes.io/auth-keepalive-requests"                  // Ingress   值类型: number
	AnnotationAuthKeepaliveTimeout                 Annotation = "nginx.ingress.kubernetes.io/auth-keepalive-timeout"                   // Ingress   值类型: number
	AnnotationAuthProxySetHeaders                  Annotation = "nginx.ingress.kubernetes.io/auth-proxy-set-headers"                   // Ingress   值类型: string
	AnnotationAuthSnippet                          Annotation = "nginx.ingress.kubernetes.io/auth-snippet"                             // Ingress   值类型: string
	AnnotationEnableGlobalAuth                     Annotation = "nginx.ingress.kubernetes.io/enable-global-auth"                       // Ingress   值类型: "true" or "false"
	AnnotationBackendProtocol                      Annotation = "nginx.ingress.kubernetes.io/backend-protocol"                         // Service   指定后端服务使用的协议，默认为HTTP，MSE 下支持： HTTP / HTTP2 / HTTPS / gRPC / gRPCS. MSE 不支持 AJP 和 FCGI。   值类型: string
	AnnotationCanary                               Annotation = "nginx.ingress.kubernetes.io/canary"                                   // Ingress   开启或关闭灰度发布    值类型: "true" or "false"
	AnnotationCanaryByHeader                       Annotation = "nginx.ingress.kubernetes.io/canary-by-header"                         // Ingress   基于Request Header Key流量切分     值类型: string
	AnnotationCanaryByHeaderValue                  Annotation = "nginx.ingress.kubernetes.io/canary-by-header-value"                   // Ingress   基于Request Header Value流量切分，Value为精确匹配    值类型: string
	AnnotationCanaryByHeaderPattern                Annotation = "nginx.ingress.kubernetes.io/canary-by-header-pattern"                 // Ingress   基于Request Header Value流量切分，Value为正则匹配    值类型: string
	AnnotationCanaryByCookie                       Annotation = "nginx.ingress.kubernetes.io/canary-by-cookie"                         // Ingress   基于Request Cookie Key流量切分    值类型: string
	AnnotationCanaryWeight                         Annotation = "nginx.ingress.kubernetes.io/canary-weight"                            // Ingress   基于权重和流量切分     值类型: number
	AnnotationCanaryWeightTotal                    Annotation = "nginx.ingress.kubernetes.io/canary-weight-total"                      // Ingress   权重总和     值类型: number
	AnnotationClientBodyBufferSize                 Annotation = "nginx.ingress.kubernetes.io/client-body-buffer-size"                  // Ingress   值类型: string
	AnnotationConfigurationSnippet                 Annotation = "nginx.ingress.kubernetes.io/configuration-snippet"                    // Ingress   值类型: string
	AnnotationFallbackCustomHttpErrors             Annotation = "nginx.ingress.kubernetes.io/custom-http-errors"                       // Ingress   和default-backend一起工作。当后端服务返回指定HTTP响应码，原始请求会被再次转发至容灾服务      值类型: []int
	AnnotationFallbackDefaultBackend               Annotation = "nginx.ingress.kubernetes.io/default-backend"                          // Ingress   容灾服务。当Ingress定义的服务没有可用节点时，请求会自动转发该容灾服务      值类型: string
	AnnotationEnableCORS                           Annotation = "nginx.ingress.kubernetes.io/enable-cors"                              // Ingress   开启或关闭跨域      值类型: "true" or "false"
	AnnotationCORSAllowOrigin                      Annotation = "nginx.ingress.kubernetes.io/cors-allow-origin"                        // Ingress	允许的第三方站点      值类型: string
	AnnotationCORSAllowMethods                     Annotation = "nginx.ingress.kubernetes.io/cors-allow-methods"                       // Ingress	允许的请求方法，如GET、POST、PUT等      值类型: string
	AnnotationCORSAllowHeaders                     Annotation = "nginx.ingress.kubernetes.io/cors-allow-headers"                       // Ingress	允许的请求Header          值类型: string
	AnnotationCORSExposeHeaders                    Annotation = "nginx.ingress.kubernetes.io/cors-expose-headers"                      // Ingress	允许的暴露给浏览器的响应Header       值类型: string
	AnnotationCORSAllowCredentials                 Annotation = "nginx.ingress.kubernetes.io/cors-allow-credentials"                   // Ingress	是否允许携带凭证信息       值类型: "true" or "false"
	AnnotationCORSMaxAge                           Annotation = "nginx.ingress.kubernetes.io/cors-max-age"                             // Ingress   预检结果的最大缓存时间。    值类型: number
	AnnotationForceSSLRedirect                     Annotation = "nginx.ingress.kubernetes.io/force-ssl-redirect"                       // Ingress   HTTP重定向为HTTPS    值类型: "true" or "false"
	AnnotationFromToWWWRedirect                    Annotation = "nginx.ingress.kubernetes.io/from-to-www-redirect"                     // Ingress   值类型: "true" or "false"
	AnnotationHttp2PushPreload                     Annotation = "nginx.ingress.kubernetes.io/http2-push-preload"                       // Ingress   值类型: "true" or "false"
	AnnotationLimitConnections                     Annotation = "nginx.ingress.kubernetes.io/limit-connections"                        // Ingress   值类型: number
	AnnotationLimitRPS                             Annotation = "nginx.ingress.kubernetes.io/limit-rps"                                // Ingress   值类型: number
	AnnotationGlobalRateLimit                      Annotation = "nginx.ingress.kubernetes.io/global-rate-limit"                        // Ingress   值类型: number
	AnnotationGlobalRateLimitWindow                Annotation = "nginx.ingress.kubernetes.io/global-rate-limit-window"                 // Ingress   值类型: duration
	AnnotationGlobalRateLimitKey                   Annotation = "nginx.ingress.kubernetes.io/global-rate-limit-key"                    // Ingress   值类型: string
	AnnotationGlobalRateLimitIgnoredCIDRs          Annotation = "nginx.ingress.kubernetes.io/global-rate-limit-ignored-cidrs"          // Ingress   值类型: string
	AnnotationPermanentRedirect                    Annotation = "nginx.ingress.kubernetes.io/permanent-redirect"                       // Ingress   永久重定向     值类型: string
	AnnotationPermanentRedirectCode                Annotation = "nginx.ingress.kubernetes.io/permanent-redirect-code"                  // Ingress   永久重定向状态码     值类型: number
	AnnotationTemporalRedirect                     Annotation = "nginx.ingress.kubernetes.io/temporal-redirect"                        // Ingress   临时重定向       值类型: string
	AnnotationPreserveTrailingSlash                Annotation = "nginx.ingress.kubernetes.io/preserve-trailing-slash"                  // Ingress   值类型: "true" or "false"
	AnnotationProxyBodySize                        Annotation = "nginx.ingress.kubernetes.io/proxy-body-size"                          // Ingress   值类型: string
	AnnotationProxyCookieDomain                    Annotation = "nginx.ingress.kubernetes.io/proxy-cookie-domain"                      // Ingress   值类型: string
	AnnotationProxyCookiePath                      Annotation = "nginx.ingress.kubernetes.io/proxy-cookie-path"                        // Ingress   值类型: string
	AnnotationProxyConnectTimeout                  Annotation = "nginx.ingress.kubernetes.io/proxy-connect-timeout"                    // Ingress   值类型: string
	AnnotationProxySendTimeout                     Annotation = "nginx.ingress.kubernetes.io/proxy-send-timeout"                       // Ingress   值类型: string
	AnnotationProxyReadTimeout                     Annotation = "nginx.ingress.kubernetes.io/proxy-read-timeout"                       // Ingress   值类型: string
	AnnotationProxyNextUpstream                    Annotation = "nginx.ingress.kubernetes.io/proxy-next-upstream"                      // Ingress   请求重试条件    值类型: string
	AnnotationProxyNextUpstreamTimeOut             Annotation = "nginx.ingress.kubernetes.io/proxy-next-upstream-timeout"              // Ingress   请求重试的超时时间，单位为秒。默认未配置超时时间   值类型: number
	AnnotationProxyNextUpstreamRetries             Annotation = "nginx.ingress.kubernetes.io/proxy-next-upstream-tries"                // Ingress   请求的最大重试次数。默认3次    值类型: number
	AnnotationProxyRequestBuffering                Annotation = "nginx.ingress.kubernetes.io/proxy-request-buffering"                  // Ingress   值类型: string
	AnnotationProxyRedirectFrom                    Annotation = "nginx.ingress.kubernetes.io/proxy-redirect-from"                      // Ingress   值类型: string
	AnnotationProxyRedirectTo                      Annotation = "nginx.ingress.kubernetes.io/proxy-redirect-to"                        // Ingress   值类型: string
	AnnotationProxyHttpVersion                     Annotation = "nginx.ingress.kubernetes.io/proxy-http-version"                       // Ingress   值类型: "1.0" or "1.1"
	AnnotationProxySSLSecret                       Annotation = "nginx.ingress.kubernetes.io/proxy-ssl-secret"                         // Service   网关使用的客户端证书，用于后端服务对网关进行身份认证    值类型: string
	AnnotationProxySSLCiphers                      Annotation = "nginx.ingress.kubernetes.io/proxy-ssl-ciphers"                        // Ingress   值类型: string
	AnnotationProxySSLName                         Annotation = "nginx.ingress.kubernetes.io/proxy-ssl-name"                           // Service   TLS握手期间使用的SNI    值类型: string
	AnnotationProxySSLProtocols                    Annotation = "nginx.ingress.kubernetes.io/proxy-ssl-protocols"                      // Ingress   值类型: string
	AnnotationProxySSLVerify                       Annotation = "nginx.ingress.kubernetes.io/proxy-ssl-verify"                         // Ingress   值类型: string
	AnnotationProxySSLVerifyDepth                  Annotation = "nginx.ingress.kubernetes.io/proxy-ssl-verify-depth"                   // Ingress   值类型: number
	AnnotationProxySSLServerName                   Annotation = "nginx.ingress.kubernetes.io/proxy-ssl-server-name"                    // Service   开启或关闭TLS握手期间使用的SNI   值类型: string
	AnnotationEnableRewriteLog                     Annotation = "nginx.ingress.kubernetes.io/enable-rewrite-log"                       // Ingress   值类型: "true" or "false"
	AnnotationRewriteRewriteTarget                 Annotation = "nginx.ingress.kubernetes.io/rewrite-target"                           // Ingress   匹配Ingress定义的路由请求在转发给后端服务时，修改头部Host值为指定值   值类型: URI
	AnnotationSatisfy                              Annotation = "nginx.ingress.kubernetes.io/satisfy"                                  // Ingress   值类型: string
	AnnotationServerAlias                          Annotation = "nginx.ingress.kubernetes.io/server-alias"                             // Ingress   值类型: string
	AnnotationServerSnippet                        Annotation = "nginx.ingress.kubernetes.io/server-snippet"                           // Ingress   值类型: string
	AnnotationServiceUpstream                      Annotation = "nginx.ingress.kubernetes.io/service-upstream"                         // Ingress   值类型: "true" or "false"
	AnnotationSessionCookieName                    Annotation = "nginx.ingress.kubernetes.io/session-cookie-name"                      // Service   配置指定Cookie的值作为Hash Key     值类型: string
	AnnotationSessionCookiePath                    Annotation = "nginx.ingress.kubernetes.io/session-cookie-path"                      // Service   当指定Cookie不存在，生成Cookie的Path值，默认为/      值类型: string
	AnnotationSessionCookieDomain                  Annotation = "nginx.ingress.kubernetes.io/session-cookie-domain"                    // Ingress   值类型: string
	AnnotationSessionCookieChangeOnFailure         Annotation = "nginx.ingress.kubernetes.io/session-cookie-change-on-failure"         // Ingress   值类型: "true" or "false"
	AnnotationSessionCookieMaxAge                  Annotation = "nginx.ingress.kubernetes.io/session-cookie-max-age"                   // Service   当指定Cookie不存在，生成Cookie的过期时间，单位为秒，默认为Session会话级别   //来自阿里云 MSE 网关 annotations 说明页面，可用性未知
	AnnotationSessionCookieExpires                 Annotation = "nginx.ingress.kubernetes.io/session-cookie-expires"                   // Service   当指定Cookie不存在，生成Cookie的过期时间，单位为秒，默认为Session会话级别   //来自阿里云 MSE 网关 annotations 说明页面，可用性未知
	AnnotationSessionCookieSameSite                Annotation = "nginx.ingress.kubernetes.io/session-cookie-samesite"                  // Ingress   值类型: string
	AnnotationSessionCookieConditionalSameSiteNone Annotation = "nginx.ingress.kubernetes.io/session-cookie-conditional-samesite-none" // Ingress   值类型: "true" or "false"
	AnnotationSSLRedirect                          Annotation = "nginx.ingress.kubernetes.io/ssl-redirect"                             // Ingress   HTTP重定向为HTTPS      值类型: "true" or "false"
	AnnotationSSLPassthrough                       Annotation = "nginx.ingress.kubernetes.io/ssl-passthrough"                          // Ingress   值类型: "true" or "false"
	AnnotationStreamSnippet                        Annotation = "nginx.ingress.kubernetes.io/stream-snippet"                           // Ingress   值类型: string
	AnnotationLoadBalanceUpstreamHashBy            Annotation = "nginx.ingress.kubernetes.io/upstream-hash-by"                         // Service   基于一致Hash的负载均衡算法。MSE 部分兼容，暂不支持Nginx变量、常量的组合使用方式。  值类型: string
	AnnotationXForwardPrefix                       Annotation = "nginx.ingress.kubernetes.io/x-forwarded-prefix"                       // Ingress   值类型: string
	AnnotationLoadBalance                          Annotation = "nginx.ingress.kubernetes.io/load-balance"                             // Service   后端服务的普通负载均衡算法。MSE 下合法值为 round_robin/least_conn/random.  默认为 round_robin。 MSE 不支持ewma算法。若配置为EWMA算法，会回退到round_robin算法。  值类型: string
	AnnotationRewriteUpstreamVHost                 Annotation = "nginx.ingress.kubernetes.io/upstream-vhost"                           // Ingress   将Ingress定义的原Path重写为指定目标，支持Group Capture。    值类型: string
	AnnotationWhiteListSourceRange                 Annotation = "nginx.ingress.kubernetes.io/whitelist-source-range"                   // Ingress   指定路由上的IP白名单，支持IP地址或CIDR地址块，以英文逗号分隔。   值类型: CIDR
	AnnotationProxyBuffering                       Annotation = "nginx.ingress.kubernetes.io/proxy-buffering"                          // Ingress   值类型: string
	AnnotationProxyBuffersNumber                   Annotation = "nginx.ingress.kubernetes.io/proxy-buffers-number"                     // Ingress   值类型: number
	AnnotationProxyBufferSize                      Annotation = "nginx.ingress.kubernetes.io/proxy-buffer-size"                        // Ingress   值类型: string
	AnnotationProxyMaxTempFileSize                 Annotation = "nginx.ingress.kubernetes.io/proxy-max-temp-file-size"                 // Ingress   值类型: string
	AnnotationSSLCiphers                           Annotation = "nginx.ingress.kubernetes.io/ssl-ciphers"                              // Ingress   指定TLS的加密套件，可以指定多个（TLS的加密套件之间使用英文逗号分隔），仅当TLS握手时采用TLSv1.0-1.2生效   值类型: string
	AnnotationSSLPreferServerCiphers               Annotation = "nginx.ingress.kubernetes.io/ssl-prefer-server-ciphers"                // Ingress   值类型: "true" or "false"
	AnnotationConnectionProxyHeader                Annotation = "nginx.ingress.kubernetes.io/connection-proxy-header"                  // Ingress   值类型: string
	AnnotationEnableAccessLog                      Annotation = "nginx.ingress.kubernetes.io/enable-access-log"                        // Ingress   值类型: "true" or "false"
	AnnotationEnableOpentracing                    Annotation = "nginx.ingress.kubernetes.io/enable-opentracing"                       // Ingress   值类型: "true" or "false"
	AnnotationOpentracingTrustIncomingSpan         Annotation = "nginx.ingress.kubernetes.io/opentracing-trust-incoming-span"          // Ingress   值类型: "true" or "false"
	AnnotationEnableInfluxDB                       Annotation = "nginx.ingress.kubernetes.io/enable-influxdb"                          // Ingress   值类型: "true" or "false"
	AnnotationInfluxDBMeasurement                  Annotation = "nginx.ingress.kubernetes.io/influxdb-measurement"                     // Ingress   值类型: string
	AnnotationInfluxDBPort                         Annotation = "nginx.ingress.kubernetes.io/influxdb-port"                            // Ingress   值类型: string
	AnnotationInfluxDBHost                         Annotation = "nginx.ingress.kubernetes.io/influxdb-host"                            // Ingress   值类型: string
	AnnotationInfluxDBServerName                   Annotation = "nginx.ingress.kubernetes.io/influxdb-server-name"                     // Ingress   值类型: string
	AnnotationUseRegex                             Annotation = "nginx.ingress.kubernetes.io/use-regex"                                // Ingress   值类型: "true" or "false"
	AnnotationEnableModsecurity                    Annotation = "nginx.ingress.kubernetes.io/enable-modsecurity"                       // Ingress   值类型: "true" or "false"
	AnnotationEnableOwaspCoreRules                 Annotation = "nginx.ingress.kubernetes.io/enable-owasp-core-rules"                  // Ingress   值类型: "true" or "false"
	AnnotationModsecurityTransactionId             Annotation = "nginx.ingress.kubernetes.io/modsecurity-transaction-id"               // Ingress   值类型: string
	AnnotationModsecuritySnippet                   Annotation = "nginx.ingress.kubernetes.io/modsecurity-snippet"                      // Ingress   值类型: string
	AnnotationMirrorRequestBody                    Annotation = "nginx.ingress.kubernetes.io/mirror-request-body"                      // Ingress   值类型: string
	AnnotationMirrorTarget                         Annotation = "nginx.ingress.kubernetes.io/mirror-target"                            // Ingress   值类型: string
	AnnotationMirrorHost                           Annotation = "nginx.ingress.kubernetes.io/mirror-host"                              // Ingress   值类型: string
)

type Annotation string

func (in Annotation) String() string {
	return string(in)
}
