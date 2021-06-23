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

package dao

import (
	"github.com/pkg/errors"
)

// dao层错误码统一定义
var (
	ErrNotFoundOrg         = errors.New("org not found")
	ErrNotFoundProject     = errors.New("project not found")
	ErrNotFoundApplication = errors.New("application not found")
	ErrNotFoundMember      = errors.New("member not found")
	ErrNotFoundTicket      = errors.New("ticket not found")
	ErrNotFoundPublisher   = errors.New("publisher not found")
	ErrNotFoundCertificate = errors.New("certificate not found")
	ErrNotFoundApprove     = errors.New("approve not found")
	ErrNotFoundUsecase     = errors.New("usecase not found")
)
