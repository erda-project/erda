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

import "fmt"

type (
	CustomerMasterKeySpec string
	KeyUsage              string
	KeyState              string
)

type (
	KeyMetadata struct {
		KeyID                 string                `json:"keyID,omitempty"`
		PrimaryKeyVersionID   string                `json:"primaryKeyVersionID,omitempty"`
		CustomerMasterKeySpec CustomerMasterKeySpec `json:"customerMasterKeySpec,omitempty"`
		KeyUsage              KeyUsage              `json:"keyUsage,omitempty"`
		KeyState              KeyState              `json:"keyState,omitempty"`
		Description           string                `json:"description,omitempty"`
	}

	KeyListEntry struct {
		KeyID string `json:"keyID,omitempty"`
	}
)

type CreateKeyRequest struct {
	PluginKind            PluginKind            `json:"pluginKind,omitempty"`
	CustomerMasterKeySpec CustomerMasterKeySpec `json:"customerMasterKeySpec,omitempty"`
	KeyUsage              KeyUsage              `json:"keyUsage,omitempty"`
	Description           string                `json:"description,omitempty"`
}

func (req *CreateKeyRequest) ValidateRequest() error {
	if req.PluginKind == "" {
		req.PluginKind = PluginKind_DICE_KMS
	}
	if req.CustomerMasterKeySpec == "" {
		req.CustomerMasterKeySpec = CustomerMasterKeySpec_SYMMETRIC_DEFAULT
	}
	if req.KeyUsage == "" {
		req.KeyUsage = KeyUsage_ENCRYPT_DECRYPT
	}
	return nil
}

type CreateKeyResponse struct {
	KeyMetadata KeyMetadata `json:"keyMetadata,omitempty"`
}

type GenerateDataKeyRequest struct {
	KeyID string `json:"keyID,omitempty"`
}

func (req *GenerateDataKeyRequest) ValidateRequest() error {
	if req.KeyID == "" {
		return fmt.Errorf("missing keyID")
	}
	return nil
}

type GenerateDataKeyResponse struct {
	KeyID            string `json:"keyID,omitempty"`
	KeyVersionID     string `json:"keyVersionID,omitempty"`
	CiphertextBase64 string `json:"ciphertextBase64,omitempty"`
	PlaintextBase64  string `json:"plaintextBase64,omitempty"`
}

type DescribeKeyRequest struct {
	KeyID string `json:"keyID,omitempty"`
}

func (req *DescribeKeyRequest) ValidateRequest() error {
	if req.KeyID == "" {
		return fmt.Errorf("missing keyID")
	}
	return nil
}

type DescribeKeyResponse struct {
	KeyMetadata KeyMetadata `json:"keyMetadata,omitempty"`
}

type ListKeysRequest struct {
}

type ListKeysResponse struct {
	Keys []KeyListEntry `json:"keys,omitempty"`
}
