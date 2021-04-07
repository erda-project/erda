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

package apistructs

import (
	"github.com/erda-project/erda/pkg/kms/kmstypes"
)

// create key
type KMSCreateKeyRequest struct {
	kmstypes.CreateKeyRequest
}
type KMSCreateKeyResponse struct {
	Header
	Data *kmstypes.CreateKeyResponse `json:"data,omitempty"`
}

// encrypt
type KMSEncryptRequest struct {
	kmstypes.EncryptRequest
}
type KMSEncryptResponse struct {
	Header
	Data *kmstypes.EncryptResponse `json:"data,omitempty"`
}

// decrypt
type KMSDecryptRequest struct {
	kmstypes.DecryptRequest
}
type KMSDecryptResponse struct {
	Header
	Data *kmstypes.DecryptResponse `json:"data,omitempty"`
}

// generate data key
type KMSGenerateDataKeyRequest struct {
	kmstypes.GenerateDataKeyRequest
}
type KMSGenerateDataKeyResponse struct {
	Header
	Data *kmstypes.GenerateDataKeyResponse `json:"data,omitempty"`
}

// rotate key version
type KMSRotateKeyVersionRequest struct {
	kmstypes.RotateKeyVersionRequest
}
type KMSRotateKeyVersionResponse struct {
	Header
	Data *kmstypes.RotateKeyVersionResponse `json:"data,omitempty"`
}

// describe key
type KMSDescribeKeyRequest struct {
	kmstypes.DescribeKeyRequest
}
type KMSDescribeKeyResponse struct {
	Header
	Data *kmstypes.DescribeKeyResponse `json:"data,omitempty"`
}
