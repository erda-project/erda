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

// Package k8serror encapsulates error of k8s object
package k8serror

import "errors"

var (
	// ErrNotFound indicates "not found" error
	ErrNotFound = errors.New("not found")
	// ErrInvalidParams indicates invalid param(s)
	ErrInvalidParams = errors.New("invalid params")
)

// NotFound return whether it is a "not found" error
func NotFound(err error) bool {
	return notFound(err)
}

func notFound(err error) bool {
	return err == ErrNotFound
}
