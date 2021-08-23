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

package clusters

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	decoder "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/conf"
	"github.com/erda-project/erda/modules/cmp/dbclient"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/k8sclient"
)

const (
	KubeconfigType   = "KUBECONFIG"
	SAType           = "SERVICEACCOUNT"
	ProxyType        = "PROXY"
	caKey            = "ca.crt"
	tokenKey         = "token"
	ModuleClusterOps = "cluster-ops"
	ClusterAgentSA   = "cluster-agent"
	ClusterAgentCR   = "cluster-agent-cr"
	ClusterAgentCRB  = "cluster-agent-crb"
)

var (
	initRetryTimeout  = 30 * time.Second
	getClusterTimeout = 2 * time.Second
)

type RenderDeploy struct {
	ErdaNamespace string
	JobImage      string
	Envs          []corev1.EnvVar
}

// importCluster import cluster
func (c *Clusters) importCluster(userID string, req *apistructs.ImportCluster) error {
	mc, err := ParseManageConfigFromCredential(req.CredentialType, req.Credential)
	if err != nil {
		return err
	}

	// TODO: support tag switch, current force true
	// e.g. modules/scheduler/impl/cluster/hook.go line:136
	req.ScheduleConfig.EnableTag = true

	// create cluster request to cluster-manager and core-service
	if err = c.bdl.CreateClusterWithOrg(userID, req.OrgID, &apistructs.ClusterCreateRequest{
		OrgID:           int64(req.OrgID),
		Name:            req.ClusterName,
		DisplayName:     req.DisplayName,
		Description:     req.Description,
		Type:            req.ClusterType,
		SchedulerConfig: &req.ScheduleConfig,
		WildcardDomain:  req.WildcardDomain,
		ManageConfig:    mc,
	}, http.Header{httputil.InternalHeader: []string{"cmp"}}); err != nil {
		return err
	}

	if mc.Type == apistructs.ManageProxy {
		return nil
	}

	ci, err := c.bdl.GetCluster(req.ClusterName)
	if err != nil {
		return err
	}

	kc, err := k8sclient.NewWithTimeOut(req.ClusterName, getClusterTimeout)
	if err != nil {
		logrus.Errorf("get kubernetes client error, clusterName: [%s]", req.ClusterName)
		return err
	}

	status, err := c.getClusterStatus(kc, ci)
	if err != nil {
		logrus.Errorf("get cluster status error: %v", err)
		return err
	}

	if req.ClusterName == conf.ErdaClusterName() || !(status == statusOffline || status == statusUnknown) {
		return nil
	}

	workNs := getWorkerNamespace()

	// check resource before execute cluster init job
	if err = c.importPreCheck(kc, workNs); err != nil {
		return err
	}

	// check init job, if already exist, return
	if _, err = kc.ClientSet.BatchV1().Jobs(workNs).Get(context.Background(),
		generateInitJobName(req.OrgID, req.ClusterName), metav1.GetOptions{}); err == nil {
		return nil
	}

	// create init job
	initJob, err := c.generateClusterInitJob(req.OrgID, req.ClusterName, false)
	if err != nil {
		logrus.Errorf("generate cluster init job error: %v", err)
		return err
	}

	if _, err = kc.ClientSet.BatchV1().Jobs(workNs).Create(context.Background(), initJob,
		metav1.CreateOptions{}); err != nil {
		logrus.Errorf("create cluster init job error: %v", err)
		return err
	}

	return err
}

// ImportClusterWithRecord import cluster with record
func (c *Clusters) ImportClusterWithRecord(userID string, req *apistructs.ImportCluster) error {
	var (
		err        error
		detailInfo string
		status     = dbclient.StatusTypeSuccess
	)

	if err = c.importCluster(userID, req); err != nil {
		detailInfo = err.Error()
		status = dbclient.StatusTypeFailed
	}

	_, recordError := c.db.RecordsWriter().Create(&dbclient.Record{
		RecordType:  dbclient.RecordTypeImportKubernetesCluster,
		UserID:      userID,
		OrgID:       strconv.FormatUint(req.OrgID, 10),
		ClusterName: req.ClusterName,
		Status:      status,
		Detail:      detailInfo,
	})

	logrus.Errorf("recorde import cluster error: %v", recordError)

	return err
}

