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

package pbutil

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
)

func GetBool(v *bool) (vv bool, set bool) {
	if v == nil {
		return false, false
	}
	return *v, true
}
func MustGetBool(v *bool) bool {
	if vv, set := GetBool(v); set {
		return vv
	}
	return false
}

func GetUint64(v *uint64) (vv uint64, set bool) {
	if v == nil {
		return 0, false
	}
	return *v, true
}
func MustGetUint64(v *uint64) uint64 {
	if vv, set := GetUint64(v); set {
		return vv
	}
	return 0
}

func GetTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
func GetTime(timestamp *timestamppb.Timestamp) *time.Time {
	if timestamp == nil {
		return nil
	}
	t := timestamp.AsTime()
	return &t
}
func MustGetTime(timestamp *timestamppb.Timestamp) time.Time {
	t := GetTime(timestamp)
	if t != nil {
		return *t
	}
	return time.Time{}
}

func GetIdentityUser(identityInfo *commonpb.IdentityInfo) string {
	if identityInfo == nil {
		return ""
	}
	return identityInfo.UserID
}
func GetIdentityInternalClient(identityInfo *commonpb.IdentityInfo) string {
	if identityInfo == nil {
		return ""
	}
	return identityInfo.InternalClient
}
