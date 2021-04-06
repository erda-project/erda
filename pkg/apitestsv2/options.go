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

package apitestsv2

type option struct {
	tryV1RenderJsonBodyFirst bool
}

type OpOption func(*option)

// WithTryV1RenderJsonBodyFirst 尝试先使用 v1 严格模式渲染 json body。不论是否打开开关，都会再使用 v2 逻辑渲染一遍。
// 为手动测试的接口测试提供兼容处理；自动化测试无需打开该开关。
func WithTryV1RenderJsonBodyFirst() OpOption {
	return func(opt *option) {
		opt.tryV1RenderJsonBodyFirst = true
	}
}