func (c *Clusters) ClusterInitRetry(orgID uint64, req *apistructs.ClusterInitRetry) error {
	cs, err := k8sclient.New(req.ClusterName)
	if err != nil {
		return err
	}

	logrus.Infof("start retry init cluster %s", req.ClusterName)

	workNs := getWorkerNamespace()
	if err = c.importPreCheck(cs, workNs); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), initRetryTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry init job timeout, please try again")
		default:
			// delete old init job
			propagationPolicy := metav1.DeletePropagationBackground
			if err = cs.ClientSet.BatchV1().Jobs(workNs).Delete(context.Background(), generateInitJobName(orgID,
				req.ClusterName), metav1.DeleteOptions{
				PropagationPolicy: &propagationPolicy,
			}); err != nil {
				// if delete error is job not found, try again
				if !k8serrors.IsNotFound(err) {
					time.Sleep(500 * time.Millisecond)
					continue
				}
				// generate init job
				initJob, err := c.generateClusterInitJob(orgID, req.ClusterName, true)
				if err != nil {
					logrus.Errorf("generate retry cluster init job error: %v", err)
					continue
				}
				// create job, if create error, tip retry again
				if _, err = cs.ClientSet.BatchV1().Jobs(workNs).Create(context.Background(),
					initJob, metav1.CreateOptions{}); err != nil {
					return fmt.Errorf("create retry job error: %v, please try again", err)
				}
				return nil
			}
		}
	}
}

// importPreCheck check before import cluster
func (c *Clusters) importPreCheck(kc *k8sclient.K8sClient, ns string) error {
	if kc == nil || kc.ClientSet == nil {
		return fmt.Errorf("import cluster precheck error, kuberentes client is nil")
	}

	// check namespace, create if not exist
	if _, err := kc.ClientSet.CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		if _, err = kc.ClientSet.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	if _, err := kc.ClientSet.CoreV1().ServiceAccounts(ns).Get(context.Background(), ClusterAgentSA,
		metav1.GetOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			logrus.Errorf("get cluster-agent serviceAccount error: %v", err)
			return err
		}
		logrus.Infof("service account %s doesn't exist, create it", ClusterAgentSA)
		if _, err = kc.ClientSet.CoreV1().ServiceAccounts(ns).Create(context.Background(),
			&corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ServiceAccount",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ClusterAgentSA,
					Namespace: conf.ErdaNamespace(),
				},
			}, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	if _, err := kc.ClientSet.RbacV1().ClusterRoles().Get(context.Background(), ClusterAgentCR,
		metav1.GetOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			logrus.Errorf("get cluster-agent cluster role error: %v", err)
			return err
		}
		logrus.Infof("cluster role %s doesn't exist, create it", ClusterAgentCR)

		allRole := []string{"*"}

		if _, err = kc.ClientSet.RbacV1().ClusterRoles().Create(context.Background(),
			&rbacv1.ClusterRole{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "rbac.authorization.k8s.io/v1",
					Kind:       "ClusterRole",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: ClusterAgentCR,
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     allRole,
						APIGroups: allRole,
						Resources: allRole,
					},
					{
						Verbs:           allRole,
						NonResourceURLs: allRole,
					},
				},
			}, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	if _, err := kc.ClientSet.RbacV1().ClusterRoleBindings().Get(context.Background(), ClusterAgentCRB,
		metav1.GetOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			logrus.Errorf("get cluster-agent cluster role binding error: %v", err)
			return err
		}
		logrus.Infof("cluster role binding %s doesn't exist, create it", ClusterAgentCRB)

		if _, err = kc.ClientSet.RbacV1().ClusterRoleBindings().Create(context.Background(),
			&rbacv1.ClusterRoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "rbac.authorization.k8s.io/v1",
					Kind:       "ClusterRoleBinding",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: ClusterAgentCRB,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      ClusterAgentSA,
						Namespace: ns,
					},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     ClusterAgentCR,
				},
			}, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	return nil
}

