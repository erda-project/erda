// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"fmt"
	"strconv"

	"k8s.io/utils/pointer"
)

const (
	Red    = "red"
	Yellow = "yellow"
	Green  = "green"
)

type MysqlSolo struct {
	//+optional
	Spec MysqlSoloSpec `json:"spec,omitempty"`

	//+optional
	Status MysqlSoloStatus `json:"status,omitempty"`
}

type MysqlSoloStatus struct {
	//+optional
	Color string `json:"color,omitempty"`

	//+optional
	Hang int `json:"hang,omitempty"`
}

type MysqlSoloSpec struct {
	//+optional
	Id int `json:"id,omitempty"`
	//+optional
	SourceId *int `json:"sourceId,omitempty"`

	//+optional
	Port int `json:"port,omitempty"`
	//+optional
	MyletPort int `json:"myletPort,omitempty"`

	//+optional
	GroupPort int `json:"groupPort,omitempty"`
	//+optional
	ExporterPort int `json:"exporterPort,omitempty"`

	//+optional
	Mydir string `json:"mydir,omitempty"`
	//+optional
	Host string `json:"host,omitempty"`

	//+optional
	Name string `json:"name,omitempty"`
	//+optional
	ServerId int `json:"serverId,omitempty"`
}

func (r *Mysql) SoloName(id int) string {
	return r.BuildName(strconv.Itoa(id))
}
func (r *Mysql) SoloHost(id int) string {
	return r.SoloName(id) + "." + r.Spec.HeadlessHost
}

func (r *Mysql) SoloShortHost(id int) string {
	if r.Spec.ShortHeadlessHost == "" {
		return r.Status.Solos[id].Spec.Host
	}
	return r.SoloName(id) + "." + r.Spec.ShortHeadlessHost
}

func (r *Mysql) NormalizeSolo(id int) (s MysqlSolo, err error) {
	n := r.Spec.Size()
	if !Between(id, 0, n-1) {
		return s, fmt.Errorf("mysql solo %d id out of range", id)
	}

	i := 0
	for _, v := range r.Spec.Solos {
		if v.Id == id {
			s.Spec = v
			i++
		}
	}
	if i > 1 {
		return s, fmt.Errorf("mysql solo %d duplicated", id)
	}

	s.Spec.Id = id
	s.Spec.ServerId = s.Spec.Id + serverId

	if s.Spec.SourceId == nil {
		if r.Spec.PrimaryMode == ModeClassic {
			id = *r.Spec.PrimaryId
			if s.Spec.Id > id {
				id = s.Spec.Id - 1
			} else if s.Spec.Id < id {
				id = s.Spec.Id + 1
			} else {
				id = -1
			}
		} else {
			id = -1
			if s.Spec.Id >= r.Spec.Primaries {
				id = s.Spec.Id - r.Spec.Primaries
			}
		}
		s.Spec.SourceId = pointer.IntPtr(id)
	} else {
		id = *s.Spec.SourceId
	}
	if id != -1 {
		if id == s.Spec.Id {
			return s, fmt.Errorf("mysql solo %d source id must not equal id", s.Spec.Id)
		}
		if !Between(id, 0, n-1) {
			return s, fmt.Errorf("mysql solo %d source id out of range", s.Spec.Id)
		}
	}

	if s.Spec.Port == 0 {
		s.Spec.Port = r.Spec.Port
	}
	if s.Spec.MyletPort == 0 {
		s.Spec.MyletPort = r.Spec.MyletPort
	}
	if s.Spec.GroupPort == 0 {
		s.Spec.GroupPort = r.Spec.GroupPort
	}
	if s.Spec.ExporterPort == 0 {
		s.Spec.ExporterPort = r.Spec.ExporterPort
	}
	if !Between(s.Spec.Port, minPort, maxPort) {
		return s, fmt.Errorf("mysql solo %d port not in [%d, %d]: %d", s.Spec.Id, minPort, maxPort, s.Spec.Port)
	}
	if !Between(s.Spec.MyletPort, minPort, maxPort) {
		return s, fmt.Errorf("mysql solo %d mylet port not in [%d, %d]: %d", s.Spec.Id, minPort, maxPort, s.Spec.MyletPort)
	}
	if !Between(s.Spec.GroupPort, minPort, maxPort) {
		return s, fmt.Errorf("mysql solo %d group port not in [%d, %d]: %d", s.Spec.Id, minPort, maxPort, s.Spec.GroupPort)
	}
	if !Between(s.Spec.ExporterPort, minPort, maxPort) {
		return s, fmt.Errorf("mysql solo %d exporter port not in [%d, %d]: %d", s.Spec.Id, minPort, maxPort, s.Spec.ExporterPort)
	}
	if HasEqual(s.Spec.Port, s.Spec.MyletPort, s.Spec.GroupPort, s.Spec.ExporterPort) {
		return s, fmt.Errorf("mysql solo %d ports must not equal", s.Spec.Id)
	}

	if s.Spec.Mydir == "" {
		s.Spec.Mydir = r.Spec.Mydir
	}
	if s.Spec.Name == "" {
		s.Spec.Name = r.SoloName(s.Spec.Id)
	}
	if s.Spec.Host == "" {
		s.Spec.Host = r.SoloHost(s.Spec.Id)
	}
	if host, port := SplitHostPort(s.Spec.Host); host == "" {
		return s, fmt.Errorf("mysql solo %d host invalid: %s", s.Spec.Id, s.Spec.Host)
	} else if port != "" {
		return s, fmt.Errorf("mysql solo %d host must not contains port: %s", s.Spec.Id, s.Spec.Host)
	}
	if !r.Status.Version.ValidateMasterHost(r.SoloShortHost(s.Spec.Id)) {
		return s, fmt.Errorf("mysql solo %d master host too long", s.Spec.Id)
	}

	return s, nil
}
