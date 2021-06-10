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
	"strings"

	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/envconf"
)

// Conf 定义配置对象.
type Conf struct {
	RepoRoot        string `env:"GITTAR_REPOSITORY_ROOT" default:"/repository"`
	SelfAddr        string `env:"SELF_ADDR"`
	SkipAuthUrlsStr string `env:"GITTAR_SKIP_AUTH_URL"`
	SkipAuthUrls    []string
	ListenPort      string `env:"GITTAR_PORT" default:"5566"`
	Debug           bool   `env:"DEBUG" default:"false"`

	UCAddr            string `env:"UC_ADDR"`
	UCClientID        string `env:"UC_CLIENT_ID"`
	UCClientSecret    string `env:"UC_CLIENT_SECRET"`
	UIPublicURL       string `env:"UI_PUBLIC_URL"`
	RepoPathTemplate  string `env:"REPO_PATH_TEMPLATE" default:"/workBench/projects/{{projectId}}/apps/{{appId}}/repo"`
	MergePathTemplate string `env:"MERGE_PATH_TEMPLATE" default:"/workBench/projects/{{projectId}}/apps/{{appId}}/repo/mr/open/{{mrId}}"`
	EventBoxAddr      string `env:"EVENTBOX_ADDR"`

	// git config
	GitMaxDiffLines          int    `env:"GIT_MAX_DIFF_LINES" default:"200"`
	GitMaxDiffLineCharacters int    `env:"GIT_MAX_DIFF_LINE_CHARACTERS" default:"10000"`
	GitMaxDiffFiles          int    `env:"GIT_DIFF_FILES" default:"300"`
	GitMaxDiffSize           int    `env:"GIT_MAX_DIFF_SIZE" default:"256000"`
	GitDiffContextLines      int    `env:"GIT_DIFF_CONTEXT_LINES" default:"3"`
	GitInnerUserName         string `env:"GIT_INNER_USER_NAME"`
	GitInnerUserPassword     string `env:"GIT_INNER_USER_PASSWORD"`
	GitMergeTemplatePath     string `env:"GIT_MERGE_TEMPLATE_PATH" default:".gitlab/merge_request_templates"`
	GitTokenUserName         string `env:"GIT_TOKEN_USER_NAME" default:"git"`
	GitGCMaxNum              int    `env:"GIT_GC_MAX_NUM" default:"1"`
	GitGCCronExpression      string `env:"GIT_GC_CRON_EXPRESSION" default:"0 0 1 * * ?"`

	// ory/kratos config
	OryEnabled           bool   `default:"false" env:"ORY_ENABLED"`
	OryKratosAddr        string `default:"kratos:4433" env:"KRATOS_ADDR"`
	OryKratosPrivateAddr string `default:"kratos:4434" env:"KRATOS_PRIVATE_ADDR"`
}

var cfg Conf

// Load 从环境变量加载配置选项.
func Load() {
	envconf.MustLoad(&cfg)
	cfg.SkipAuthUrls = strings.Split(cfg.SkipAuthUrlsStr, ",")
	cfg.SkipAuthUrls = append(cfg.SkipAuthUrls, cfg.SelfAddr)
	cfg.SkipAuthUrls = append(cfg.SkipAuthUrls, discover.Gittar())
	cfg.SkipAuthUrls = append(cfg.SkipAuthUrls, "gittar:5566")
}

// RepoRoot 仓库存储目录
func RepoRoot() string {
	return cfg.RepoRoot
}

// GittarUrl gittar 内网url
func GittarUrl() string {
	return "http://" + cfg.SelfAddr
}

// SkipAuthUrls 不做权限校验的url
func SkipAuthUrls() []string {
	return cfg.SkipAuthUrls
}

// ListenPort 监听端口
func ListenPort() string {
	return cfg.ListenPort
}

// Debug 是否开启Debug模式.
func Debug() bool {
	return cfg.Debug
}

// UCAddr 返回UC的地址
func UCAddr() string {
	return cfg.UCAddr
}

// UCClientID 返回UC的ClientID
func UCClientID() string {
	return cfg.UCClientID
}

// UCClientSecret 返回UC ClientSecret
func UCClientSecret() string {
	return cfg.UCClientSecret
}

// UIPublicURL UI URL
func UIPublicURL() string {
	return cfg.UIPublicURL
}

// RepoPathTemplate 在线代码仓库path模板,用于从git url直接跳转到dice ui url
func RepoPathTemplate() string {
	return cfg.RepoPathTemplate
}

// MergePathTemplate 在线代码仓库merge地址模板
func MergePathTemplate() string {
	return cfg.MergePathTemplate
}

// EventBoxAddr 返回 eventbox 地址
func EventBoxAddr() string {
	return cfg.EventBoxAddr
}

// GitMaxDiffLineCharacters 单行diff最大字符串
func GitMaxDiffLineCharacters() int {
	return cfg.GitMaxDiffLineCharacters
}

// GitMaxDiffFiles diff最大文件数
func GitMaxDiffFiles() int {
	return cfg.GitMaxDiffFiles
}

// GitMaxDiffSize 最大diff的文件大小,单位Byte
func GitMaxDiffSize() int {
	return cfg.GitMaxDiffSize
}

// GitDiffContextLines diff显示的上下文关联行数
func GitDiffContextLines() int {
	return cfg.GitDiffContextLines
}

// GitDiffContextLines 最大diff行数
func GitMaxDiffLines() int {
	return cfg.GitMaxDiffLines
}

// GitInnerUserName 内部用户名
func GitInnerUserName() string {
	return cfg.GitInnerUserName
}

// GitInnerUserPassword 内部用户名密码
func GitInnerUserPassword() string {
	return cfg.GitInnerUserPassword
}

// GitMergeTemplatePath merge模板文件对应路径
func GitMergeTemplatePath() string {
	return cfg.GitMergeTemplatePath
}

// GitTokenUserName token认证使用的用户名
func GitTokenUserName() string {
	return cfg.GitTokenUserName
}

// GitGCMaxNum  git repository gc Concurrency
func GitGCMaxNum() int {
	return cfg.GitGCMaxNum
}

// GitGCCronExpression cron run gc
func GitGCCronExpression() string {
	return cfg.GitGCCronExpression
}

func OryEnabled() bool {
	return cfg.OryEnabled
}

func OryKratosAddr() string {
	return cfg.OryKratosAddr
}

func OryCompatibleClientID() string {
	return "kratos"
}

func OryCompatibleClientSecret() string {
	return ""
}

func OryKratosPrivateAddr() string {
	return cfg.OryKratosPrivateAddr
}
