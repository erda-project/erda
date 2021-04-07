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

// Store the key information storage interface
type Store interface {
	// PluginKind is key store type
	GetKind() StoreKind

	// Create create and store new CMK
	CreateKey(info KeyInfo) error

	// GetKey use keyID to find CMK
	GetKey(keyID string) (KeyInfo, error)

	// ListByKind use plugin type to list CMKs
	ListKeysByKind(kind PluginKind) ([]string, error)

	// DeleteByKeyID use keyID to delete CMK
	DeleteByKeyID(keyID string) error

	// GetKeyVersion use keyID and keyVersionID to find keyVersion
	GetKeyVersion(keyID, keyVersionID string) (KeyVersionInfo, error)

	// RotateKeyVersion rotate key version
	RotateKeyVersion(keyID string, newKeyVersionInfo KeyVersionInfo) (KeyVersionInfo, error)
}
