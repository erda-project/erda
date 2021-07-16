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

package httputil

// dice 公共 Header
const (
	UserHeader           = "User-ID"
	OrgHeader            = "Org-ID"
	InternalHeader       = "Internal-Client"        // 内部服务间调用时使用
	InternalActionHeader = "Internal-Action-Client" // action calls the api header
	RequestIDHeader      = "RequestID"

	ClientIDHeader   = "Client-ID"
	ClientNameHeader = "Client-Name"
)
