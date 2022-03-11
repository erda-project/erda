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

package k8sflink

import (
	"context"
	"fmt"
	"strings"

	flinkoperatorv1beta1 "github.com/googlecloudplatform/flink-operator/api/v1beta1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/logic"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/modules/pipeline/pkg/clusterinfo"
	"github.com/erda-project/erda/modules/pipeline/pkg/container_provider"
	"github.com/erda-project/erda/modules/pipeline/pkg/task_uuid"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	DiceRootDomain    = "DICE_ROOT_DOMAIN"
	K8SFlinkLogPrefix = "[k8sflink]"
)

var Kind = types.Kind(spec.PipelineTaskExecutorKindK8sFlink)

func init() {
	types.MustRegister(Kind, func(name types.Name, options map[string]string) (types.ActionExecutor, error) {
		clusterName, err := Kind.GetClusterNameByExecutorName(name)
		if err != nil {
			return nil, err
		}
		cluster, err := clusterinfo.GetClusterByName(clusterName)
		if err != nil {
			return nil, err
		}
		return New(name, clusterName, cluster)
	})
}

func (k *K8sFlink) Status(ctx context.Context, task *spec.PipelineTask) (statusDesc apistructs.PipelineStatusDesc, err error) {
	defer k.errWrapper.WrapTaskError(&err, "status job", task)
	if err := logic.ValidateAction(task); err != nil {
		return apistructs.PipelineStatusDesc{}, err
	}

	bigDataConf, err := logic.GetBigDataConf(task)
	if err != nil {
		return statusDesc, err
	}

	logrus.Debugf("get status from name %s in namespace %s", task.Extra.UUID, task.Extra.Namespace)
	flinkCluster, err := k.GetFlinkClusterInfo(ctx, bigDataConf)
	if err != nil {
		logrus.Errorf("get status err %v", err)

		if strings.Contains(err.Error(), "not found") {
			statusDesc.Status = logic.TransferStatus(string(apistructs.StatusNotFoundInCluster))
			return statusDesc, nil
		}
		statusDesc.Status = logic.TransferStatus(string(apistructs.StatusError))
		statusDesc.Desc = err.Error()
		return statusDesc, err
	}

	status := composeStatusDesc(flinkCluster.Status)
	statusDesc.Status = logic.TransferStatus(string(status.Status))
	statusDesc.Desc = status.Reason

	return statusDesc, nil
}

func (k *K8sFlink) Exist(ctx context.Context, task *spec.PipelineTask) (created bool, started bool, err error) {
	statusDesc, err := k.Status(ctx, task)
	if err != nil {
		created = false
		started = false
		if strutil.Contains(err.Error(), "failed to inspect job, err: not found") {
			err = nil
			return
		}
		return
	}
	return logic.JudgeExistedByStatus(statusDesc)
}

func (k *K8sFlink) Create(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	defer k.errWrapper.WrapTaskError(&err, "create job", task)
	if err := logic.ValidateAction(task); err != nil {
		return nil, err
	}
	created, _, err := k.Exist(ctx, task)
	if err != nil {
		return nil, err
	}
	if created {
		logrus.Warnf("%s: task already created, taskInfo: %s", k.Kind(), logic.PrintTaskInfo(task))
	}
	return nil, nil
}

