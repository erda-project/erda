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

// UnauthorizedError .
type UnauthorizedError struct {
	Reason string
}

var _ Error = (*UnauthorizedError)(nil)

func NewUnauthorizedError(reason string) *UnauthorizedError {
	return &UnauthorizedError{Reason: reason}
}

func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("unauthorized%s", suffixIfNotEmpty(":", e.Reason))
}
func (e *UnauthorizedError) HTTPStatus() int { return http.StatusUnauthorized }
func (e *UnauthorizedError) Translate(t i18n.Translator, langs i18n.LanguageCodes) string {
	return t.Sprintf(langs, "unauthorized%s", suffixIfNotEmpty(":", e.Reason))
}

// PermissionError .
type PermissionError struct {
	Resource string
	Action   string
	Reason   string
}

var _ Error = (*PermissionError)(nil)

// NewPermissionError .
func NewPermissionError(resource, action, reason string) *PermissionError {
	return &PermissionError{
		Resource: resource,
		Action:   action,
		Reason:   reason,
	}
}

func (e *PermissionError) Error() string {
	if len(e.Resource) > 0 {
		if len(e.Action) > 0 {
			return fmt.Sprintf("permission denied to %s %s%s", e.Action, e.Resource, suffixIfNotEmpty(":", e.Reason))
		}
		return fmt.Sprintf("permission denied to access %s%s", e.Resource, suffixIfNotEmpty(":", e.Reason))
	} else if len(e.Action) > 0 {
		return fmt.Sprintf("permission denied to %s%s", e.Action, suffixIfNotEmpty(":", e.Reason))
	}
	return "permission denied"
}
func (e *PermissionError) HTTPStatus() int { return http.StatusForbidden }
func (e *PermissionError) Translate(t i18n.Translator, langs i18n.LanguageCodes) string {
	if len(e.Resource) > 0 {
		if len(e.Action) > 0 {
			return t.Sprintf(langs, "${permission denied}: %s %s%s", t.Text(langs, e.Action), t.Text(langs, e.Resource), suffixIfNotEmpty(",", e.Reason))
		}
		return t.Sprintf(langs, "${permission denied}: %s%s", t.Text(langs, e.Resource), suffixIfNotEmpty(",", e.Reason))
	} else if len(e.Action) > 0 {
		return t.Sprintf(langs, "${permission denied}: %s%s", t.Text(langs, e.Action), suffixIfNotEmpty(",", e.Reason))
	}
	return t.Sprintf(langs, "permission denied%s", suffixIfNotEmpty(":", e.Reason))
}
