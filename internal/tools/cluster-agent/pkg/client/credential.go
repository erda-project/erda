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

package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/erda-project/erda/apistructs"
)

const maxCredentialWatchRetryInterval = 30 * time.Second

var errCredentialWatcherClosed = errors.New("cluster credential watcher closed")

type credentialWatcher interface {
	ResultChan() <-chan watch.Event
	Stop()
}

func (c *Client) watchClusterCredential(ctx context.Context) error {
	kc, err := c.newInClusterClient()
	if err != nil {
		return err
	}
	cs := kc.ClientSet

	attempt := 0
	for {
		retryWatcher, err := c.newCredentialWatcher(ctx, cs, c.cfg.ErdaNamespace)
		if err != nil {
			logrus.Errorf("get retry watcher, %v", err)
		} else {
			attempt = 0
			logrus.Info("start retry watcher")
			err = c.consumeClusterCredentialEvents(ctx, retryWatcher)
			if err == nil {
				return nil
			}
			if errors.Is(err, errCredentialWatcherClosed) {
				logrus.Warn("cluster credential watcher closed, rebuild watcher")
			} else {
				logrus.Errorf("cluster credential watcher stopped, rebuild watcher: %v", err)
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.credentialWatchRetryDelay(attempt)):
			attempt++
		}
	}
}

func (c *Client) consumeClusterCredentialEvents(ctx context.Context, retryWatcher credentialWatcher) error {
	defer retryWatcher.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-retryWatcher.ResultChan():
			if !ok {
				return errCredentialWatcherClosed
			}

			if event.Type == watch.Error {
				logrus.Warnf("cluster credential watcher error event: %v", k8serrors.FromObject(event.Object))
				continue
			}

			sec, ok := event.Object.(*corev1.Secret)
			if !ok {
				logrus.Errorf("illegal secret object, ignore, content: %+v", event.Object)
				continue
			}

			logrus.Debugf("event type: %v, object: %+v", event.Type, sec)

			switch event.Type {
			case watch.Deleted:
				// ignore delete event, if cluster offline, reconnect will be failed.
				continue
			case watch.Added, watch.Modified:
				ak, ok := sec.Data[apistructs.ClusterAccessKey]
				// If accidentally deleted credential key, use the latest access key
				if !ok {
					logrus.Errorf("cluster info doesn't contain cluster access key value")
					continue
				}

				currentAccessKey := c.getAccessKey()

				// Access key values doesn't change, skip reconnect
				if string(ak) == currentAccessKey {
					logrus.Debug("cluster access key doesn't change, skip")
					continue
				}

				if currentAccessKey == "" {
					logrus.Info("get cluster accesskey")
				} else {
					logrus.Info("cluster accesskey changed")
				}

				// change value
				c.setAccessKey(string(ak))
				c.requestReconnect()
			}
		}
	}
}

func (c *Client) credentialWatchRetryDelay(attempt int) time.Duration {
	delay := c.watchRetryInterval
	if delay <= 0 {
		delay = time.Second
	}

	if attempt < 0 {
		attempt = 0
	}
	if attempt > 5 {
		attempt = 5
	}

	delay *= time.Duration(1 << attempt)
	if delay > maxCredentialWatchRetryInterval {
		return maxCredentialWatchRetryInterval
	}
	return delay
}

func (c *Client) setAccessKey(ac string) {
	c.Lock()
	defer c.Unlock()
	c.accessKey = ac
}

func (c *Client) getAccessKey() string {
	c.Lock()
	defer c.Unlock()
	return c.accessKey
}

func (c *Client) getRetryWatcher(ctx context.Context, cs kubernetes.Interface, ns string) (*watchtools.RetryWatcher, error) {
	selector, err := clusterCredentialSelector()
	if err != nil {
		return nil, fmt.Errorf("parse selector error: %v", err)
	}

	secInit, err := c.listClusterCredentialSecrets(ctx, cs, ns, selector)
	if err != nil {
		return nil, fmt.Errorf("get init secret list error: %v", err)
	}

	c.loadInitialClusterAccessKey(secInit)

	retryWatcher, err := newClusterCredentialRetryWatcher(ctx, cs, ns, selector, secInit.ResourceVersion)

	if err != nil {
		return nil, fmt.Errorf("create retry watcher error: %v", err)
	}

	return retryWatcher, nil
}

func clusterCredentialSelector() (fields.Selector, error) {
	return fields.ParseSelector(fmt.Sprintf("metadata.name=%s", apistructs.ErdaClusterCredential))
}

func (c *Client) listClusterCredentialSecrets(ctx context.Context, cs kubernetes.Interface, ns string, selector fields.Selector) (*corev1.SecretList, error) {
	return cs.CoreV1().Secrets(ns).List(ctx, metav1.ListOptions{
		FieldSelector: selector.String(),
	})
}

func (c *Client) loadInitialClusterAccessKey(secInit *corev1.SecretList) {
	if secInit != nil && len(secInit.Items) > 0 {
		initData, ok := secInit.Items[0].Data[apistructs.ClusterAccessKey]
		if !ok {
			logrus.Warn("no valid cluster access key was got")
			return
		}

		logrus.Info("load initial cluster access key")
		c.setAccessKey(string(initData))
		return
	}

	logrus.Warn("no valid cluster access key was got")
}

func newClusterCredentialRetryWatcher(ctx context.Context, cs kubernetes.Interface, ns string, selector fields.Selector, resourceVersion string) (*watchtools.RetryWatcher, error) {
	return watchtools.NewRetryWatcher(resourceVersion, &cache.ListWatch{
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = selector.String()
			return cs.CoreV1().Secrets(ns).Watch(ctx, options)
		},
	})
}
