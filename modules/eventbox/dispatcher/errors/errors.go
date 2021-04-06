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
	"encoding/json"
	"errors"
	"fmt"
)

var (
	DispatcherBusyErr error = errors.New("dispatcher busy")
)

type BackendErrs map[string][]error

func (b BackendErrs) MarshalJSON() ([]byte, error) {
	backendErrs := make(map[string][]string)
	for k, v := range b {
		for _, e := range v {
			backendErrs[k] = append(backendErrs[k], e.Error())
		}
	}
	return json.Marshal(backendErrs)
}

type DispatchError struct {
	BackendErrs BackendErrs
	FilterInfo  string
	FilterErr   error
}

func New() *DispatchError {
	return &DispatchError{
		BackendErrs: make(map[string][]error),
	}
}

func (d *DispatchError) String() string {
	r := "dispatcherErr: "
	for k, v := range d.BackendErrs {
		r += fmt.Sprintf("[%s:%v] ", k, v)
	}
	r += fmt.Sprintf("FilterInfo: %s, ", d.FilterInfo)
	r += fmt.Sprintf("FilterErr: %v", d.FilterErr)
	return r
}

func (d *DispatchError) IsOK() bool {
	if len(d.BackendErrs) > 0 {
		return false
	}
	if d.FilterErr != nil {
		return false
	}
	return true
}

func (d *DispatchError) IsFiltered() bool {
	return (d.FilterErr != nil)
}

func (d DispatchError) MarshalJSON() ([]byte, error) {
	filterErrStr := ""
	if d.FilterErr != nil {
		filterErrStr = d.FilterErr.Error()
	}
	s := struct {
		BackendErrs BackendErrs
		FilterInfo  string
		FilterErr   string
	}{
		BackendErrs: d.BackendErrs,
		FilterInfo:  d.FilterInfo,
		FilterErr:   filterErrStr,
	}

	return json.Marshal(&s)
}
