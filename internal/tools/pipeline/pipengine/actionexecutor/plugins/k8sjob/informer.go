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

package k8sjob

import (
	"time"

	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/pipengine/actionexecutor/logic"
)

func (k *K8sJob) informer() {
	labelOptions := informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.LabelSelector = "dice/job="
	})
	factory := informers.NewSharedInformerFactoryWithOptions(k.Client.ClientSet, 5*time.Second, labelOptions)
	jobInformer := factory.Batch().V1().Jobs().Informer()
	podLister := factory.Core().V1().Pods().Lister()
	jobInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			latestJob := newObj.(*batchv1.Job)
			selector, err := labels.Parse("job-name==" + latestJob.Name)
			if err != nil {
				logrus.Errorf("failed to parse selector, %v", err)
				return
			}
			logrus.Infof("k8s jobInformer: job %s updated, status: %s", latestJob.Name, latestJob.Status.Conditions)
			pods, err := podLister.List(selector)
			if err != nil {
				logrus.Errorf("failed to list pods, %v", err)
				return
			}
			podItems := &corev1.PodList{}
			for _, pod := range pods {
				podItems.Items = append(podItems.Items, *pod)
			}
			status := generateKubeJobStatus(latestJob, podItems, "")
			if len(status.Status) == 0 {
				logrus.Warnf("no status found for job %s", latestJob.Name)
				return
			}
			statusDesc := apistructs.PipelineStatusDesc{
				Status: logic.TransferStatus(string(status.Status)),
				Desc:   status.LastMessage,
			}
			k.PublishEvent(latestJob.Name, statusDesc)
		},
	})
	factory.Start(k.StopCh)
}
