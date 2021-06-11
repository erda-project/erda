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

// InvalidParameterError .
type InvalidParameterError struct {
	Name    string
	Message string
}

var _ Error = (*InvalidParameterError)(nil)

// NewInvalidParameterError .
func NewInvalidParameterError(name, message string) *InvalidParameterError {
	return &InvalidParameterError{
		Name:    name,
		Message: message,
	}
}

func (e *InvalidParameterError) Error() string {
	if len(e.Message) > 0 {
		return fmt.Sprintf("parameter %s invalid: %s", e.Name, e.Message)
	}
	return fmt.Sprintf("parameter %s invalid", e.Name)
}

func (e *InvalidParameterError) HTTPStatus() int { return http.StatusBadRequest }
func (e *InvalidParameterError) Translate(t i18n.Translator, langs i18n.LanguageCodes) string {
	if len(e.Message) > 0 {
		return t.Sprintf(langs, "${parameter} %s ${invalid}: %s", t.Text(langs, e.Name), t.Text(langs, e.Message))
	}
	return t.Sprintf(langs, "${parameter} %s ${invalid}", t.Text(langs, e.Name))
}

// MissingParameterError .
type MissingParameterError struct {
	Name string
}

var _ Error = (*MissingParameterError)(nil)

func NewMissingParameterError(name string) *MissingParameterError {
	return &MissingParameterError{Name: name}
}

func (e *MissingParameterError) Error() string {
	return fmt.Sprintf("parameter %s missing", e.Name)
}
func (e *MissingParameterError) HTTPStatus() int { return http.StatusBadRequest }
func (e *MissingParameterError) Translate(t i18n.Translator, langs i18n.LanguageCodes) string {
	return t.Sprintf(langs, "${parameter} %s ${missing}", t.Text(langs, e.Name))
}

// ParameterTypeError
type ParameterTypeError struct {
	Name      string
	ValidType string
}

var _ Error = (*ParameterTypeError)(nil)

// NewParameterTypeError .
func NewParameterTypeError(name string) *ParameterTypeError {
	return &ParameterTypeError{Name: name}
}

func (e *ParameterTypeError) Error() string {
	if len(e.ValidType) > 0 {
		return fmt.Sprintf("parameter %s want %s type", e.Name, e.ValidType)
	}
	return fmt.Sprintf("parameter %s type error", e.Name)
}
func (e *ParameterTypeError) HTTPStatus() int { return http.StatusBadRequest }
func (e *ParameterTypeError) Translate(t i18n.Translator, langs i18n.LanguageCodes) string {
	if len(e.ValidType) > 0 {
		return t.Sprintf(langs, "${parameter} %s ${want} %s ${type}", t.Text(langs, e.Name), t.Text(langs, e.ValidType))
	}
	return t.Sprintf(langs, "${parameter} %s ${type error}", t.Text(langs, e.Name))
}
