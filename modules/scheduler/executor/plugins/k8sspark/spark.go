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

package k8sspark

import (
	"context"
	"fmt"
	"strings"

	sparkv1beta2 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/clusterinfo"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/namespace"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/persistentvolumeclaim"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/role"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/rolebinding"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/secret"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/serviceaccount"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8sspark/sparkapplication"
	"github.com/erda-project/erda/modules/scheduler/executor/util"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	executorKind             = "K8SSPARK"
	jobKind                  = "SparkApplication"
	jobAPIVersion            = "sparkoperator.k8s.io/v1beta2"
	rbacAPIVersion           = "rbac.authorization.k8s.io/v1"
	rbacAPIGroup             = "rbac.authorization.k8s.io"
	sparkServiceAccountName  = "spark"
	sparkRoleName            = "spark-role"
	sparkRoleBindingName     = "spark-role-binding"
	imagePullPolicyAlways    = "Always"
	prefetechVolumeName      = "pre-fetech-volume"
	defaultExecutorInstances = int32(1)
)

// k8s spark job plugin's configure
//
// EXECUTOR_K8SSPARK_K8SSPARKFORTERMINUS_ADDR=http://127.0.0.1:8080
// EXECUTOR_K8SSPARK_K8SSPARKFORTERMINUS_ENABLETAG=true
// EXECUTOR_K8SSPARK_K8SSPARKFORTERMINUS_SPARK_VERSION="2.4.0"
// EXECUTOR_K8SSPARK_K8SSPARKFORTERMINUS_BASICAUTH=
// EXECUTOR_K8SSPARK_K8SSPARKFORTERMINUS_CA_CRT=
// EXECUTOR_K8SSPARK_K8SSPARKFORTERMINUS_CLIENT_CRT=
// EXECUTOR_K8SSPARK_K8SSPARKFORTERMINUS_CLIENT_KEY=
// EXECUTOR_K8SSPARK_K8SSPARKFORTERMINUS_BEARER_TOKEN=
func init() {
	executortypes.Register(executorKind, func(name executortypes.Name, clusterName string, options map[string]string, optionsPlus interface{}) (
		executortypes.Executor, error) {
		addr, ok := options["ADDR"]
		if !ok {
			return nil, errors.Errorf("not found spark address in env variables")
		}

		if !strings.HasPrefix(addr, "inet://") {
			if !strings.HasPrefix(addr, "http") && !strings.HasPrefix(addr, "https") {
				addr = strutil.Concat("http://", addr)
			}
		}

		client := httpclient.New()
		if _, ok := options["CA_CRT"]; ok {
			logrus.Infof("k8s spark executor(%s) addr for https: %v", name, addr)
			client = httpclient.New(httpclient.WithHttpsCertFromJSON([]byte(options["CLIENT_CRT"]),
				[]byte(options["CLIENT_KEY"]),
				[]byte(options["CA_CRT"])))

			token, ok := options["BEARER_TOKEN"]
			if !ok {
				return nil, errors.Errorf("not found k8s bearer token")
			}
			// RBAC is enabled by default, and user authentication is required through token
			client.BearerTokenAuth(token)
		}

		enableTag, err := util.ParseEnableTagOption(options, "ENABLETAG", true)
		if err != nil {
			return nil, err
		}

		sparkVersion, ok := options["SPARK_VERSION"]
		if !ok {
			return nil, errors.Errorf("not found spark version in env variables")
		}

		clusterInfo, err := clusterinfo.New(clusterName, clusterinfo.WithCompleteParams(addr, client))
		if err != nil {
			return nil, errors.Errorf("failed to new cluster info, executorName: %s, clusterName: %s, (%v)",
				name, clusterName, err)
		}
		// Synchronize cluster info (every 10m)
		go clusterInfo.LoopLoadAndSync(context.Background(), false)

		return &k8sSpark{
			name:         name,
			clusterName:  clusterName,
			addr:         addr,
			options:      options,
			enableTag:    enableTag,
			sparkVersion: sparkVersion,
			client:       sparkapplication.New(sparkapplication.WithCompleteParams(addr, client)),
			pvc:          persistentvolumeclaim.New(persistentvolumeclaim.WithCompleteParams(addr, client)),
			namespace:    namespace.New(namespace.WithCompleteParams(addr, client)),
			secret:       secret.New(secret.WithCompleteParams(addr, client)),
			sa:           serviceaccount.New(serviceaccount.WithCompleteParams(addr, client)),
			role:         role.New(role.WithCompleteParams(addr, client)),
			rolebinding:  rolebinding.New(rolebinding.WithCompleteParams(addr, client)),
			clusterInfo:  clusterInfo,
		}, nil
	})
}