// ParseKubeconfig parse kubeconfig to manage config
func ParseKubeconfig(kubeconfig []byte) (*apistructs.ManageConfig, error) {
	var mc apistructs.ManageConfig

	config, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return nil, err
	}

	// TODO: front support multi contexts select
	if len(config.Contexts) != 1 {
		return nil, fmt.Errorf("please provide specified cluster context")
	}

	var clusterCtx *api.Context

	for _, clusterCtx = range config.Contexts {
		break
	}

	if clusterCtx == nil {
		return nil, fmt.Errorf("get context error")
	}

	cluster := config.Clusters[clusterCtx.Cluster]
	if len(cluster.Server) == 0 {
		return nil, fmt.Errorf("cluster server address it empty")
	}

	mc.Address = cluster.Server

	if len(cluster.CertificateAuthorityData) != 0 {
		mc.CaData = base64.StdEncoding.EncodeToString(cluster.CertificateAuthorityData)
	}

	authInfo := config.AuthInfos[clusterCtx.AuthInfo]

	if len(authInfo.ClientKeyData) != 0 && len(authInfo.ClientCertificateData) != 0 {
		mc.Type = apistructs.ManageCert
		mc.CertData = base64.StdEncoding.EncodeToString(authInfo.ClientCertificateData)
		mc.KeyData = base64.StdEncoding.EncodeToString(authInfo.ClientKeyData)
		return &mc, nil
	}

	if len(authInfo.Token) != 0 {
		mc.Type = apistructs.ManageToken
		mc.Token = authInfo.Token
	}

	// TODO: support username and password

	return nil, fmt.Errorf("illegal kubeconfig")
}

func (c *Clusters) RenderInitCmd(orgName, clusterName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var (
		cluster *apistructs.ClusterInfo
		err     error
	)

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("get cluster init command timeout")
		default:
			cluster, err = c.bdl.GetCluster(clusterName)
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if cluster.ManageConfig.Type != apistructs.ManageProxy {
				return "", fmt.Errorf("only support proxy manage type")
			}

			if cluster.ManageConfig.Token != "" || cluster.ManageConfig.Address != "" {
				return fmt.Sprintf("cluster %s already registered", clusterName), nil
			}

			cmd := fmt.Sprintf("kubectl apply -f '$REQUEST_PREFIX?orgName=%s&clusterName=%s&accessKey=%s'", orgName,
				clusterName, cluster.ManageConfig.AccessKey)
			return cmd, nil
		}
	}
}

func (c *Clusters) RenderInitContent(orgName, clusterName string, accessKey string) (string, error) {
	cluster, err := c.bdl.GetCluster(clusterName)
	if err != nil {
		return "", fmt.Errorf("get cluster error:%v", err)
	}
	if cluster.ManageConfig == nil {
		return "", fmt.Errorf("manage config is nil")
	}

	if cluster.ManageConfig.AccessKey != accessKey {
		return "", fmt.Errorf("accesskey is error")
	}

	rd, err := c.renderCommonDeployConfig(orgName, clusterName)

	tmpl := template.Must(template.New("render").Parse(ProxyDeployTemplate))
	buf := new(bytes.Buffer)

	if err = tmpl.Execute(buf, rd); err != nil {
		return "", err
	}

	return buf.String(), nil

}

func ParseSecretes(address string, secret []byte) (*apistructs.ManageConfig, error) {
	var mc apistructs.ManageConfig

	if address != "" {
		mc.Address = address
	} else {
		return nil, fmt.Errorf("please provide non-empty address")
	}

	s := corev1.Secret{}

	reader := bytes.NewReader(secret)

	if err := decoder.NewYAMLOrJSONDecoder(reader, 1024).Decode(&s); err != nil {
		return nil, fmt.Errorf("illegal secret format: %v", err)
	}

	if _, ok := s.Data[caKey]; ok {
		mc.CaData = base64.StdEncoding.EncodeToString(s.Data[caKey])
	}

	if token, ok := s.Data[tokenKey]; !ok || len(s.Data[tokenKey]) == 0 {
		return nil, fmt.Errorf("secrets must provide token")
	} else {
		mc.Token = string(token)
		mc.Type = apistructs.ManageToken
		return &mc, nil
	}
}

func ParseManageConfigFromCredential(credentialType string, credential apistructs.ICCredential) (*apistructs.ManageConfig, error) {
	mc := &apistructs.ManageConfig{
		CredentialSource: credentialType,
	}

	res, err := base64.StdEncoding.DecodeString(credential.Content)
	if err != nil {
		return nil, fmt.Errorf("decode credntial content error: %v", err)
	}

	switch strings.ToUpper(credentialType) {
	case KubeconfigType:
		mc, err = ParseKubeconfig(res)
		if err != nil {
			return nil, fmt.Errorf("parse kubeconfig error: %v", err)
		}
	case SAType:
		mc, err = ParseSecretes(credential.Address, res)
		if err != nil {
			return nil, fmt.Errorf("parse serviceAccount credntial info error: %v", err)
		}
	case ProxyType:
		mc.Type = apistructs.ManageProxy
		mc.AccessKey = generateAccessKey(64)
	default:
		return nil, fmt.Errorf("credntial type '%s' is not support", credentialType)
	}

	mc.CredentialSource = credentialType

	return mc, nil
}

