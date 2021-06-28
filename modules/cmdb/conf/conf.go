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
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/envconf"
	"github.com/erda-project/erda/pkg/http/httpclientutil"
	"github.com/erda-project/erda/pkg/strutil"
)

// Conf 定义基于环境变量的配置项
type Conf struct {
	LocalMode             bool          `env:"LOCAL_MODE" default:"false"`
	Debug                 bool          `env:"DEBUG" default:"false"`
	ListenAddr            string        `env:"LISTEN_ADDR" default:":9093"`
	KafkaBrokers          string        `env:"BOOTSTRAP_SERVERS"`
	KafkaContainerTopic   string        `env:"CMDB_CONTAINER_TOPIC"`
	KafkaHostTopic        string        `env:"CMDB_HOST_TOPIC"`
	KafkaGroup            string        `env:"CMDB_GROUP"`
	MySQLHost             string        `env:"MYSQL_HOST"`
	MySQLPort             string        `env:"MYSQL_PORT"`
	MySQLUsername         string        `env:"MYSQL_USERNAME"`
	MySQLPassword         string        `env:"MYSQL_PASSWORD"`
	MySQLDatabase         string        `env:"MYSQL_DATABASE"`
	MySQLLoc              string        `env:"MYSQL_LOC" default:"Local"`
	DicehubAddr           string        `env:"DICEHUB_ADDR"`
	QAAddr                string        `env:"QA_ADDR"`
	AddonAddr             string        `env:"ADDON_PLATFORM_ADDR"`
	GittarAddr            string        `env:"GITTAR_ADDR"`
	GittarOutterURL       string        `env:"GITTAR_PUBLIC_URL"`
	UCAddr                string        `env:"UC_ADDR"`
	UCClientID            string        `env:"UC_CLIENT_ID"`
	UCClientSecret        string        `env:"UC_CLIENT_SECRET"`
	EventBoxAddr          string        `env:"EVENTBOX_ADDR"`
	HepaAddr              string        `env:"HEPA_ADDR"`
	RootDomain            string        `env:"DICE_ROOT_DOMAIN"`
	UIPublicURL           string        `env:"UI_PUBLIC_URL"`
	UIDomain              string        `env:"UI_PUBLIC_ADDR"`
	OpenAPIDomain         string        `env:"OPENAPI_PUBLIC_ADDR"` // Deprecated: after cli refactored
	AvatarStorageURL      string        `env:"AVATAR_STORAGE_URL"`  // file:///avatars or oss://appkey:appsecret@endpoint/bucket
	LicenseKey            string        `env:"LICENSE_KEY"`
	HostSyncInterval      time.Duration `env:"INTERVAL" default:"2m"`                      // 主机实际资源使用同步间隔
	TaskSyncDuration      time.Duration `env:"TASK_SYNC_DURATION" default:"2h"`            // 任务状态信息同步间隔
	TaskCleanDuration     time.Duration `env:"TASK_CLEAN_DURATION" default:"24h"`          // 任务信息回收间隔
	AuditCleanCron        string        `env:"AUDIT_CLEAN_CRON" default:"0 0 3 * * ?"`     // 审计软删除任务执行周期
	AuditArchiveCron      string        `env:"AUDIT_ARCHIVE_CRON" default:"0 0 4 * * ?"`   // 审计归档任务执行周期
	MetricsIssueCron      string        `env:"METRICS_ISSUE_CRON" default:"0 0 0 1/7 * ?"` // metrics issue report monitor execution cycle
	SysAuditCleanIterval  int           `env:"SYS_AUDIT_CLEAN_ITERVAL" default:"-7"`       // 系统审计清除周期
	RedisMasterName       string        `default:"my-master" env:"REDIS_MASTER_NAME"`
	RedisSentinelAddrs    string        `default:"" env:"REDIS_SENTINELS_ADDR"`
	RedisAddr             string        `default:"127.0.0.1:6379" env:"REDIS_ADDR"`
	RedisPwd              string        `default:"anywhere" env:"REDIS_PASSWORD"`
	ProjectStatsCacheCron string        `env:"PROJECT_STATS_CACHE_CRON" default:"0 0 1 * * ?"`
	EnableProjectNS       bool          `env:"ENABLE_PROJECT_NS" default:"true"`
	LegacyUIDomain        string        `env:"LEGACY_UI_PUBLIC_ADDR"`

	// --- 文件管理 begin ---
	FileMaxUploadSizeStr string `env:"FILE_MAX_UPLOAD_SIZE" default:"300MB"` // 文件上传限制大小，默认 300MB
	FileMaxUploadSize    datasize.ByteSize
	// the size of the file parts stored in memory, the default value 32M refer to https://github.com/golang/go/blob/5c489514bc5e61ad9b5b07bd7d8ec65d66a0512a/src/net/http/request.go
	FileMaxMemorySizeStr string `env:"FILE_MAX_MEMORY_SIZE" default:"32MB"`
	FileMaxMemorySize    datasize.ByteSize

	// disable file download permission validate temporarily for multi-domain
	DisableFileDownloadPermissionValidate bool `env:"DISABLE_FILE_DOWNLOAD_PERMISSION_VALIDATE" default:"false"`

	// fs
	// 修改该值的话，注意同步修改 dice.yml 中 '<%$.Storage.MountPoint%>/dice/cmdb/files:/files:rw' 容器内挂载点的值
	StorageMountPointInContainer string `env:"STORAGE_MOUNT_POINT_IN_CONTAINER" default:"/files"`

	// oss
	OSSEndpoint     string `env:"OSS_ENDPOINT"`
	OSSAccessID     string `env:"OSS_ACCESS_ID"`
	OSSAccessSecret string `env:"OSS_ACCESS_SECRET"`
	OSSBucket       string `env:"OSS_BUCKET"`
	OSSPathPrefix   string `env:"OSS_PATH_PREFIX" default:"/dice/cmdb/files"`
	// --- 文件管理 end ---

	CentralNexusPublicURL string `env:"NEXUS_PUBLIC_URL" required:"true"`
	CentralNexusAddr      string `env:"NEXUS_ADDR" required:"true"`
	CentralNexusUsername  string `env:"NEXUS_USERNAME" required:"true"`
	CentralNexusPassword  string `env:"NEXUS_PASSWORD" required:"true"`

	DiceClusterName string `env:"DICE_CLUSTER_NAME" required:"true"`

	// rsa
	Base64EncodedRsaPublicKey  string `env:"BASE64_ENCODED_RSA_PUBLIC_KEY" default:"LS0tLS1CRUdJTiBwdWJsaWMga2V5LS0tLS0KTUlJQ0lqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FnOEFNSUlDQ2dLQ0FnRUFrOCtVK3QyeHhoM1hpREJnRjM2dApxWU5UZmN2NDA4aTdsZnFZRG9TRHMxbDA5bitsLzFOZTQ5b0xxZ0h1ZTQ5MmJHNFI0T0ZHZW1IMktIZmUya3BnCjZpd2tFM0xrZW5KMm56NFdPQWNnOUhiWlA0TFpReGxoeUVwNlE2aHQyekgxZ25Uc2p0QUlzMEZxbXJXZmlVVkQKdFdib1lmSDMvNWZReSs3V00yWkU3bzdnWWxIM1RLR2M5amEvWmgwOTBUZXdULzV3TVhPb1llcFRsWVBmTDVoTwo0em9GeGFpbzltanhpQmVveDNrUkM5RlZsSFM4ZDVlYWRHNkttR2cydjlTaE96SThDaGErRkJHSm83b3E4UEZEClRFMUFuZnBjZml5ckVxVVpzbDZTckl1TjVZUTREM3h1clZnY1RkcG9MV1dpallJbVZ0bytJU3FScW9QemxqVWQKTzdDa2NVRXUvVno2UCt2Vjc4b1JWRktYM0E0aG9vYlFFSkphNlFISmlzN1JQRW5TTjZXS2k4RXkzSlFhT3hXWAppejR3aDk3VmIyZDU4c3l1M0pJSTFOWVlyemtqTitEd1RLV1dqcjVYaVhHSGVCRDFtMmpaMytxV1RCTW1oNC9QCmtWc2M0T29lOG40ZXFoYVc1d2QyaU5jUlRHUS9sUmY4ekNSRlhCN1lvbWJrVlQwc1hVcllXQWFkWURFUEFmazUKTncvUjJaTXkyNGVhd0ZCcTVmYVB6VVJWRUY4WC9uUm5kL1YwUFZBSGgySG9CeFJaZzFkSGJrSWQ3SUo5R2cxbwpKVzJZOTlobzRpK0QvTDl2cWNPOVRyOXN0dStWcG1UQ1BRdFZqWHlpY0FuZmN4MWxhOEI0Q2Y4azhWN1RBSmJWCm14SjdaUTJEbGs3TTdBYzNTamVEUmJrQ0F3RUFBUT09Ci0tLS0tRU5EIHB1YmxpYyBrZXktLS0tLQo="`
	Base64EncodedRsaPrivateKey string `env:"BASE64_ENCODED_RSA_PRIVATE_KEY" default:"LS0tLS1CRUdJTiBwcml2YXRlIGtleS0tLS0tCk1JSUpLUUlCQUFLQ0FnRUFrOCtVK3QyeHhoM1hpREJnRjM2dHFZTlRmY3Y0MDhpN2xmcVlEb1NEczFsMDluK2wKLzFOZTQ5b0xxZ0h1ZTQ5MmJHNFI0T0ZHZW1IMktIZmUya3BnNml3a0UzTGtlbkoybno0V09BY2c5SGJaUDRMWgpReGxoeUVwNlE2aHQyekgxZ25Uc2p0QUlzMEZxbXJXZmlVVkR0V2JvWWZIMy81ZlF5KzdXTTJaRTdvN2dZbEgzClRLR2M5amEvWmgwOTBUZXdULzV3TVhPb1llcFRsWVBmTDVoTzR6b0Z4YWlvOW1qeGlCZW94M2tSQzlGVmxIUzgKZDVlYWRHNkttR2cydjlTaE96SThDaGErRkJHSm83b3E4UEZEVEUxQW5mcGNmaXlyRXFVWnNsNlNySXVONVlRNApEM3h1clZnY1RkcG9MV1dpallJbVZ0bytJU3FScW9QemxqVWRPN0NrY1VFdS9WejZQK3ZWNzhvUlZGS1gzQTRoCm9vYlFFSkphNlFISmlzN1JQRW5TTjZXS2k4RXkzSlFhT3hXWGl6NHdoOTdWYjJkNThzeXUzSklJMU5ZWXJ6a2oKTitEd1RLV1dqcjVYaVhHSGVCRDFtMmpaMytxV1RCTW1oNC9Qa1ZzYzRPb2U4bjRlcWhhVzV3ZDJpTmNSVEdRLwpsUmY4ekNSRlhCN1lvbWJrVlQwc1hVcllXQWFkWURFUEFmazVOdy9SMlpNeTI0ZWF3RkJxNWZhUHpVUlZFRjhYCi9uUm5kL1YwUFZBSGgySG9CeFJaZzFkSGJrSWQ3SUo5R2cxb0pXMlk5OWhvNGkrRC9MOXZxY085VHI5c3R1K1YKcG1UQ1BRdFZqWHlpY0FuZmN4MWxhOEI0Q2Y4azhWN1RBSmJWbXhKN1pRMkRsazdNN0FjM1NqZURSYmtDQXdFQQpBUUtDQWdCSlhxbngyS2ZNMHJWUTJjcG8veTJPeml4Y2Jpb21YaWFYTE52Ym9QV0t5aVhmMGI4QlBVNEZ4Zzh5CkpXRk9uZ2pIaTk5K0EvU3EvUU5tVlJJZXd2cldZbkRKNHFiOURPSks2MU8ySGZ2Q3ZWZmJTY1UwcEYzQVFRL3QKazZac1BxRkNUMjI0K2hUSGZmby9yMVh3bXB3Z2FHT0Rjc3VLYUw1dzdDNFJOM3VSK3dQd2FnVmFXWUtEU091Ngo4VnJsQmtLVGdwWUlSZ1BZRHF2TXRMZk5kVW43U3FyZzBYYUZVZFJLbkl2ZjcvMkJJempheHhOaVBiT2loZGh3CkRKTFlwK0FjZFRRT1FmbTZGblorK2dNa3RHMldhMlpleEk2eTV0TklId0hoWTBabE5hU0t3Qlhmd2dGaU5ERmcKaDhCY2dHMnUxbUxYaTk5NU1SczdTK0pXdGlpNjNqYUxsMmN0eXIxYVJPNlJRdmtYMlgzbTU4MFRKODVwZ0dBbgpYY3hENW1HNTRsRFlqdTllbVlRUUZlRDNrQXQzV3pGWmhkWXFtUEU4VEk1clBCbGtxNzkwVzF3K3BveUpHc0VoCjJIZXlMekQ5WGNSdXBoVE1QaEFXUTlBMzVHbnBOM0NrUXA0c1lBSnlNck5FQWhHQ285cXVGZW5rRDMxQmFPSjMKQWN1MDBISVpyUy9JTkRZNGJ4MmExd3RFbnd4dDkxSnJ4R1I5eFFtZE5DejFLZkR2Z2laSldTR2ZPYXhkbmtoMgp0WHkrYkl1WER0ZVhsK3dDN0wySG9PTmZKUjM5RDNPTjVLdzRvbTdlMjVSM3dKcEN2UGFFR01ELzZ0YWt6N1dCCnVEbmVBTHRSb1ZkSzhsK1RuaGVKcmtQRG1tUnkwL1BPRzgxWXp3VVlEVCs1VmtuTnNRS0NBUUVBd3paQ0JWTk4KaGlza3psTmdRTTRZTkVLT3lLUkcwL05LUlBIM1lzVUZyeDJEQjhESzVITWJCSVNqQ0ZrQUhJRXI1R1VlVVUwRgpCMndURFFScnQ0KzZYcG92aE5SZmdOQWZodXZDdXZNajk2eG05N3hpa01GeDI3ZW9VcVQ0amlKSU5NTG12MXh1CnY1UTNGbnhhQVgxdE05UHpEbHJQN3ZXaWsrWWprZFRZZlFLL0dHdU90d2RwdC96V2xXREtaSmp6RmozamtVNUoKK3lNcjVlbDBSdVJFQlZkaHdHTkRqWHRGNXg0UkQ2YWpXRXFtNEJ4VXBHeGM3UEZ2NU96d0NMNm50NkxTUGwzMwo0OGpOVVhsYWhESk0rUStQeHBBR2UxTXVUM2dTdmNaVk5hMG83eXNuZHE2b3VaSUx5VFZ3S0xoZG1kbnpZY1BWCmpKL0lFOTdZR0RBblZRS0NBUUVBd2RhbENoWDlacFVGalFOWG1KcnUyaGwxck9UYmE1V3FCWGh1ME91bnNHaG4KUkN2YjgyckQ4dnR4VVNLV1drSTd3cFFEakVxOVQyVmRRbmswR3ZVK3lIUW1WVE9MenNTcVlBRk0rdnUrWnJ0UApMcWxSa05lYll4UGx0TkwyVWJQYUMrOVpWblNOYUJDeEpOKzBwcTJnZG5IRWd5YWRJQ1FUVE41OGFDSVN4SGg5ClRqQ2N6Njd1RDMrU3FTVHVEZHFnQUhmeDlNTytjb3JQaW1Zc2c4d2FwZ3JtR0paMkpRWC9QdllLUlkrVXFoSngKd2VBaGlYVTFDY2Y3eWpMUEdWZURrZTQwTktEcGFNZjJnYW9qakgxei8yaERuU3A2enc4Wnc1VWg3dlA2azBkeQpvYzVzQ0FxYURWNkZEMG4xNWF5Q1RVWG1EQU1VTWU1Qlh5SjVCQkJjMVFLQ0FRRUFwNVcrMjkrRjREYk5wQ3RECnFKN0ZmS2ZlK0RTL2NWbWRXczcyOTkzNFlUdE9yNnM5QXg0bUJaendjVXdtb2xIcUltc0V1ZnNLNURKTnNKRXAKQUM3dGFpV253YnFvT21keGlWeUFrZ29GeUt4Q3dVOEN0dzY2OWtzV3Y4eE1iWWpVd0NiSi9XSVcyWFVlVGJsMwpjMndBQWN4bER0KzdQb08xakk2MzNvd0JSbURET08ydFdVZU41SnUwaEF6Ujg4YXllVmVzTTZRb011Y2cyb0d1Cmh1V1QxNW9LbXlVY2F5dDIrVkNBaVJVZmliNmN3Q3pTSlUyNkFOZk1uWlVqQS83WThQZGcwcFhOSjhuTktiS3EKbUc2dVVlcWdIWENyZjlnTEc4SVRKTVJOaG9VZmJTTjQvNVExMlFtZUFLQlZweitQYTNNR1U5blJUS1luRjVmcApuK3BHK1FLQ0FRRUF3UTMrb2VUMDFFNW5rT0piUStwTEtYMWg3aWloUUsxM0FLdko4dHBCMFRpcVlRTXR0V29JCmJ1QnZJOWZHMTI1UUJxTlVSVTNLN21DT1diNU5YdXdTODZKNjZ6RERkZFA1dkZTUFR3bWJ3TVdkUDJQemtNYXMKUkNsMUJudDJTRGxRV2NLd3Y2S2xrNWZNVm1WWGp3b3VYc2xBWno3MkR5VGU5QmhDMzVQUURVM1R2eVE3aWIwMwo3TWVxVWp3dHZDNmFYTjBaWmlYdWNEWkFMaDlGQnA4cGkyWWZkUzJsellvRGhibVcwV0VITjd2WEFMa3hyYTNHCmZVOW9QeUlMa2JuUG1IQWVIcXlFeTQ4Y3ZGZXZ3Q1RTZXZabElRdEY5U09kRFdaaXZaTFJaZzRxNVd5cHUvaVQKSmUyVnFIeUpJNDZFMkdGZGxXa2JtLzhuckpDdzVwTkZZUUtDQVFCWTNjKzVnY0dkcGJoYkRjenlHNngrcHJVbwpWclVlanNJTUlIZEh1MUUzWW5rWVdrL2hMWEh4anYrbW14RTVOb0hpMmJSTG53VXZhelhReW1CV2l2cW9FRGtmClZsMGR1Tk5xZmpWcG1JY092U1Rkb0wwRHlsZVExT2JEUjBZMEhMZmNvUnJLSUhzOE13b3BWWGtIZjFzamdxVEIKMjNycHd6aDVBTHdaMXJNVDJGNC9CL3RRdlBBMzRTeGlhanQ0RlBHU3pwUE50b0trU1ZOTnM3TS96SXFmY1JtZAppNjNhcW1Tc2pJWEdrZWY1cGpZZTV2WHM2ODZkVldMVGZ4SllMa0N6TkZ4c1dZdXBFaWhZQVJQbnViUGpNd21QCkVWcDBiQTdJNlpnQUQ3Vmw1TkR3UGg4bFZYUlYzWHVyV2dFVXUrcytqMHNTMEZyS2tUeldsRmF1aUhmawotLS0tLUVORCBwcml2YXRlIGtleS0tLS0tCg=="`

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
	// 审计模版配置
	auditsTemplate apistructs.AuditTemplateMap
	// 域名白名单
	OrgWhiteList map[string]bool
	// legacy redirect paths
	RedirectPathList map[string]bool
)

