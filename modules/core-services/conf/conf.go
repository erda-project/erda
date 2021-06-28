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

// Package conf 定义配置选项
package conf

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/pkg/envconf"
	"github.com/erda-project/erda/pkg/strutil"
)

// Conf 定义基于环境变量的配置项
type Conf struct {
	LocalMode             bool   `env:"LOCAL_MODE" default:"false"`
	Debug                 bool   `env:"DEBUG" default:"false"`
	ListenAddr            string `env:"LISTEN_ADDR" default:":9526"`
	MySQLHost             string `env:"MYSQL_HOST"`
	MySQLPort             string `env:"MYSQL_PORT"`
	MySQLUsername         string `env:"MYSQL_USERNAME"`
	MySQLPassword         string `env:"MYSQL_PASSWORD"`
	MySQLDatabase         string `env:"MYSQL_DATABASE"`
	MySQLLoc              string `env:"MYSQL_LOC" default:"Local"`
	GittarOutterURL       string `env:"GITTAR_PUBLIC_URL"`
	UCClientID            string `env:"UC_CLIENT_ID"`
	UCClientSecret        string `env:"UC_CLIENT_SECRET"`
	RootDomain            string `env:"DICE_ROOT_DOMAIN"`
	UIPublicURL           string `env:"UI_PUBLIC_URL"`
	UIDomain              string `env:"UI_PUBLIC_ADDR"`
	OpenAPIDomain         string `env:"OPENAPI_PUBLIC_ADDR"` // Deprecated: after cli refactored
	AvatarStorageURL      string `env:"AVATAR_STORAGE_URL"`  // file:///avatars or oss://appkey:appsecret@endpoint/bucket
	LicenseKey            string `env:"LICENSE_KEY"`
	AuditCleanCron        string `env:"AUDIT_CLEAN_CRON" default:"0 0 3 * * ?"`   // 审计软删除任务执行周期
	AuditArchiveCron      string `env:"AUDIT_ARCHIVE_CRON" default:"0 0 4 * * ?"` // 审计归档任务执行周期
	SysAuditCleanIterval  int    `env:"SYS_AUDIT_CLEAN_ITERVAL" default:"-7"`     // 系统审计清除周期
	RedisMasterName       string `default:"my-master" env:"REDIS_MASTER_NAME"`
	RedisSentinelAddrs    string `default:"" env:"REDIS_SENTINELS_ADDR"`
	RedisAddr             string `default:"127.0.0.1:6379" env:"REDIS_ADDR"`
	RedisPwd              string `default:"anywhere" env:"REDIS_PASSWORD"`
	ProjectStatsCacheCron string `env:"PROJECT_STATS_CACHE_CRON" default:"0 0 1 * * ?"`
	EnableProjectNS       bool   `env:"ENABLE_PROJECT_NS" default:"true"`
	LegacyUIDomain        string `env:"LEGACY_UI_PUBLIC_ADDR"`

	// ory/kratos config
	OryEnabled           bool   `default:"false" env:"ORY_ENABLED"`
	OryKratosAddr        string `default:"kratos:4433" env:"KRATOS_ADDR"`
	OryKratosPrivateAddr string `default:"kratos:4434" env:"KRATOS_PRIVATE_ADDR"`

	// Allow people who are not admin to create org
	CreateOrgEnabled bool `default:"false" env:"CREATE_ORG_ENABLED"`
}

var (
	cfg Conf
	// 存储权限配置
	permissions []model.RolePermission
	// 审计模版配置
	auditsTemplate apistructs.AuditTemplateMap
	// 域名白名单
	OrgWhiteList map[string]bool
	// legacy redirect paths
	RedirectPathList map[string]bool
)

func initPermissions() {
	permissions = getAllFiles("pkg/erda-configs/permission", permissions)
}

func initAuditTemplate() {
	auditsTemplate = genTempFromFiles("pkg/erda-configs/audit/template.json")
}

func genTempFromFiles(fileName string) apistructs.AuditTemplateMap {
	templateJSON, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	var result apistructs.AuditTemplateMap
	if err := json.Unmarshal(templateJSON, &result); err != nil {
		panic(err)
	}

	for _, v := range result {
		v.ConvertContent2GoTemplateFormart()
	}

	return result
}

func getAllFiles(pathname string, perms []model.RolePermission) []model.RolePermission {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		panic(err)
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			perms = getAllFiles(fullDir, perms)
		} else {
			fullName := pathname + "/" + fi.Name()
			yamlFile, err := ioutil.ReadFile(fullName)
			if err != nil {
				panic(err)
			}
			var per []model.RolePermission
			if err := yaml.Unmarshal(yamlFile, &per); err != nil {
				panic(err)
			}
			perms = append(perms, per...)
		}
	}
	return perms
}

// Load 加载配置项.
func Load() {
	initPermissions()
	initAuditTemplate()
	envconf.MustLoad(&cfg)

	OrgWhiteList = map[string]bool{
		UIDomain():                          true,
		LegacyUIDomain():                    true,
		OpenAPIDomain():                     true,
		"openapi.default.svc.cluster.local": true,
	}

	RedirectPathList = map[string]bool{
		"/microService": true,
		"/workBench":    true,
		"/dataCenter":   true,
		"/orgCenter":    true,
		"/edge":         true,
		"/sysAdmin":     true,
		"/org-list":     true,
		"/noAuth":       true,
		"/freshMan":     true,
		"/inviteToOrg":  true,
		"/perm":         true,
	}
}

