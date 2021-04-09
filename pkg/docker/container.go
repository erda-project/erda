// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package docker

type container struct {
	id     string
	name   string
	image  string
	envs   map[string]string
	labels map[string]string
}

func (c *container) GetID() string {
	return c.id
}

func (c *container) GetName() string {
	return c.name
}

func (c *container) GetImage() string {
	return c.image
}

func (c *container) GetEnvs() map[string]string {
	envs := make(map[string]string)
	for k, v := range c.envs {
		envs[k] = v
	}
	return envs
}

func (c *container) Envs() map[string]string {
	return c.envs
}

func (c *container) GetEnv(key string) string {
	return c.envs[key]
}

func (c *container) LookUpEnv(key string) (val string, ok bool) {
	val, ok = c.envs[key]
	return
}

func (c *container) GetLabels() map[string]string {
	labels := make(map[string]string)
	for k, v := range c.labels {
		labels[k] = v
	}
	return labels
}

func (c *container) Labels() map[string]string {
	return c.labels
}

func (c *container) GetLabel(key string) string {
	return c.labels[key]
}

func (c *container) LookUpLabel(key string) (val string, ok bool) {
	val, ok = c.labels[key]
	return
}
