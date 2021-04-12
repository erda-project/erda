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

// Package errcode Dicehub错误码
package errcode

// ImageErrorCode 错误码类型
type ImageErrorCode string

// Dicehub错误码
const (
	NetworkError        ImageErrorCode = "IMG0001"
	InternalServerError ImageErrorCode = "IMG0002"
	PermissionDenied    ImageErrorCode = "IMG0003"
	ResourceNotFound    ImageErrorCode = "IMG0004"
	HeaderMissing       ImageErrorCode = "IMG0005"
	BodyMissing         ImageErrorCode = "IMG0006"
	BodyInvalidFormat   ImageErrorCode = "IMG0007"
	ParamMissing        ImageErrorCode = "IMG0008"
	ParamInvalidFormat  ImageErrorCode = "IMG0009"
	ParamInvalid        ImageErrorCode = "IMG0010"
	ParamExist          ImageErrorCode = "IMG0011"
	ResourceInUse       ImageErrorCode = "IMG0012"
)
