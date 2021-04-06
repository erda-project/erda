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

package k8sflink

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8sjob"
	flinkoperatorv1beta1 "github.com/erda-project/erda/pkg/clientgo/apis/flinkoperator/v1beta1"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	kind               = "K8SFLINK"
	DiceRootDomainKEY  = "DICE_ROOT_DOMAIN"
	DiceClusterInfoKey = "dice-cluster-info"
)

func init() {
	executortypes.Register(kind, func(name executortypes.Name, clustername string, options map[string]string, moreoptions interface{}) (executortypes.Executor, error) {
		addr := options["ADDR"]
		f := New(
			WithClient(addr),
			WithKind(kind),
			WithName(name),
		)
		return f, nil
	})
}

// Kind implements executortypes.Executor interface
func (f *Flink) Kind() executortypes.Kind {
	return f.ExecutorKind
}

// Name implements executortypes.Executor interface
func (f *Flink) Name() executortypes.Name {
	return f.ExecutorName
}

// Create method create flinkcluster cr to flink-operator by controller-runtime client
func (f *Flink) Create(ctx context.Context, spec interface{}) (interface{}, error) {
	job, ok := spec.(apistructs.Job)
	if !ok {
		return nil, fmt.Errorf("invalid job spec")
	}

	ns := &corev1.Namespace{}
	var err error
	if ns, err = f.Client.K8sClient.CoreV1().Namespaces().Get(ctx, job.Namespace, metav1.GetOptions{}); err != nil {
		if !strings.Contains(err.Error(), "not found") {
			job.Status = apistructs.StatusError
			job.Reason = err.Error()
			job.LastMessage = err.Error()
			return nil, fmt.Errorf("get namespace err:%s", err.Error())
		}
		// create namespace for flinkcluster
		logrus.Infof("create namespace %s", job.Namespace)
		ns.Name = job.Namespace

		var nsErr error
		if ns, nsErr = f.Client.K8sClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{}); nsErr != nil {
			job.Status = apistructs.StatusError
			job.Reason = nsErr.Error()
			job.LastMessage = nsErr.Error()
			return nil, fmt.Errorf("create namespace err: %s", nsErr.Error())
		}
	}

	if err := f.createImageSecretIfNotExist(job.Namespace); err != nil {
		return nil, err
	}
	_, _, pvcs := k8sjob.GenerateK8SVolumes(&job)
	for _, pvc := range pvcs {
		if pvc == nil {
			continue
		}
		if _, err := f.Client.K8sClient.CoreV1().PersistentVolumeClaims(job.Namespace).Create(context.Background(), pvc, metav1.CreateOptions{}); err != nil {
			return nil, err
		}
	}
	for index := range pvcs {
		if pvcs[index] == nil {
			continue
		}
		job.Volumes[index].ID = &(pvcs[index].Name)
	}

	logrus.Infof("create flinkcluster cr name %s in namespace %s", job.Name, ns.Name)

	data, err := f.GetClusterInfo(DiceClusterInfoKey)
	if err != nil {
		return nil, err
	}
	hosts := append([]string{FlinkIngressPrefix}, job.Namespace, data[DiceRootDomainKEY])
	hostURL := strings.Join(hosts, ".")
	flinkCluster := ComposeFlinkCluster(job.BigdataConf, hostURL)
	flinkCluster.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		composeOwnerReferences("v1", "Namespace", ns.Name, ns.UID),
	}
	if _, err := f.Client.CustomClient.FlinkoperatorV1beta1().FlinkClusters(ns.Name).Create(ctx, flinkCluster, metav1.CreateOptions{}); err != nil {
		if strings.Contains(err.Error(), "already exist") {
			return flinkCluster, nil
		}
		if delErr := f.Client.K8sClient.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{}); delErr != nil {
			return nil, fmt.Errorf("delete namespace err: %s", delErr.Error())
		}
		job.Status = apistructs.StatusError
		job.Reason = err.Error()
		job.LastMessage = err.Error()
		return nil, fmt.Errorf("create flinkcluster %+v err: %s", flinkCluster, err.Error())
	}
	return job, nil
}

