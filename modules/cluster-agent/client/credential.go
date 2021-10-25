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
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cluster-agent/config"
)

var (
	accessKey string
	lock      sync.Mutex
)

func setAccessKey(ak string) {
	lock.Lock()
	defer lock.Unlock()
	accessKey = ak
}

func getAccessKey() string {
	return accessKey
}

func WatchClusterCredential(ctx context.Context, cfg *config.Config) error {
	rc, err := rest.InClusterConfig()
	if err != nil {
		logrus.Errorf("get incluster config error: %v", err)
		return err
	}

	cs, err := kubernetes.NewForConfig(rc)
	if err != nil {
		logrus.Errorf("create clientset error: %v", err)
		return err
	}

	secInit, err := cs.CoreV1().Secrets(cfg.ErdaNamespace).Get(context.Background(), apistructs.ErdaClusterCredential, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get configmap error: %v", err)
	}

	// create retry watcher
	retryWatcher, err := watchtools.NewRetryWatcher(secInit.ResourceVersion, &cache.ListWatch{
		WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
			return cs.CoreV1().Secrets(cfg.ErdaNamespace).Watch(context.Background(), v1.ListOptions{
				FieldSelector: fmt.Sprintf("metadata.name=%s", apistructs.ErdaClusterCredential),
			})
		},
	})

	if err != nil {
		logrus.Errorf("create retry watcher error: %v", err)
		return err
	}

	defer retryWatcher.Stop()

	logrus.Info("start retry watcher")

	for {
		select {
		case event := <-retryWatcher.ResultChan():
			sec, ok := event.Object.(*corev1.Secret)
			if !ok {
				logrus.Errorf("illegal secret object, igonre")
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

				// Access key values doesn't change, skip reconnect
				if string(ak) == getAccessKey() {
					logrus.Info("cluster access key doesn't change, skip")
					continue
				}

				logrus.Infof("cluster accesskey change from %s to %s", getAccessKey(), string(ak))

				// change value
				setAccessKey(string(ak))

				select {
				case <-Connected():
					disConnected <- struct{}{}
				default:
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}
