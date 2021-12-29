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

package projectyml

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type ProjectYml struct {
	data []byte
	s    *Spec
}

func New(b []byte, ops ...Option) (*ProjectYml, error) {
	p := ProjectYml{
		data: b,
		s:    &Spec{},
	}
	if err := p.parse(); err != nil {
		return nil, err
	}
	return &p, nil
}

type Option func(yml *ProjectYml)

func (p *ProjectYml) parse() error {
	var s Spec
	if err := yaml.Unmarshal(p.data, &s); err != nil {
		return errors.Wrap(err, "fail to yaml unmarshal")
	}
	p.s = &s
	return nil
}
