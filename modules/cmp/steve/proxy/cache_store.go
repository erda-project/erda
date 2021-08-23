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

package proxy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/rancher/apiserver/pkg/apierror"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/steve/pkg/accesscontrol"
	"github.com/rancher/steve/pkg/attributes"
	"github.com/rancher/wrangler/pkg/schemas/validation"
	"github.com/sirupsen/logrus"
	"k8s.io/apiserver/pkg/endpoints/request"

	"github.com/erda-project/erda/modules/cmp/cache"
)

type cacheStore struct {
	types.Store

	ctx     context.Context
	asl     accesscontrol.AccessSetLookup
	cache   *cache.Cache
	cancels sync.Map
}

type cacheKey struct {
	gvk       string
	namespace string
}

func (k *cacheKey) getKey() string {
	d := sha256.New()
	d.Write([]byte(k.gvk))
	d.Write([]byte(k.namespace))
	return hex.EncodeToString(d.Sum(nil))
}

func (c *cacheStore) List(apiOp *types.APIRequest, schema *types.APISchema) (types.APIObjectList, error) {
	if !c.hasAccess(apiOp, schema, "list") {
		return types.APIObjectList{}, apierror.NewAPIError(validation.PermissionDenied, "access denied")
	}
	gvk := attributes.GVK(schema)
	key := cacheKey{
		gvk:       gvk.String(),
		namespace: apiOp.Namespace,
	}

	values, _, err := c.cache.Get(key.getKey())
	if values == nil || err != nil {
		if apiOp.Namespace != "" {
			key := cacheKey{
				gvk:       gvk.String(),
				namespace: "",
			}
			allNsValues, expired, err := c.cache.Get(key.getKey())
			if allNsValues != nil && err == nil && !expired {
				var list types.APIObjectList
				if err = json.Unmarshal(allNsValues[0].Value().([]byte), &list); err == nil {
					return getByNamespace(list, apiOp.Namespace), nil
				}
			}
		}

		list, err := c.Store.List(apiOp, schema)
		if err != nil {
			return types.APIObjectList{}, err
		}
		vals, err := cache.MarshalValue(list)
		if err != nil {
			logrus.Errorf("failed to marshal cache data for %s, %v", gvk.Kind, err)
			return types.APIObjectList{}, apierror.NewAPIError(validation.ServerError, "internal error")
		}
		c.cache.Set(key.getKey(), vals, time.Second.Nanoseconds()*30)
		return list, nil
	}

	go func() {
		list, err := c.Store.List(apiOp, schema)
		if err != nil {
			logrus.Errorf("failed to list %s in steve cache store, %v", gvk.Kind, err)
			return
		}
		data, err := cache.MarshalValue(list)
		if err != nil {
			logrus.Errorf("failed to marshal cache data for %s, %v", gvk.Kind, err)
			return
		}
		c.cache.Set(key.getKey(), data, time.Second.Nanoseconds()*30)
	}()

	var list types.APIObjectList
	if err = json.Unmarshal(values[0].Value().([]byte), &list); err != nil {
		logrus.Errorf("failed to marshal list %s result, %v", gvk.Kind, err)
		return types.APIObjectList{}, apierror.NewAPIError(validation.ServerError, "internal error")
	}
	return list, nil
}

func (c *cacheStore) Create(apiOp *types.APIRequest, schema *types.APISchema, data types.APIObject) (types.APIObject, error) {
	gvk := attributes.GVK(schema)
	key := cacheKey{
		gvk:       gvk.String(),
		namespace: apiOp.Namespace,
	}
	c.cache.Remove(key.getKey())
	return c.Store.Create(apiOp, schema, data)
}

func (c *cacheStore) Update(apiOp *types.APIRequest, schema *types.APISchema, data types.APIObject, id string) (types.APIObject, error) {
	gvk := attributes.GVK(schema)
	key := cacheKey{
		gvk:       gvk.String(),
		namespace: apiOp.Namespace,
	}
	c.cache.Remove(key.getKey())
	return c.Store.Update(apiOp, schema, data, id)
}

func (c *cacheStore) Delete(apiOp *types.APIRequest, schema *types.APISchema, id string) (types.APIObject, error) {
	gvk := attributes.GVK(schema)
	key := cacheKey{
		gvk:       gvk.String(),
		namespace: apiOp.Namespace,
	}
	c.cache.Remove(key.getKey())
	return c.Store.Delete(apiOp, schema, id)
}

func (c *cacheStore) hasAccess(apiOp *types.APIRequest, schema *types.APISchema, verb string) bool {
	user, ok := request.UserFrom(apiOp.Context())
	if !ok {
		return false
	}
	access := c.asl.AccessFor(user)
	gr := attributes.GR(schema)
	ns := apiOp.Namespace
	if ns == "" {
		ns = "*"
	}
	return access.Grants(verb, gr, ns, attributes.Resource(schema))
}

func getByNamespace(list types.APIObjectList, namespace string) types.APIObjectList {
	res := types.APIObjectList{
		Revision: "-1",
	}
	for _, apiObj := range list.Objects {
		if apiObj.Namespace() == namespace {
			res.Objects = append(res.Objects, apiObj)
		}
	}
	return res
}