// Destroy method delete flinkcluster cr by Remove method
func (f *Flink) Destroy(ctx context.Context, spec interface{}) error {
	return f.Remove(ctx, spec)
}

// Remove method delete flinkcluster cr by controller-runtime client
func (f *Flink) Remove(ctx context.Context, spec interface{}) error {
	job, ok := spec.(apistructs.Job)
	if !ok {
		return fmt.Errorf("invalid job spec")
	}

	flinkCluster, err := f.GetFlinkClusterInfo(ctx, job.BigdataConf)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}

	err = f.Client.CustomClient.FlinkoperatorV1beta1().FlinkClusters(flinkCluster.Namespace).Delete(ctx, flinkCluster.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("delete flinkcluster %+v is err: %s", job.BigdataConf, err.Error())
	}
	return nil
}

// Update method update flinkcluster cr by controller-runtime client. and
// Update method is not used in Dice 3.21
func (f *Flink) Update(ctx context.Context, spec interface{}) (interface{}, error) {

	job, ok := spec.(apistructs.Job)
	if !ok {
		return nil, fmt.Errorf("invalid job spec")
	}

	bigdataConf := apistructs.BigdataConf{
		BigdataMetadata: apistructs.BigdataMetadata{
			Name:      job.Name,
			Namespace: job.Namespace,
		},
		Spec: apistructs.BigdataSpec{},
	}

	if value, ok := job.Params["bigDataConf"]; ok {
		if err := json.Unmarshal([]byte(value.(string)), &bigdataConf.Spec); err != nil {
			return nil, fmt.Errorf("unmarshal job params byte err: %v", err.Error())
		}
	}

	data, err := f.GetClusterInfo(DiceClusterInfoKey)
	if err != nil {
		return nil, err
	}
	hosts := append([]string{FlinkIngressPrefix}, job.Namespace, data[DiceRootDomainKEY])
	hostURL := strings.Join(hosts, ".")
	flinkCluster := ComposeFlinkCluster(bigdataConf, hostURL)

	_, err = f.Client.CustomClient.FlinkoperatorV1beta1().FlinkClusters(job.Namespace).Update(ctx, flinkCluster, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("update flinkcluster err: %s", err.Error())
	}

	job.Status = apistructs.StatusNotFoundInCluster
	return job, nil
}

// Inspect method get flinkCluster cr by controller-runtime client
func (f *Flink) Inspect(ctx context.Context, spec interface{}) (interface{}, error) {

	job, ok := spec.(apistructs.Job)
	if !ok {
		return nil, fmt.Errorf("invalid job spec")
	}

	flinkCluster, err := f.GetFlinkClusterInfo(ctx, job.BigdataConf)
	if err != nil {
		return nil, err
	}

	return flinkCluster, nil
}

// Status method get status from flinkcluster cr by controller-runtime client
func (f *Flink) Status(ctx context.Context, spec interface{}) (apistructs.StatusDesc, error) {

	statusDesc := apistructs.StatusDesc{}

	job, ok := spec.(apistructs.Job)
	if !ok {
		return statusDesc, fmt.Errorf("invalid job spec")
	}

	logrus.Infof("get status from name %s in namespace %s", job.Name, job.Namespace)
	flinkCluster, err := f.GetFlinkClusterInfo(ctx, job.BigdataConf)
	if err != nil {
		logrus.Infof("get status err %v", err)
		status := apistructs.StatusDesc{}

		if strings.Contains(err.Error(), "not found") {
			status.Status = apistructs.StatusNotFoundInCluster
			return status, nil
		}
		status.Status = apistructs.StatusError
		status.Reason = err.Error()
		return status, err
	}

	statusDesc = composeStatusDesc(flinkCluster.Status)

	return statusDesc, nil
}

