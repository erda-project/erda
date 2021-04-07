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

// Package bundle 定义了整个 dice 的 client 方法
// 用法：
// b := bundle.New()
// b.CallXXX()
package bundle

import (
	"os"
	"time"

	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/i18n"
)

// Bundle 定义了所有方法的集合对象.
type Bundle struct {
	hc         *httpclient.HTTPClient
	i18nLoader *i18n.LocaleResourceLoader
	urls       urls
}

// Option 定义 Bundle 对象的配置选项.
type Option func(*Bundle)

// New 创建一个新的 Bundle 实例对象，通过 Bundle 对象可以直接调用所有方法.
func New(options ...Option) *Bundle {
	b := &Bundle{
		urls: make(map[string]string),
	}
	for _, op := range options {
		op(b)
	}
	if b.hc == nil {
		b.hc = httpclient.New(
			httpclient.WithTimeout(time.Second, time.Second*3),
		)
	}
	if b.i18nLoader == nil {
		b.i18nLoader = i18n.NewLoader()
		b.i18nLoader.LoadDir("dice-configs/i18n")
		b.i18nLoader.DefaultLocale("zh-CN")
	}
	return b
}

// WithHTTPClient 配置 http 客户端对象.
func WithHTTPClient(hc *httpclient.HTTPClient) Option {
	return func(b *Bundle) {
		b.hc = hc
	}
}

// WithI18nLoader 配置 i18n对象
func WithI18nLoader(i18nLoader *i18n.LocaleResourceLoader) Option {
	return func(b *Bundle) {
		b.i18nLoader = i18nLoader
	}
}

// WithEventBox 根据环境变量配置创建 eventbox 客户端.
func WithEventBox() Option {
	return func(b *Bundle) {
		k := discover.EnvEventBox
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithScheduler 根据环境变量配置创建 scheduler 客户端.
func WithScheduler() Option {
	return func(b *Bundle) {
		k := discover.EnvScheduler
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithCMDB 根据环境变量配置创建 cmdb 客户端.
func WithCMDB() Option {
	return func(b *Bundle) {
		k := discover.EnvCMDB
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithDiceHub 根据环境变量配置创建 dicehub 客户端.
func WithDiceHub() Option {
	return func(b *Bundle) {
		k := discover.EnvDiceHub
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithOrchestrator 根据环境变量配置创建 orachestrator 客户端.
func WithOrchestrator() Option {
	return func(b *Bundle) {
		k := discover.EnvOrchestrator
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithOps 根据环境变量配置创建 ops 客户端.
func WithOps() Option {
	return func(b *Bundle) {
		k := discover.EnvOps
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithOpenapi 根据环境变量配置创建 openapi 客户端.
func WithOpenapi() Option {
	return func(b *Bundle) {
		k := discover.EnvOpenapi
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithQA 根据环境变量配置创建 QA 客户端.
func WithQA() Option {
	return func(b *Bundle) {
		k := discover.EnvQA
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithSoldier 根据环境变量配置创建 soldier 客户端.
// 支持直接设置 addr，优先级高于环境变量.
func WithSoldier(addr ...string) Option {
	return func(b *Bundle) {
		k := discover.EnvSoldier
		if len(addr) > 0 {
			b.urls.Put(k, addr[0])
			return
		}
		b.urls.Put(k, os.Getenv(k))
	}
}

// WithAddOnPlatform 根据环境变量配置创建 addOnPlatform 客户端.
func WithAddOnPlatform() Option {
	return func(b *Bundle) {
		k := discover.EnvAddOnPlatform
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithGittar 根据环境变量创建 gittar 客户端
func WithGittar() Option {
	return func(b *Bundle) {
		k := discover.EnvGittar
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithGittarAdaptor 根据环境变量创建 gittar-adaptor 客户端
func WithGittarAdaptor() Option {
	return func(b *Bundle) {
		k := discover.EnvGittarAdaptor
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithCollector 根据环境变量创建 collector 客户端
func WithCollector() Option {
	return func(b *Bundle) {
		k := discover.EnvCollector
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithMonitor 根据环境变量创建 monitor 客户端
func WithMonitor() Option {
	return func(b *Bundle) {
		k := discover.EnvMonitor
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithTMC 根据环境变量创建 tmc 客户端
func WithTMC() Option {
	return func(b *Bundle) {
		k := discover.EnvTMC
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

func WithPipeline() Option {
	return func(b *Bundle) {
		k := discover.EnvPipeline
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithHepa 根据环境变量创建 hepa 客户端
func WithHepa() Option {
	return func(b *Bundle) {
		k := discover.EnvHepa
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

func WithKMS() Option {
	return func(b *Bundle) {
		k := discover.EnvKMS
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

func WithAPIM() Option {
	return func(b *Bundle) {
		k := discover.EnvAPIM
		v := os.Getenv(k)
		b.urls.Put(k, v)
	}
}

// WithAllAvailableClients 将环境变量中所有可以拿到的客户端均注入
func WithAllAvailableClients() Option {
	return func(b *Bundle) {
		b.urls.PutAllAvailable()
	}
}
