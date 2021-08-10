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

package register

import (
	"context"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/eventbox/constant"
	"github.com/erda-project/erda/modules/eventbox/register/label"
	"github.com/erda-project/erda/modules/eventbox/types"
	"github.com/erda-project/erda/pkg/jsonstore"
)

type Register interface {
	PrefixGet(key string) map[types.LabelKey]map[types.LabelKey]interface{}
	Put(key string, labels map[types.LabelKey]interface{}) error
	Del(key string) error
}

type registerImpl struct {
	js     jsonstore.JsonStore
	stopch chan struct{}
	stopwg sync.WaitGroup
}

func New() (*registerImpl, error) {
	opt := jsonstore.UseMemEtcdStore(context.Background(), constant.RegisterDir, nil, nil)
	js, err := jsonstore.New(opt)
	if err != nil {
		return nil, err
	}
	r := &registerImpl{
		js:     js,
		stopch: make(chan struct{}),
	}
	// put default labels
	for registerLabel, v := range label.DefaultLabels {
		if err := r.Put(string(registerLabel), v); err != nil {
			logrus.Errorf("Register: %v", err)
			return nil, err
		}
	}
	return r, nil
}

// return: map[registered-label]map[converted-label]<label-value>
func (r *registerImpl) PrefixGet(key string) map[types.LabelKey]map[types.LabelKey]interface{} {
	normalizedKey := types.LabelKey(key).Normalize()
	path := filepath.Join(constant.RegisterDir, normalizedKey)
	res := map[types.LabelKey]map[types.LabelKey]interface{}{}
	if err := r.js.ForEach(context.Background(), path, map[types.LabelKey]interface{}{},
		func(k string, v_ interface{}) error {
			v := v_.(*map[types.LabelKey]interface{})
			stripped := strings.TrimPrefix(k, constant.RegisterDir)
			res[types.LabelKey(stripped).NormalizeLabelKey()] = *v
			return nil
		}); err != nil {
		return nil
	}
	return res
}

func (r *registerImpl) Put(key string, labels map[types.LabelKey]interface{}) error {
	normalizedKey := types.LabelKey(key).Normalize()
	normalizedLabels := map[types.LabelKey]interface{}{}
	for k, l := range labels {
		normalizedLabels[k.NormalizeLabelKey()] = l
	}

	path := filepath.Join(constant.RegisterDir, normalizedKey)
	if err := r.js.Put(context.Background(), path, normalizedLabels); err != nil {
		return err
	}
	return nil
}

func (r *registerImpl) Del(key string) error {
	normalizedKey := types.LabelKey(key).Normalize()
	path := filepath.Join(constant.RegisterDir, normalizedKey)
	labels := make(map[types.LabelKey]interface{})
	if err := r.js.Remove(context.Background(), path, &labels); err != nil {
		return err
	}
	return nil
}