func (k *K8sFlink) Start(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	defer k.errWrapper.WrapTaskError(&err, "start job", task)
	if err := logic.ValidateAction(task); err != nil {
		return nil, err
	}
	created, started, err := k.Exist(ctx, task)
	if err != nil {
		return nil, err
	}
	if !created {
		logrus.Warnf("%s: task not created, try to create actionInfo: %s", k.Kind().String(), logic.PrintTaskInfo(task))
		_, err = k.Create(ctx, task)
		if err != nil {
			return nil, err
		}
		logrus.Warnf("k8sjob: action created, continue to start, actionInfo: %s", logic.PrintTaskInfo(task))
	}
	if started {
		logrus.Warnf("%s: task already started, actionInfo: %s", k.Kind().String(), logic.PrintTaskInfo(task))
		return nil, nil
	}

	job, err := logic.TransferToSchedulerJob(task)
	if err != nil {
		return nil, err
	}

	bigDataConf, err := logic.GetBigDataConf(task)
	if err != nil {
		return nil, err
	}

	clusterInfo, err := logic.GetCLusterInfo(job.ClusterName)
	if err != nil {
		return apistructs.Job{
			JobFromUser: job,
		}, err
	}

	ns := &corev1.Namespace{}
	statusDesc := apistructs.StatusDesc{}
	if ns, err = k.client.ClientSet.CoreV1().Namespaces().Get(ctx, task.Extra.Namespace, metav1.GetOptions{}); err != nil {
		if !strings.Contains(err.Error(), "not found") {
			statusDesc.Status = apistructs.StatusError
			statusDesc.Reason = err.Error()
			statusDesc.LastMessage = err.Error()
			return apistructs.Job{
				JobFromUser: job,
				StatusDesc:  statusDesc,
			}, fmt.Errorf("get namespace err: %v", err)
		}

		logrus.Debugf("create namespace %s", job.Namespace)
		ns = container_provider.GenNamespaceByJob(&job)

		var nsErr error
		if ns, nsErr = k.client.ClientSet.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{}); nsErr != nil {
			statusDesc.Status = apistructs.StatusError
			statusDesc.Reason = nsErr.Error()
			statusDesc.LastMessage = nsErr.Error()
			return apistructs.Job{
				JobFromUser: job,
				StatusDesc:  statusDesc,
			}, fmt.Errorf("create namespace err: %v", nsErr)
		}
	}

	if err := k.createImageSecretIfNotExist(job.Namespace); err != nil {
		return apistructs.Job{
			JobFromUser: job,
		}, err
	}

	_, _, pvcs := logic.GenerateK8SVolumes(&job, clusterInfo)
	for _, pvc := range pvcs {
		if pvc == nil {
			continue
		}
		_, err := k.client.ClientSet.
			CoreV1().
			PersistentVolumeClaims(job.Namespace).
			Create(context.Background(), pvc, metav1.CreateOptions{})
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			return nil, err
		}
	}
	for i := range pvcs {
		if pvcs[i] == nil {
			continue
		}
		job.Volumes[i].ID = &(pvcs[i].Name)
	}

	logrus.Debugf("create flink cluster cr name %s in namespace %s", job.Name, ns.Name)

	hosts := append([]string{FlinkIngressPrefix}, job.Namespace, clusterInfo[DiceRootDomain])
	hostURL := strings.Join(hosts, ".")
	flinkCluster := k.ComposeFlinkCluster(job, bigDataConf, hostURL)
	flinkCluster.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		composeOwnerReferences("v1", "Namespace", ns.Name, ns.UID),
	}
	if err := k.client.CRClient.Create(ctx, flinkCluster); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return apistructs.Job{
				JobFromUser: job,
			}, nil
		}
		if !job.NotPipelineControlledNs {
			if delErr := k.client.ClientSet.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{}); delErr != nil {
				return nil, fmt.Errorf("delete namespace err: %v", delErr)
			}
		}
		statusDesc.Status = apistructs.StatusError
		statusDesc.Reason = err.Error()
		statusDesc.LastMessage = err.Error()
		return nil, fmt.Errorf("create flink cluster %s err: %s", flinkCluster.ClusterName, err.Error())
	}
	return apistructs.Job{
		JobFromUser: job,
	}, nil
}

func (k *K8sFlink) Update(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	return nil, errors.New("k8s(flink) not support update operation")
}

func (k *K8sFlink) Cancel(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	defer k.errWrapper.WrapTaskError(&err, "cancel job", task)
	if err := logic.ValidateAction(task); err != nil {
		return nil, err
	}
	// TODO move all makeJobID to framework
	// now move makeJobID to framework may change task uuid in database
	// Restore the task uuid after remove, because gc will make the job id, but cancel don't make the job id
	oldUUID := task.Extra.UUID
	task.Extra.UUID = task_uuid.MakeJobID(task)
	d, err := k.delete(ctx, task)
	task.Extra.UUID = oldUUID
	return d, err
}

func (k *K8sFlink) Remove(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	defer k.errWrapper.WrapTaskError(&err, "remove job", task)
	if err := logic.ValidateAction(task); err != nil {
		return nil, err
	}
	task.Extra.UUID = task_uuid.MakeJobID(task)
	return k.delete(ctx, task)
}

