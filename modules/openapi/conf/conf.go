// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package conf

import (
	"errors"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/openapi/i18n"
	"github.com/erda-project/erda/pkg/envconf"
	"github.com/erda-project/erda/pkg/strutil"
)

type Conf struct {
	ListenAddr          string `default:":9529" env:"LISTEN_ADDR"`
	RedisMasterName     string `default:"my-master" env:"REDIS_MASTER_NAME"`
	RedisSentinelAddrs  string `default:"" env:"REDIS_SENTINELS_ADDR"`
	RedisAddr           string `default:"127.0.0.1:6379" env:"REDIS_ADDR"`
	RedisPwd            string `default:"anywhere" env:"REDIS_PASSWORD"`
	UCAddrFront         string `default:"" env:"UC_PUBLIC_ADDR"`
	UCRedirectHost      string `default:"openapi.test.terminus.io" env:"SELF_PUBLIC_ADDR"`
	UCClientID          string `default:"dice" env:"UC_CLIENT_ID"`
	UCClientSecret      string `default:"secret" env:"UC_CLIENT_SECRET"`
	RedirectAfterLogin  string `default:"//dice.test.terminus.io/" env:"UI_PUBLIC_ADDR"`
	CMDBCurrentUserHost string `default:"cmdb.marathon.l4lb.thisdcos.directory:9093" env:"CMDB_ADDR"`
	CookieDomain        string `default:".terminus.io,.erda.cloud" env:"COOKIE_DOMAIN"`
	OldCookieDomain     string `default:"" env:"OLD_COOKIE_DOMAIN"`
	SessionCookieName   string `default:"OPENAPISESSION" env:"SESSION_COOKIE_NAME"`
	CSRFCookieDomain    string `default:"" env:"CSRF_COOKIE_DOMAIN"`
	UseK8S              string `env:"DICE_CLUSTER_TYPE"`
	SurveyDingding      string `env:"SURVEY_DINGDING"`
	DiceProtocol        string `env:"DICE_PROTOCOL"`
	CustomNamespace     string `env:"CUSTOM_NAMESPACE"`
	SelfPublicURL       string `env:"SELF_PUBLIC_URL" required:"true"`
	ExportUserWithRole  string `default:"false" env:"EXPORT_USER_WITH_ROLE"`

	// 修改该值的话，注意同步修改 dice.yml 中 '<%$.Storage.MountPoint%>/dice/openapi/oauth2/:/oauth2/:rw' 容器内挂载点的值
	OAuth2NetdataDir string `env:"OAUTH2_NETDATA_DIR" default:"/oauth2/"`

	CSRFWhiteList string `env:"CSRF_WHITE_LIST"`
}

var cfg Conf

// Load 加载环境变量配置
func Load() {
	envconf.MustLoad(&cfg)
	i18n.InitI18N()
}

// ListenAddr return LISTEN_ADDR
func ListenAddr() string {
	return cfg.ListenAddr
}

// RedisMasterName
func RedisMasterName() string {
	return cfg.RedisMasterName
}

func RedisSentinelAddrs() string {
	return cfg.RedisSentinelAddrs
}

func RedisAddr() string {
	return cfg.RedisAddr
}

func RedisPwd() string {
	return cfg.RedisPwd
}

func UCAddrFront() string {
	return cfg.UCAddrFront
}

func UCRedirectHost() string {
	return cfg.UCRedirectHost
}

func UCClientID() string {
	return cfg.UCClientID
}

func UCClientSecret() string {
	return cfg.UCClientSecret
}

func RedirectAfterLogin() string {
	return cfg.RedirectAfterLogin
}

func CMDBCurrentUserHost() string {
	return cfg.CMDBCurrentUserHost
}

func CookieDomain() string {
	return cfg.CookieDomain
}

func OldCookieDomain() string {
	return cfg.OldCookieDomain
}

func SessionCookieName() string {
	return cfg.SessionCookieName
}

func CSRFCookieDomain() string {
	return cfg.CSRFCookieDomain
}

func UseK8S() bool {
	return cfg.UseK8S == "kubernetes"
}

func SurveyDingding() string {
	return cfg.SurveyDingding
}

func DiceProtocol() string {
	return cfg.DiceProtocol
}

func OAuth2NetdataDir() string {
	return cfg.OAuth2NetdataDir
}

func CSRFWhiteList() []string {
	return strutil.Split(cfg.CSRFWhiteList, ",", true)
}

func CustomNamespace() string {
	return cfg.CustomNamespace
}

func SelfPublicURL() string {
	return cfg.SelfPublicURL
}

func ExportUserWithRole() bool {
	return cfg.ExportUserWithRole == "true"
}

// GetDomain get a domian by request host
func GetDomain(host, confDomain string) (string, error) {
	if strings.Contains(host, ":") {
		host = strings.SplitN(host, ":", -1)[0]
	}
	err := errors.New("invalid domain")
	domainSlice := strings.SplitN(host, ".", -1)
	l := len(domainSlice)
	if l < 2 {
		return "", err
	}
	domain := "." + domainSlice[l-2] + "." + domainSlice[l-1]
	logrus.Infof("domain is: %s", domain)

	confDomains := strings.SplitN(confDomain, ",", -1)
	for _, v := range confDomains {
		if strings.Contains(v, domain) {
			return v, nil
		}
	}

	return "", err
}

// GetUCRedirectHost get a uc redirect host by referer
func GetUCRedirectHost(referer string) string {
	rh := strings.SplitN(UCRedirectHost(), ",", -1)
	for _, v := range rh {
		domainSlice := strings.SplitN(v, ".", -1)
		l := len(domainSlice)
		if l < 2 {
			return ""
		}
		domain := domainSlice[l-2] + "." + domainSlice[l-1]
		logrus.Infof("redirect domain is: %s", domain)
		if strings.Contains(referer, domain) {
			return v
		}
	}

	return ""
}
