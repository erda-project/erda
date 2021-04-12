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

type AsymmetricDecryptRequest struct {
	KeyID            string `json:"keyID,omitempty"`
	CiphertextBase64 []byte `json:"ciphertextBase64,omitempty"`
}
type AsymmetricDecryptResponse struct {
	PlaintextBase64 []byte `json:"plaintextBase64,omitempty"`
}

type GetPublicKeyRequest struct {
	KeyID string `json:"keyID,omitempty"`
}

type PublicKey struct {
	Pem       string `json:"pem,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
}