// generateClusterInitJob generate cluster init job
func (c *Clusters) generateClusterInitJob(orgID uint64, clusterName string, reInstall bool) (*batchv1.Job, error) {
	var (
		backOffLimit int32
		jobName      = generateInitJobName(orgID, clusterName)
	)

	orgDto, err := c.bdl.GetOrg(orgID)
	if err != nil {
		return nil, err
	}

	rd, err := c.renderCommonDeployConfig(orgDto.Name, clusterName)
	if err != nil {
		return nil, err
	}

	if reInstall {
		rd.Envs = append(rd.Envs, corev1.EnvVar{
			Name:  "REINSTALL",
			Value: "true",
		})
	}

	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: getWorkerNamespace(),
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backOffLimit,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: ClusterAgentSA,
					RestartPolicy:      "Never",
					Containers: []corev1.Container{
						{
							Name:            jobName,
							Image:           renderReleaseImageAddr(),
							ImagePullPolicy: "Always",
							Command:         []string{"sh", "-c", fmt.Sprintf("/app/%s", ModuleClusterOps)},
							Env:             rd.Envs,
						},
					},
				},
			},
		},
	}, nil
}

// renderCommonDeployConfig render deploy struct with common config
func (c *Clusters) renderCommonDeployConfig(orgName, clusterName string) (*RenderDeploy, error) {
	ci, err := c.bdl.GetCluster(clusterName)
	if err != nil {
		logrus.Errorf("render deploy config error: %v", err)
		return nil, err
	}

	rd := RenderDeploy{
		ErdaNamespace: getWorkerNamespace(),
		JobImage:      renderReleaseImageAddr(),
		Envs: []corev1.EnvVar{
			{Name: "DEBUG", Value: "true"},
			{Name: "ERDA_CHART_VERSION", Value: conf.ErdaVersion()},
			{Name: "HELM_NAMESPACE", Value: getWorkerNamespace()},
			{Name: "NODE_LABELS", Value: fmt.Sprintf("dice/org-%s=true", orgName)},
			{Name: "ERDA_CHART_VALUES", Value: generateSetValues(ci)},
			{Name: "HELM_REPO_URL", Value: conf.HelmRepoURL()},
			{Name: "HELM_REPO_USERNAME", Value: conf.HelmRepoUsername()},
			{Name: "HELM_REPO_PASSWORD", Value: conf.HelmRepoPassword()},
		},
	}

	return &rd, nil
}

// generateSetValues generate the values of helm chart install set
func generateSetValues(ci *apistructs.ClusterInfo) string {
	// current cluster type in database is k8s, dice-cluster-info need kubernetes
	if ci.Type == "k8s" {
		ci.Type = "kubernetes"
	}
	return "tags.work=true,tags.master=false," +
		fmt.Sprintf("global.domain=%s,erda.clusterName=%s,", ci.WildcardDomain, ci.Name) +
		fmt.Sprintf("erda.clusterConfig.clusterType=%s,", strings.ToLower(ci.Type)) +
		fmt.Sprintf("erda.masterCluster.domain=%s,erda.masterCluster.protocol=%s", conf.ErdaDomain(), conf.ErdaProtocol())
}

// generateAccessKey generate accessKey
func generateAccessKey(customLen int) string {
	res := make([]string, customLen)
	rand.Seed(time.Now().UnixNano())
	for start := 0; start < customLen; start++ {
		switch rand.Intn(3) {
		// rand lower
		case 0:
			randInt := rand.Intn(26) + 65
			res = append(res, string(rune(randInt)))
		// rand upper
		case 1:
			randInt := rand.Intn(26) + 97
			res = append(res, string(rune(randInt)))
		// rand number
		case 2:
			randInt := rand.Intn(10)
			res = append(res, strconv.Itoa(randInt))
		}
	}

	return strings.Join(res, "")
}

// renderReleaseImageAddr render release image with module name and version
// e.g. registry.erda.cloud/erda/cluster-init:v0.1
func renderReleaseImageAddr() string {
	return fmt.Sprintf("%s/%s:v%s", conf.ReleaseRegistry(), ModuleClusterOps, conf.ClusterInitVersion())
}

// generateInitJobName generate init job name with orgID and clusterName
func generateInitJobName(orgID uint64, clusterName string) string {
	return fmt.Sprintf("erda-cluster-init-%d-%s", orgID, clusterName)
}

// getWorkerNamespace get work node namespace
func getWorkerNamespace() string {
	// TODO: support different namespace of master and slave
	return conf.ErdaNamespace()
}