// AuditTemplate 返回权限列表
func AuditTemplate() apistructs.AuditTemplateMap {
	return auditsTemplate
}

// Permissions 获取权限配置
func Permissions() map[string]model.RolePermission {
	pm := make(map[string]model.RolePermission, len(permissions))
	for _, v := range permissions {
		k := strutil.Concat(v.Scope, v.Resource, v.Action)
		pm[k] = v
	}
	return pm
}

// RolePermissions 获取角色对应的权限配置
func RolePermissions(roles []string) (map[string]model.RolePermission, []model.RolePermission) {
	pm := make(map[string]model.RolePermission, len(permissions))
	resourceRoles := make([]model.RolePermission, 0)
	for _, v := range permissions {
		for _, role := range roles {
			confRoles := strings.SplitN(v.Role, ",", -1)
			for _, cR := range confRoles {
				if role == cR {
					k := strutil.Concat(v.Scope, v.Resource, v.Action)
					pm[k] = v
				}
			}

			if v.ResourceRole != "" {
				resourceRoles = append(resourceRoles, v)
			}
		}
	}
	return pm, resourceRoles
}

// ListenAddr 返回 ListenAddr 选项.
func ListenAddr() string {
	return cfg.ListenAddr
}

// LocalMode 本地调试模式
func LocalMode() bool {
	return cfg.LocalMode
}

// Debug 返回 Debug 选项.
func Debug() bool {
	return cfg.Debug
}

// MySQLHost 返回 MySQLHost 选项.
func MySQLHost() string {
	return cfg.MySQLHost
}

// MySQLPort 返回 MySQLPort 选项.
func MySQLPort() string {
	return cfg.MySQLPort
}

// MySQLUsername 返回 MySQLUsername 选项.
func MySQLUsername() string {
	return cfg.MySQLUsername
}

// MySQLPassword 返回 MySQLPassword 选项.
func MySQLPassword() string {
	return cfg.MySQLPassword
}

// MySQLDatabase 返回 MySQLDatabase 选项.
func MySQLDatabase() string {
	return cfg.MySQLDatabase
}

// MySQLLoc 返回 MySQLLoc 选项.
func MySQLLoc() string {
	return cfg.MySQLLoc
}

// GittarOutterURL 返回 GittarOutterURL 选项.
func GittarOutterURL() string {
	return cfg.GittarOutterURL
}

// UCClientID 返回 UCClientID 选项.
func UCClientID() string {
	return cfg.UCClientID
}

// UCClientSecret 返回 UCClientSecret 选项.
func UCClientSecret() string {
	return cfg.UCClientSecret
}

// RootDomain 返回 RootDomain 选项
func RootDomain() string {
	return RootDomainList()[0]
}

// Multiple domain
func RootDomainList() []string {
	return strutil.Split(cfg.RootDomain, ",")
}

// UIPublicURL 返回 UIPublicURL 选项
func UIPublicURL() string {
	return cfg.UIPublicURL
}

// UIDomain 返回 UIDomain 选项
func UIDomain() string {
	return cfg.UIDomain
}

// LegacyUIDomain
func LegacyUIDomain() string {
	return cfg.LegacyUIDomain
}

// OpenAPIDomain 返回 OpenAPIDomain 选项
func OpenAPIDomain() string {
	return cfg.OpenAPIDomain
}

// AvatarStorageURL 返回 OSSUsage 选项
func AvatarStorageURL() string {
	return cfg.AvatarStorageURL
}

// LicenseKey 返回 LicenseKey 选项.
func LicenseKey() string {
	return cfg.LicenseKey
}

// AuditCleanCron 返回审计事件软删除周期
func AuditCleanCron() string {
	return cfg.AuditCleanCron
}

// AuditArchiveCron 返回审计事件归档周期
func AuditArchiveCron() string {
	return cfg.AuditArchiveCron
}

// SysAuditCleanIterval 返回 sys scope 审计事件软删除周期
func SysAuditCleanIterval() int {
	return cfg.SysAuditCleanIterval
}

// RedisMasterName 返回redis master name
func RedisMasterName() string {
	return cfg.RedisMasterName
}

// RedisSentinelAddrs 返回 redis 哨兵地址
func RedisSentinelAddrs() string {
	return cfg.RedisSentinelAddrs
}

// RedisAddr 返回 redis 地址
func RedisAddr() string {
	return cfg.RedisAddr
}

// RedisPwd 返回 redis 密码
func RedisPwd() string {
	return cfg.RedisPwd
}

// ProjectStatsCacheCron 项目状态缓存刷新周期
func ProjectStatsCacheCron() string {
	return cfg.ProjectStatsCacheCron
}

// EnableNS 是否打开项目级命名空间
func EnableNS() bool {
	return cfg.EnableProjectNS
}

func OryEnabled() bool {
	return cfg.OryEnabled
}

func OryKratosPrivateAddr() string {
	return cfg.OryKratosPrivateAddr
}

func OryCompatibleClientID() string {
	return "kratos"
}

func OryCompatibleClientSecret() string {
	return ""
}

func CreateOrgEnabled() bool {
	return cfg.CreateOrgEnabled
}
