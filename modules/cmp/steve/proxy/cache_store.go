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
	"os"
	"strconv"
	"time"

	jsi "github.com/json-iterator/go"
	"github.com/rancher/apiserver/pkg/apierror"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/steve/pkg/accesscontrol"
	"github.com/rancher/steve/pkg/attributes"
	"github.com/rancher/wrangler/pkg/schemas/validation"
	"github.com/sirupsen/logrus"
	"k8s.io/apiserver/pkg/endpoints/request"

	"github.com/erda-project/erda/modules/cmp/cache"
	"github.com/erda-project/erda/modules/cmp/queue"
)

var queryQueue *queue.QueryQueue

func init() {
	queueSize := 10
	if size, err := strconv.Atoi(os.Getenv("LIST_QUEUE_SIZE")); err == nil && size > queueSize {
		queueSize = size
	}
	queryQueue = queue.NewQueryQueue(queueSize)
}

type cacheStore struct {
	types.Store

	ctx         context.Context
	asl         accesscontrol.AccessSetLookup
	cache       *cache.Cache
	clusterName string
}

type cacheKey struct {
	gvk         string
	namespace   string
	clusterName string
}

func (k *cacheKey) getKey() string {
	d := sha256.New()
	d.Write([]byte(k.gvk))
	d.Write([]byte(k.namespace))
	d.Write([]byte(k.clusterName))
	return hex.EncodeToString(d.Sum(nil))
}

func (c *cacheStore) List(apiOp *types.APIRequest, schema *types.APISchema) (types.APIObjectList, error) {
	if !c.hasAccess(apiOp, schema, "list") {
		return types.APIObjectList{}, apierror.NewAPIError(validation.PermissionDenied, "access denied")
	}
	gvk := attributes.GVK(schema)
	key := cacheKey{
		gvk:         gvk.String(),
		namespace:   apiOp.Namespace,
		clusterName: c.clusterName,
	}

	logrus.Infof("[DEBUG %s] start get cache at %s", apiOp.Type, time.Now().Format(time.StampNano))
	values, lexpired, err := c.cache.Get(key.getKey())
	logrus.Infof("[DEBUG %s] end get cache at %s", apiOp.Type, time.Now().Format(time.StampNano))
	if values == nil || err != nil {
		if apiOp.Namespace != "" {
			key := cacheKey{
				gvk:         gvk.String(),
				namespace:   "",
				clusterName: c.clusterName,
			}
			allNsValues, expired, err := c.cache.Get(key.getKey())
			if allNsValues != nil && err == nil && !expired {
				var list types.APIObjectList
				logrus.Infof("[DEBUG %s] start jsi unmarshal data from cache at %s", apiOp.Type, time.Now().Format(time.StampNano))
				if err = jsi.Unmarshal(allNsValues[0].Value().([]byte), &list); err == nil {
					logrus.Infof("[DEBUG %s] end jsi unmarshal data from cache at %s", apiOp.Type, time.Now().Format(time.StampNano))
					logrus.Infof("[DEBUG %s] start get by namespace at %s", apiOp.Type, time.Now().Format(time.StampNano))
					list := getByNamespace(list, apiOp.Namespace)
					logrus.Infof("[DEBUG %s] end get by namespace at %s", apiOp.Type, time.Now().Format(time.StampNano))
					return list, nil
				}
			}
		}

		logrus.Infof("[DEBUG %s] start list at %s", apiOp.Type, time.Now().Format(time.StampNano))
		queryQueue.Acquire(c.clusterName, 1)
		list, err := c.Store.List(apiOp, schema)
		queryQueue.Release(c.clusterName, 1)
		if err != nil {
			return types.APIObjectList{}, err
		}
		logrus.Infof("[DEBUG %s] end list at %s", apiOp.Type, time.Now().Format(time.StampNano))
		logrus.Infof("[DEBUG %s] start marshal for cache at %s", apiOp.Type, time.Now().Format(time.StampNano))
		vals, err := cache.MarshalValue(list)
		logrus.Infof("[DEBUG %s] end marshal for cache at %s", apiOp.Type, time.Now().Format(time.StampNano))
		if err != nil {
			logrus.Errorf("failed to marshal cache data for %s, %v", gvk.Kind, err)
			return types.APIObjectList{}, apierror.NewAPIError(validation.ServerError, "internal error")
		}
		logrus.Infof("[DEBUG %s] start set cache at %s", apiOp.Type, time.Now().Format(time.StampNano))
		if err = c.cache.Set(key.getKey(), vals, time.Second.Nanoseconds()*30); err != nil {
			logrus.Errorf("failed to set cache for %s, %v", gvk.String(), err)
		}
		logrus.Infof("[DEBUG %s] end set cache at %s", apiOp.Type, time.Now().Format(time.StampNano))
		return list, nil
	}

	if lexpired {
		logrus.Infof("list data is expired, need update, key:%s", key.getKey())
		if !cache.ExpireFreshQueue.IsFull() {
			task := &queue.Task{
				Key: key.getKey(),
				Do: func() {
					user, ok := request.UserFrom(apiOp.Context())
					if !ok {
						logrus.Errorf("user not found in context when steve auth")
						return
					}
					ctx := request.WithUser(c.ctx, user)
					newOp := apiOp.WithContext(ctx)
					list, err := c.Store.List(newOp, schema)
					if err != nil {
						logrus.Errorf("failed to list %s in steve cache store, %v", gvk.Kind, err)
						return
					}
					data, err := cache.MarshalValue(list)
					if err != nil {
						logrus.Errorf("failed to marshal cache data for %s, %v", gvk.Kind, err)
						return
					}
					if err = c.cache.Set(key.getKey(), data, time.Second.Nanoseconds()*30); err != nil {
						logrus.Errorf("failed to set cache for %s, %v", gvk.String(), err)
					}
				},
			}
			cache.ExpireFreshQueue.Enqueue(task)
		} else {
			logrus.Warnf("queue size is full, task is ignored, key:%s", key.getKey())
		}
	}

	var list types.APIObjectList
	logrus.Infof("[DEBUG %s] start unmarshal data from cache at %s", apiOp.Type, time.Now().Format(time.StampNano))
	if err = jsi.Unmarshal(values[0].Value().([]byte), &list); err != nil {
		logrus.Errorf("failed to marshal list %s result, %v", gvk.Kind, err)
		return types.APIObjectList{}, apierror.NewAPIError(validation.ServerError, "internal error")
	}
	logrus.Infof("[DEBUG %s] end unmarshal data from cache at %s", apiOp.Type, time.Now().Format(time.StampNano))
	return list, nil
}

