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

package api

import "errors"

type Map map[string]interface{}

var (
	ERROR_CODE_INTERNAL     = "500"
	ERROR_CODE_NOT_FOUND    = "404"
	ERROR_CODE_INVALID_ARGS = "400"
)

var (
	ERROR_NOT_FILE       = errors.New("path not file")
	ERROR_PATH_NOT_FOUND = errors.New("path not exist")
	ERROR_DB             = errors.New("db error")
	ERROR_ARG_ID         = errors.New("id parse failed")
	ERROR_HOOK_NOT_FOUND = errors.New("hook not found")
	ERROR_LOCKED_DENIED  = errors.New("locked denied")
	ERROR_REPO_LOCKED    = errors.New("repo locked")
)
