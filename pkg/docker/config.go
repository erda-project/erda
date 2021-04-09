// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package docker

import (
	"crypto/tls"
	"time"
)

type Config struct {
	Host      string `config:"host"`
	TLSConfig *tls.Config

	RequestTimeout time.Duration `config:"request_timeout"` // docker请求超时时间
	EventTimeout   time.Duration `config:"event_timeout"`   // docker event请求超时时间
	WatchTimeout   time.Duration `config:"watch_timeout"`   // 监听间隔时间
	WatchInterval  time.Duration `config:"watch_interval"`  // 监听超时时间
	CleanupTimeout time.Duration `config:"cleanup_timeout"` // 清理死亡容器的超时时间
}
