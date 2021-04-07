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

package kmstypes

type RequestValidator interface {
	ValidateRequest() error
}

const (
	CtxKeyKmsRequestID = "KmsRequestID"
)

const (
	PluginKind_DICE_KMS   PluginKind = "DICE_KMS"
	PluginKind_AWS_KMS    PluginKind = "AWS_KMS"
	PluginKind_ALIYUN_KMS PluginKind = "ALIYUN_KMS"

	StoreKind_ETCD  StoreKind = "ETCD"
	StoreKind_MYSQL StoreKind = "MYSQL"

	CustomerMasterKeySpec_SYMMETRIC_DEFAULT   CustomerMasterKeySpec = "SYMMETRIC_DEFAULT" // AES-256-GCM ; default
	CustomerMasterKeySpec_ASYMMETRIC_RSA_2048 CustomerMasterKeySpec = "RSA_2048"
	CustomerMasterKeySpec_ASYMMETRIC_RSA_3072 CustomerMasterKeySpec = "RSA_3072"
	CustomerMasterKeySpec_ASYMMETRIC_RSA_4096 CustomerMasterKeySpec = "RSA_4096"

	KeyUsage_ENCRYPT_DECRYPT KeyUsage = "ENCRYPT_DECRYPT"
	KeyUsage_SIGN_VERIFY     KeyUsage = "SIGN_VERIFY"

	KeyStateEnabled         KeyState = "Enabled"
	KeyStateDisabled        KeyState = "Disabled"
	KeyStatePendingDeletion KeyState = "PendingDeletion"
	KeyStatePendingImport   KeyState = "PendingImport"
	KeyStateUnavailable     KeyState = "Unavailable"
)
