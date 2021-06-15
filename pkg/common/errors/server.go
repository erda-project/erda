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

// InternalServerError .
type InternalServerError struct {
	Cause error
}

// NewInternalServerError .
func NewInternalServerError(err error) *InternalServerError {
	return &InternalServerError{Cause: err}
}

func (e *InternalServerError) Error() string {
	return fmt.Sprintf("internal error: %s", e.Cause)
}
func (e *InternalServerError) HTTPStatus() int { return http.StatusInternalServerError }
func (e *InternalServerError) Translate(t i18n.Translator, langs i18n.LanguageCodes) string {
	return t.Sprintf(langs, "${internal} ${error}: %s", e.Cause)
}

// DataBaseError .
type DataBaseError struct {
	Cause error
}

// NewDataBaseError .
func NewDataBaseError(err error) *DataBaseError {
	return &DataBaseError{Cause: err}
}

func (e *DataBaseError) Error() string {
	return fmt.Sprintf("database error: %s", e.Cause)
}
func (e *DataBaseError) HTTPStatus() int { return http.StatusInternalServerError }
func (e *DataBaseError) Translate(t i18n.Translator, langs i18n.LanguageCodes) string {
	return t.Sprintf(langs, "${database} ${error}: %s", e.Cause)
}

// ServiceInvokingError .
type ServiceInvokingError struct {
	Service string
	Cause   error
}

// NewServiceInvokingError .
func NewServiceInvokingError(service string, err error) *ServiceInvokingError {
	return &ServiceInvokingError{
		Service: service,
		Cause:   err,
	}
}

func (e *ServiceInvokingError) Error() string {
	return fmt.Sprintf("service %s error: %s", e.Service, e.Cause)
}
func (e *ServiceInvokingError) HTTPStatus() int { return http.StatusBadGateway }
func (e *ServiceInvokingError) Translate(t i18n.Translator, langs i18n.LanguageCodes) string {
	return t.Sprintf(langs, "${service} % ${error}: %s", t.Text(langs, e.Service), e.Cause)
}

// UnimplementedError .
type UnimplementedError struct {
	Service string
}

// NewUnimplementedError .
func NewUnimplementedError(service string) *UnimplementedError {
	return &UnimplementedError{Service: service}
}

func (e *UnimplementedError) Error() string {
	return fmt.Sprintf("service %s not implemented", e.Service)
}
func (e *UnimplementedError) HTTPStatus() int { return http.StatusNotImplemented }
func (e *UnimplementedError) Translate(t i18n.Translator, langs i18n.LanguageCodes) string {
	return t.Sprintf(langs, "${service} %s ${not implemented}", t.Text(langs, e.Service))
}
