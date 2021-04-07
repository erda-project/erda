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

// IdentityInfo represents operator identity info.
// Fields will not be json marshal/unmarshal.
type IdentityInfo struct {
	// UserID is user id. It must be provided in some cases.
	// Cannot be null if InternalClient is null.
	// +optional
	UserID string `json:"userID"`

	// InternalClient records the internal client, such as: bundle.
	// Cannot be null if UserID is null.
	// +optional
	InternalClient string `json:"-"`
}

func (info *IdentityInfo) IsInternalClient() bool {
	return info.InternalClient != ""
}

func (info *IdentityInfo) Empty() bool {
	return info.UserID == "" && info.InternalClient == ""
}