func initAuditTemplate() {
	auditsTemplate = genTempFromFiles("erda-configs/audit/template.json")
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

// Load 加载配置项.
func Load() {
	initAuditTemplate()
	envconf.MustLoad(&cfg)

	// parse FileMaxUploadSize
	var fileMaxUploadByte datasize.ByteSize
	if err := fileMaxUploadByte.UnmarshalText([]byte(cfg.FileMaxUploadSizeStr)); err != nil {
		panic(fmt.Sprintf("failed to parse FILE_MAX_UPLOAD_SIZE, err: %v", err))
	}
	fmt.Println(fileMaxUploadByte.String())
	cfg.FileMaxUploadSize = fileMaxUploadByte

	// parse FileMaxMemorySize
	var fileMaxMemoryByte datasize.ByteSize
	if err := fileMaxMemoryByte.UnmarshalText([]byte(cfg.FileMaxMemorySizeStr)); err != nil {
		panic(fmt.Sprintf("failed to parse FILE_MAX_MEMORY_SIZE, err: %v", err))
	}
	cfg.FileMaxMemorySize = fileMaxMemoryByte

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

// KafkaBrokers 返回 KafkaBrokers 选项.
func KafkaBrokers() string {
	return cfg.KafkaBrokers
}

// KafkaContainerTopic 返回 KafkaContainerTopic 选项.
func KafkaContainerTopic() string {
	return cfg.KafkaContainerTopic
}

// KafkaHostTopic 返回 KafkaHostTopic 选项.
func KafkaHostTopic() string {
	return cfg.KafkaHostTopic
}

// KafkaGroup 返回 KafkaGroup 选项.
func KafkaGroup() string {
	return cfg.KafkaGroup
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

// DicehubAddr 返回 DicehubAddr 选项.
func DicehubAddr() string {
	return cfg.DicehubAddr
}

// AddonAddr 返回 AddonAddr 选项.
func AddonAddr() string {
	return cfg.AddonAddr
}

// GittarAddr 返回 GittarAddr 选项
func GittarAddr() string {
	return cfg.GittarAddr
}

// GittarOutterURL 返回 GittarOutterURL 选项.
func GittarOutterURL() string {
	return cfg.GittarOutterURL
}

// UCAddr 返回 UCAddr 选项.
func UCAddr() string {
	return cfg.UCAddr
}

// UCClientID 返回 UCClientID 选项.
func UCClientID() string {
	return cfg.UCClientID
}

// UCClientSecret 返回 UCClientSecret 选项.
func UCClientSecret() string {
	return cfg.UCClientSecret
}

// EventBoxAddr 返回 EventBoxAddr 选项.
func EventBoxAddr() string {
	return cfg.EventBoxAddr
}

// HepaAddr 返回 HepaAddr 选项.
func HepaAddr() string {
	return cfg.HepaAddr
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

// HostSyncInterval 返回 HostSyncInterval 选项
func HostSyncInterval() time.Duration {
	return cfg.HostSyncInterval
}

// TaskSyncDuration 返回 TaskSyncDuration 时间间隔
func TaskSyncDuration() time.Duration {
	return cfg.TaskSyncDuration
}

// TaskCleanDuration 返回 TaskCleanDuration 时间间隔
func TaskCleanDuration() time.Duration {
	return cfg.TaskCleanDuration
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

// FileMaxUploadSize 返回 文件上传的大小限制.
func FileMaxUploadSize() datasize.ByteSize {
	return cfg.FileMaxUploadSize
}

// FileMaxMemorySize return the size of the file parts stored in memory
func FileMaxMemorySize() datasize.ByteSize {
	return cfg.FileMaxMemorySize
}

// DisableFileDownloadPermissionValidate return switch for file download permission check.
func DisableFileDownloadPermissionValidate() bool {
	return cfg.DisableFileDownloadPermissionValidate
}

// StorageMountPointInContainer 返回 files 在容器内的挂载点.
func StorageMountPointInContainer() string {
	return cfg.StorageMountPointInContainer
}

// OSSEndpoint 返回 oss endpoint.
func OSSEndpoint() string {
	return cfg.OSSEndpoint
}

// OSSAccessID 返回 oss access id.
func OSSAccessID() string {
	return cfg.OSSAccessID
}

// OSSAccessSecret 返回 oss access secret
func OSSAccessSecret() string {
	return cfg.OSSAccessSecret
}

// OSSBucket 返回 oss bucket.
func OSSBucket() string {
	return cfg.OSSBucket
}

// OSSPathPrefix 返回 文件在指定 bucket 下的路径前缀.
func OSSPathPrefix() string {
	return cfg.OSSPathPrefix
}

// CentralNexusPublicURL 返回 中心集群 nexus 公网地址
func CentralNexusPublicURL() string {
	return cfg.CentralNexusPublicURL
}

// CentralNexusAddr 返回 中心集群 nexus 地址
func CentralNexusAddr() string {
	return cfg.CentralNexusAddr
}

// CentralNexusComponentName 返回 中心集群 nexus 组件名
func CentralNexusComponentName() string {
	return strings.Split(httpclientutil.RmProto(cfg.CentralNexusAddr), ".")[0]
}

// CentralNexusUsername 返回 中心集群 nexus 用户名
func CentralNexusUsername() string {
	return cfg.CentralNexusUsername
}

// CentralNexusPassword 返回 中心集群 nexus 密码
func CentralNexusPassword() string {
	return cfg.CentralNexusPassword
}

// DiceClusterName 返回 cmdb 组件运行的中心集群名
func DiceClusterName() string {
	return cfg.DiceClusterName
}

// Base64EncodedRsaPublicKey 返回 rsa 公钥
func Base64EncodedRsaPublicKey() string {
	return cfg.Base64EncodedRsaPublicKey
}

// Base64EncodedRsaPrivateKey 返回 rsa 私钥
func Base64EncodedRsaPrivateKey() string {
	return cfg.Base64EncodedRsaPrivateKey
}

// EnableNS 是否打开项目级命名空间
func EnableNS() bool {
	return cfg.EnableProjectNS
}

func QAAddr() string {
	return cfg.QAAddr
}

func OryEnabled() bool {
	return cfg.OryEnabled
}

func OryKratosAddr() string {
	return cfg.OryKratosAddr
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

func MetricsIssueCron() string {
	return cfg.MetricsIssueCron
}

func CreateOrgEnabled() bool {
	return cfg.CreateOrgEnabled
}
