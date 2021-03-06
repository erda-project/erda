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

package spec

import "github.com/erda-project/erda/pkg/parser/diceyml"

type LegacyDice struct {
	Name         string                 `json:"name"`
	Endpoints    map[string]*Service    `json:"endpoints,omitempty"`
	Services     map[string]*Service    `json:"services,omitempty"`
	Addons       map[string]*Addon      `json:"addons,omitempty"`
	AddonActions map[string]interface{} `json:"addonActions,omitempty"`
	Branch       string                 `json:"branch"`
	GlobalEnv    map[string]string      `json:"globalEnv,omitempty"`
}

type Service struct {
	Scale       *int                        `json:"scale,omitempty"`
	Ports       []int                       `json:"ports,omitempty"`
	Depends     []string                    `json:"depends,omitempty"`
	Environment map[string]string           `json:"environment,omitempty"`
	Resources   *Resources                  `json:"resources,omitempty"`
	Cmd         string                      `json:"cmd"`
	Hosts       []string                    `json:"hosts"`
	Volumes     []string                    `json:"volumes"`
	Image       string                      `json:"image"`
	HealthCheck HealthCheck                 `json:"health_check"`
	SideCars    map[string]*diceyml.SideCar `yaml:"sidecars,omitempty" json:"sidecars,omitempty"`
}

type HealthCheck struct {
	HTTP *HTTPCheck `json:"http,omitempty"`
	Exec *ExecCheck `json:"exec,omitempty"`
}

type HTTPCheck struct {
	Port     int    `json:"port,string"`
	Path     string `json:"path"`
	Duration int    `json:"duration,string"`
}

type ExecCheck struct {
	Cmd      string `json:"cmd"`
	Duration int    `json:"duration,string"`
}

type Resources struct {
	CPU  float64 `json:"cpu"`
	Mem  float64 `json:"mem"`
	Disk float64 `json:"disk"`
}

type Addon struct {
	Id      string                 `json:"id"`
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Plan    string                 `json:"plan"`
	As      string                 `json:"as"`
	Options map[string]string      `json:"options"`
	Actions map[string]interface{} `json:"actions"`
}
