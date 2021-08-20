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

// Package conf Define the configuration
package conf

import (
	"time"

	"github.com/erda-project/erda/pkg/envconf"
)

// Conf Define the configuration
type Conf struct {
	Debug              bool          `env:"DEBUG" default:"false"`
	EnableEss          bool          `env:"ENABLE_ESS" default:"false"`
	ListenAddr         string        `env:"LISTEN_ADDR" default:":9027"`
	SoldierAddr        string        `env:"SOLDIER_ADDR"`
	SchedulerAddr      string        `env:"SCHEDULER_ADDR"`
	TaskSyncDuration   time.Duration `env:"TASK_SYNC_DURATION" default:"2h"`
	TaskCleanDuration  time.Duration `env:"TASK_CLEAN_DURATION" default:"24h"`
	LocalMode          bool          `env:"LOCAL_MODE" default:"false"`
	RedisMasterName    string        `default:"my-master" env:"REDIS_MASTER_NAME"`
	RedisSentinelAddrs string        `default:"" env:"REDIS_SENTINELS_ADDR"`
	RedisAddr          string        `default:"127.0.0.1:6379" env:"REDIS_ADDR"`
	RedisPwd           string        `default:"anywhere" env:"REDIS_PASSWORD"`
	UCClientID         string        `env:"UC_CLIENT_ID"`
	UCClientSecret     string        `env:"UC_CLIENT_SECRET"`
	// ory/kratos config
	OryEnabled           bool   `default:"false" env:"ORY_ENABLED"`
	OryKratosAddr        string `default:"kratos:4433" env:"KRATOS_ADDR"`
	OryKratosPrivateAddr string `default:"kratos:4434" env:"KRATOS_PRIVATE_ADDR"`

	ReleaseRegistry    string `env:"RELEASE_REGISTRY" default:"registry.erda.cloud/erda"`
	ClusterInitVersion string `env:"CLUSTER_INIT_VERSION" default:"0.1"`

	HelmChartRepoURL      string `env:"HELM_REPO_URL"`
	HelmChartRepoUserName string `env:"HELM_REPO_USERNAME"`
	HelmChartRepoPassword string `env:"HELM_REPO_PASSWORD"`

	ErdaNamespace   string `env:"DICE_NAMESPACE" default:"erda-system"`
	ErdaVersion     string `env:"DICE_VERSION"`
	ErdaProtocol    string `env:"DICE_PROTOCOL"`
	ErdaClusterName string `env:"DICE_CLUSTER_NAME"`
	ErdaDomain      string `env:"DICE_ROOT_DOMAIN"`
}

var cfg Conf

// Load Load envs
func Load() {
	envconf.MustLoad(&cfg)
}

// ListenAddr return the address of listen.
func ListenAddr() string {
	return cfg.ListenAddr
}

// Debug Return the switch of debug.
func Debug() bool {
	return cfg.Debug
}

func EnableEss() bool {
	return cfg.EnableEss
}

func TaskSyncDuration() time.Duration {
	return cfg.TaskSyncDuration
}

func TaskCleanDuration() time.Duration {
	return cfg.TaskCleanDuration
}

func LocalMode() bool {
	return cfg.LocalMode
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

// UCClientID 返回 UCClientID 选项.
func UCClientID() string {
	return cfg.UCClientID
}

// UCClientSecret 返回 UCClientSecret 选项.
func UCClientSecret() string {
	return cfg.UCClientSecret
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

func ReleaseRegistry() string {
	return cfg.ReleaseRegistry
}

func ClusterInitVersion() string {
	return cfg.ClusterInitVersion
}

func HelmRepoURL() string {
	return cfg.HelmChartRepoURL
}

func HelmRepoUsername() string {
	return cfg.HelmChartRepoUserName
}

func HelmRepoPassword() string {
	return cfg.HelmChartRepoPassword
}

func ErdaNamespace() string {
	return cfg.ErdaNamespace
}

func ErdaVersion() string {
	return cfg.ErdaVersion
}

func ErdaProtocol() string {
	return cfg.ErdaProtocol
}

func ErdaClusterName() string {
	return cfg.ErdaClusterName
}

func ErdaDomain() string {
	return cfg.ErdaDomain
}