// Cancel method cancel flinkcluster job mod and job state must is running
func (f *Flink) Cancel(ctx context.Context, spec interface{}) (interface{}, error) {

	job, ok := spec.(apistructs.Job)
	if !ok {
		return nil, fmt.Errorf("invalid job spec")
	}

	flinkCluster, err := f.GetFlinkClusterInfo(ctx, job.BigdataConf)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			job.Status = apistructs.StatusStoppedByKilled
			return job, nil
		}
		job.Status = apistructs.StatusError
		return nil, fmt.Errorf("get flink cluster error:%s", err.Error())
	}

	if flinkCluster.Spec.Job != nil && job.Status == apistructs.StatusRunning {
		flinkCluster.Spec.Job.CancelRequested = func(cancel bool) *bool { return &cancel }(true)
		if _, err = f.Client.CustomClient.FlinkoperatorV1beta1().FlinkClusters(job.Namespace).Update(ctx, flinkCluster, metav1.UpdateOptions{}); err != nil {
			job.Status = apistructs.StatusError
			return nil, fmt.Errorf("cancel flink err: %s", err.Error())
		}
		job.Status = apistructs.StatusStoppedByKilled
	}
	return job, nil
}

func (f *Flink) Precheck(ctx context.Context, spec interface{}) (apistructs.ServiceGroupPrecheckData, error) {
	return apistructs.ServiceGroupPrecheckData{}, nil
}

func (f *Flink) JobVolumeCreate(ctx context.Context, spec interface{}) (string, error) {
	return "", nil
}

func (f *Flink) SetNodeLabels(setting executortypes.NodeLabelSetting, hosts []string, labels map[string]string) error {
	return nil
}

func (f *Flink) CleanUpBeforeDelete() {
	return
}

func (f *Flink) CapacityInfo() apistructs.CapacityInfoData {
	return apistructs.CapacityInfoData{}
}

func (f *Flink) ResourceInfo(brief bool) (apistructs.ClusterResourceInfoData, error) {
	return apistructs.ClusterResourceInfoData{}, nil
}

func (f *Flink) KillPod(podname string) error {
	return nil
}

// GetFlinkClusterInfo get flinkcluster info from controller-runtime client
func (f *Flink) GetFlinkClusterInfo(ctx context.Context, data apistructs.BigdataConf) (*flinkoperatorv1beta1.FlinkCluster, error) {

	logrus.Infof("get flinkCluster name %s in ns %s", data.Name, data.Namespace)

	flinkCluster, err := f.Client.CustomClient.FlinkoperatorV1beta1().FlinkClusters(data.Namespace).Get(ctx, data.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get flinkcluster %s in ns %s is err: %s", data.Name, data.Namespace, err.Error())
	}
	return flinkCluster, nil
}

func (f *Flink) GetClusterInfo(name string) (map[string]string, error) {
	cm, err := f.Client.K8sClient.CoreV1().ConfigMaps(metav1.NamespaceDefault).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		errMsg := fmt.Errorf("get config map error %v", err)
		logrus.Error(errMsg)
		return nil, errMsg
	}
	return cm.Data, nil
}

func (f *Flink) createImageSecretIfNotExist(namespace string) error {
	if _, err := f.Client.K8sClient.CoreV1().Secrets(namespace).Get(context.Background(), AliyunPullSecret, metav1.GetOptions{}); err == nil {
		return nil
	}

	// 集群初始化的时候会在 default namespace 下创建一个拉镜像的 secret
	s, err := f.Client.K8sClient.CoreV1().Secrets(metav1.NamespaceDefault).Get(context.Background(), AliyunPullSecret, metav1.GetOptions{})
	if err != nil {
		return err
	}
	mysecret := &corev1.Secret{
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

	if _, err := f.Client.K8sClient.CoreV1().Secrets(namespace).Create(context.Background(), mysecret, metav1.CreateOptions{}); err != nil {
		if strutil.Contains(err.Error(), "AlreadyExists") {
			return nil
		}
		return err
	}
	return nil
}