type k8sSpark struct {
	name         executortypes.Name
	clusterName  string
	addr         string
	options      map[string]string
	enableTag    bool   // Whether to enable label scheduling
	sparkVersion string // Spark deployment version
	client       *sparkapplication.SparkApplication
	pvc          *persistentvolumeclaim.PersistentVolumeClaim
	namespace    *namespace.Namespace
	secret       *secret.Secret
	sa           *serviceaccount.ServiceAccount
	role         *role.Role
	rolebinding  *rolebinding.RoleBinding
	clusterInfo  *clusterinfo.ClusterInfo
}

// Kind implements executortypes.Executor interface
func (s *k8sSpark) Kind() executortypes.Kind {
	return executorKind
}

// Name implements executortypes.Executor interface
func (s *k8sSpark) Name() executortypes.Name {
	return s.name
}

// Create implements creating servicegroup based on sparkapplication crd api
func (s *k8sSpark) Create(ctx context.Context, specObj interface{}) (interface{}, error) {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return nil, errors.New("invalid job spec")
	}

	logrus.Debugf("start to create k8s spark job, body: %+v", job)

	if err := s.prepareNamespaceResouce(job.Namespace); err != nil {
		return nil, err
	}

	if err := s.preparePVCForJob(&job); err != nil {
		return nil, err
	}

	app, err := s.generateKubeSparkJob(&job)
	if err != nil {
		return nil, errors.Errorf("failed to generate spark application, namespace: %s, name: %s, (%v)",
			job.Namespace, job.Name, err)
	}

	if err := s.client.Create(app); err != nil {
		return nil, err
	}

	logrus.Debugf("succeed to create spark application, namespace: %s, name: %s", job.Namespace, job.Name)
	job.Status = apistructs.StatusUnschedulable
	return job, nil
}

// Destroy implements deleting servicegroup based on sparkapplication crd api
func (s *k8sSpark) Destroy(ctx context.Context, specObj interface{}) error {
	return s.Remove(ctx, specObj)
}

// Status implements getting job status based on sparkapplication crd api
func (s *k8sSpark) Status(ctx context.Context, specObj interface{}) (apistructs.StatusDesc, error) {
	var status apistructs.StatusDesc

	job, ok := specObj.(apistructs.Job)
	if !ok {
		return status, errors.New("invalid job spec")
	}

	app, err := s.client.Get(job.Namespace, job.Name)
	if err != nil {
		errMsg := fmt.Sprintf("failed to get the status of k8s spark job, name: %s, (%v)", job.Name, err)
		logrus.Warningf(errMsg)

		if err == k8serror.ErrNotFound {
			status.Status = apistructs.StatusNotFoundInCluster
			return status, nil
		}

		return status, errors.New(errMsg)
	}

	status.LastMessage = app.Status.AppState.ErrorMessage
	switch app.Status.AppState.State {
	case sparkv1beta2.NewState, sparkv1beta2.SubmittedState:
		status.Status = apistructs.StatusUnschedulable
	case sparkv1beta2.RunningState:
		status.Status = apistructs.StatusRunning
	case sparkv1beta2.CompletedState, sparkv1beta2.SucceedingState:
		status.Status = apistructs.StatusStoppedOnOK
	case sparkv1beta2.FailedState, sparkv1beta2.FailingState, sparkv1beta2.FailedSubmissionState,
		sparkv1beta2.InvalidatingState, sparkv1beta2.PendingRerunState:
		status.Status = apistructs.StatusStoppedOnFailed
	case sparkv1beta2.UnknownState:
		status.Status = apistructs.StatusUnknown
	default:
		status.Status = apistructs.StatusUnknown
		status.LastMessage = fmt.Sprintf("unknown status, sparkAppState: %v", app.Status.AppState.State)
	}

	logrus.Debugf("succeed to get spark application status, namespace: %s, name: %s, status: %+v",
		job.Namespace, job.Name, status)

	return status, nil
}