func (k *K8sFlink) delete(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	bigDataConf, err := logic.GetBigDataConf(task)

	flinkCluster, err := k.GetFlinkClusterInfo(ctx, bigDataConf)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s failed to get flink cluster: %s, err: %v", K8SFlinkLogPrefix, task.Name, err)
	}

	err = k.client.CRClient.Delete(ctx, flinkCluster)
	if err != nil {
		return nil, fmt.Errorf("delete flink cluster %s err: %s", bigDataConf.Name, err.Error())
	}

	// delete namespace after gc flinkcluster
	namespace := task.Extra.Namespace
	if !task.Extra.NotPipelineControlledNs {
		flinkClusters := flinkoperatorv1beta1.FlinkClusterList{}
		err = k.client.CRClient.List(context.Background(), &flinkClusters, &client.ListOptions{
			Namespace: namespace,
		})
		if err != nil {
			return nil, fmt.Errorf("%s list k8sflink clusters error: %+v, namespace: %s", K8SFlinkLogPrefix, err, namespace)
		}
		remainCount := 0
		if len(flinkClusters.Items) != 0 {
			for _, f := range flinkClusters.Items {
				if f.DeletionTimestamp == nil {
					remainCount++
				}
			}
		}

		if remainCount < 1 {
			ns, err := k.client.ClientSet.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return nil, nil
				}
				return nil, fmt.Errorf("%s get the namespace %s, error: %+v", K8SFlinkLogPrefix, namespace, err)
			}

			if ns.DeletionTimestamp == nil {
				logrus.Debugf("%s start to delete the namespace %s", K8SFlinkLogPrefix, namespace)
				err = k.client.ClientSet.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
				if err != nil {
					if !k8serrors.IsNotFound(err) {
						errMsg := fmt.Errorf("%s delete the namespace %s, error: %+v", K8SFlinkLogPrefix, namespace, err)
						return nil, errMsg
					}
					logrus.Warningf("%s not found the namespace %s", K8SFlinkLogPrefix, namespace)
				}
				logrus.Debugf("%s clean namespace %s successfully", K8SFlinkLogPrefix, namespace)
			}
		}
	}
	return nil, nil
}

func (k *K8sFlink) BatchDelete(ctx context.Context, tasks []*spec.PipelineTask) (data interface{}, err error) {
	if len(tasks) == 0 {
		return nil, nil
	}
	task := tasks[0]
	defer k.errWrapper.WrapTaskError(&err, "batch delete job", task)
	for _, task := range tasks {
		if len(task.Extra.UUID) <= 0 {
			continue
		}
		_, err := k.delete(ctx, task)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// Inspect k8sflink doesn`t support inspect, flinkcluster`s logs are too long
func (k *K8sFlink) Inspect(ctx context.Context, task *spec.PipelineTask) (apistructs.TaskInspect, error) {
	return apistructs.TaskInspect{}, errors.New("k8sflink doesn`t support inspect")
}

func (k *K8sFlink) GetFlinkClusterInfo(ctx context.Context, data apistructs.BigdataConf) (*flinkoperatorv1beta1.FlinkCluster, error) {
	logrus.Debugf("get flinkCluster name %s in ns %s", data.Name, data.Namespace)

	flinkCluster := flinkoperatorv1beta1.FlinkCluster{}
	err := k.client.CRClient.Get(context.Background(), client.ObjectKey{
		Name:      data.Name,
		Namespace: data.Namespace,
	}, &flinkCluster)
	if err != nil {
		return nil, fmt.Errorf("get flinkcluster %s in ns %s is err: %s", data.Name, data.Namespace, err.Error())
	}

	return &flinkCluster, nil
}

func (k *K8sFlink) createImageSecretIfNotExist(namespace string) error {
	if _, err := k.client.ClientSet.CoreV1().Secrets(namespace).Get(context.Background(), apistructs.AliyunRegistry, metav1.GetOptions{}); err == nil {
		return nil
	}

	s, err := k.client.ClientSet.CoreV1().Secrets(metav1.NamespaceDefault).Get(context.Background(), apistructs.AliyunRegistry, metav1.GetOptions{})
	if err != nil {
		return err
	}
	mySecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: namespace,
		},
		Data:       s.Data,
		StringData: s.StringData,
		Type:       s.Type,
	}

	if _, err := k.client.ClientSet.CoreV1().Secrets(namespace).Create(context.Background(), mySecret, metav1.CreateOptions{}); err != nil {
		if strutil.Contains(err.Error(), "AlreadyExists") {
			return nil
		}
		return err
	}
	return nil
}