func (c *cacheStore) Create(apiOp *types.APIRequest, schema *types.APISchema, data types.APIObject) (types.APIObject, error) {
	gvk := attributes.GVK(schema)
	key := cacheKey{
		gvk:         gvk.String(),
		namespace:   apiOp.Namespace,
		clusterName: c.clusterName,
	}
	if _, err := c.cache.Remove(key.getKey()); err != nil {
		logrus.Errorf("failed to remove cache for %s, %v", gvk.String(), err)
	}
	return c.Store.Create(apiOp, schema, data)
}

func (c *cacheStore) Update(apiOp *types.APIRequest, schema *types.APISchema, data types.APIObject, id string) (types.APIObject, error) {
	gvk := attributes.GVK(schema)
	key := cacheKey{
		gvk:         gvk.String(),
		namespace:   apiOp.Namespace,
		clusterName: c.clusterName,
	}
	if _, err := c.cache.Remove(key.getKey()); err != nil {
		logrus.Errorf("failed to remove cache for %s, %v", gvk.String(), err)
	}
	return c.Store.Update(apiOp, schema, data, id)
}

func (c *cacheStore) Delete(apiOp *types.APIRequest, schema *types.APISchema, id string) (types.APIObject, error) {
	gvk := attributes.GVK(schema)
	key := cacheKey{
		gvk:         gvk.String(),
		namespace:   apiOp.Namespace,
		clusterName: c.clusterName,
	}
	if _, err := c.cache.Remove(key.getKey()); err != nil {
		logrus.Errorf("failed to remove cache for %s, %v", gvk.String(), err)
	}
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