// Remove implements removing job based on sparkapplication crd api
func (s *k8sSpark) Remove(ctx context.Context, specObj interface{}) error {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return errors.New("invalid job spec")
	}

	if job.Name == "" {
		return s.removePipelineJobs(job.Namespace)
	}

	if err := s.client.DeleteIfExists(job.Namespace, job.Name); err != nil {
		return errors.Errorf("failed to remove spark application, namespace: %s, name: %s, (%v)",
			job.Namespace, job.Name, err)
	}

	return nil
}

// Update implements update job based on sparkapplication crd api
func (s *k8sSpark) Update(ctx context.Context, specObj interface{}) (interface{}, error) {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return nil, errors.New("invalid job spec")
	}

	app, err := s.generateKubeSparkJob(&job)
	if err != nil {
		return nil, errors.Errorf("failed to generate spark application, namespace: %s, name: %s, (%v)",
			job.Namespace, job.Name, err)
	}

	if err := s.client.Update(app); err != nil {
		return nil, err
	}

	logrus.Debugf("succeed to update spark application, namespace: %s, name: %s", job.Namespace, job.Name)
	return nil, nil
}

// Inspect implements getting job info
func (s *k8sSpark) Inspect(ctx context.Context, specObj interface{}) (interface{}, error) {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return nil, errors.New("invalid job spec")
	}

	app, err := s.client.Get(job.Namespace, job.Name)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// Cancel implements canceling manipulating job
func (s *k8sSpark) Cancel(ctx context.Context, specObj interface{}) (interface{}, error) {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return nil, errors.New("invalid job spec")
	}

	if err := s.client.DeletePod(job.Namespace, job.Name); err != nil {
		return nil, errors.Errorf("failed to cancel spark application, namespace: %s, name: %s, (%v)",
			job.Namespace, job.Name, err)
	}

	logrus.Debugf("succeed to cancel spark application, namespace: %s, name: %s", job.Namespace, job.Name)
	return nil, nil
}
func (s *k8sSpark) Precheck(ctx context.Context, specObj interface{}) (apistructs.ServiceGroupPrecheckData, error) {
	return apistructs.ServiceGroupPrecheckData{Status: "ok"}, nil
}

// SetNodeLabels set the lables of node
func (s *k8sSpark) SetNodeLabels(setting executortypes.NodeLabelSetting, hosts []string, labels map[string]string) error {
	return errors.New("set node labels not implemented in K8SSpark")
}

func (s *k8sSpark) CapacityInfo() apistructs.CapacityInfoData {
	return apistructs.CapacityInfoData{}
}
func (s *k8sSpark) ResourceInfo(brief bool) (apistructs.ClusterResourceInfoData, error) {
	return apistructs.ClusterResourceInfoData{}, fmt.Errorf("resourceinfo not support for k8sspark")
}

func (s *k8sSpark) removePipelineJobs(ns string) error {
	return s.namespace.Delete(ns)
}
func (*k8sSpark) CleanUpBeforeDelete() {}
func (*k8sSpark) JobVolumeCreate(ctx context.Context, spec interface{}) (string, error) {
	return "", fmt.Errorf("not support for k8sspark")
}
func (*k8sSpark) KillPod(podname string) error {
	return fmt.Errorf("not support for k8sspark")
}
