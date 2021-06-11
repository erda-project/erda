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

package errors

import (
	"fmt"
	"net/http"

	"github.com/erda-project/erda-infra/providers/i18n"
)

// NotFoundError .
type NotFoundError struct {
	Resource string
}

var _ Error = (*NotFoundError)(nil)

// NewNotFoundError .
func NewNotFoundError(resource string) *NotFoundError {
	return &NotFoundError{Resource: resource}
}

func (e *NotFoundError) Error() string {
	if len(e.Resource) > 0 {
		return fmt.Sprintf("%s not found", e.Resource)
	}
	return "not found"
}
func (e *NotFoundError) HTTPStatus() int { return http.StatusNotFound }
func (e *NotFoundError) Translate(t i18n.Translator, langs i18n.LanguageCodes) string {
	if len(e.Resource) > 0 {
		return t.Sprintf(langs, "%s ${not found}", t.Text(langs, e.Resource))
	}
	return t.Text(langs, "not found")
}

// AlreadyExistsError .
type AlreadyExistsError struct {
	Resource string
}

var _ Error = (*AlreadyExistsError)(nil)

// NewAlreadyExistsError .
func NewAlreadyExistsError(resource string) *AlreadyExistsError {
	return &AlreadyExistsError{Resource: resource}
}

func (e *AlreadyExistsError) Error() string {
	if len(e.Resource) > 0 {
		return fmt.Sprintf("%s already exists", e.Resource)
	}
	return "already exists"
}
func (e *AlreadyExistsError) HTTPStatus() int { return http.StatusNotFound }
func (e *AlreadyExistsError) Translate(t i18n.Translator, langs i18n.LanguageCodes) string {
	if len(e.Resource) > 0 {
		return t.Sprintf(langs, "%s ${already exists}", t.Text(langs, e.Resource))
	}
	return t.Text(langs, "already exists")
}
