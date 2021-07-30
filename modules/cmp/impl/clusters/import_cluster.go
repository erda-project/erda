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

package clusters

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"os"
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
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httputil"
)

const (
	KubeconfigType     = "KUBECONFIG"
	SAType             = "SERVICEACCOUNT"
	ProxyType          = "PROXY"
	caKey              = "ca.crt"
	tokenKey           = "token"
	ModuleClusterInit  = "cluster-init"
	ModuleClusterAgent = "cluster-agent"
	ClusterAgentSA     = ModuleClusterAgent
	ClusterAgentCR     = "cluster-agent-cr"
	ClusterAgentCRB    = "cluster-agent-crb"
)

var (
	initRetryTimeout = 30 * time.Second
)

type RenderDeploy struct {
	ClusterName           string
	MasterClusterDomain   string // Master cluster domain, collector or openapi public
	MasterClusterProtocol string
	PlateFormVersion      string
	CustomDomain          string // Target cluster custom domain
	InitJobImage          string
	ClusterAgentImage     string
	ErdaHelmChartVersion  string
	DialerPublicAddr      string
	ErdaSystem            string
	OrgName               string
}

// importCluster import cluster
func (c *Clusters) importCluster(userID string, req *apistructs.ImportCluster) error {
	var err error

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

	status, err := c.getClusterStatus(ci)
	if err != nil {
		return err
	}

	if !(status == statusOffline || status == statusUnknown) {
		return nil
	}

	cs, err := c.k8s.GetInClusterClient()
	if err != nil {
		return err
	}

	if err = c.checkNamespace(); err != nil {
		return err
	}

	tc, err := c.k8s.GetClient(req.ClusterName)
	if err != nil {
		return err
	}

	orgDto, err := c.bdl.GetOrg(req.OrgID)
	if err != nil {
		return err
	}

	if _, err = tc.ClientSet.CoreV1().ServiceAccounts(conf.ErdaNamespace()).Get(context.Background(), ClusterAgentSA,
		metav1.GetOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			logrus.Errorf("get cluster-agent serviceAccount error: %v", err)
			return err
		}
		if _, err = tc.ClientSet.CoreV1().ServiceAccounts(conf.ErdaNamespace()).Create(context.Background(),
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

	if _, err = tc.ClientSet.RbacV1().ClusterRoles().Get(context.Background(), ClusterAgentCR,
		metav1.GetOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			logrus.Errorf("get cluster-agent cluster role error: %v", err)
			return err
		}
		allRole := []string{"*"}

		if _, err = tc.ClientSet.RbacV1().ClusterRoles().Create(context.Background(),
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

	if _, err = tc.ClientSet.RbacV1().ClusterRoleBindings().Get(context.Background(), ClusterAgentCRB,
		metav1.GetOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			logrus.Errorf("get cluster-agent cluster role binding error: %v", err)
			return err
		}

		if _, err = tc.ClientSet.RbacV1().ClusterRoleBindings().Create(context.Background(),
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
						Namespace: conf.ErdaNamespace(),
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

	nodes, err := tc.ClientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, node := range nodes.Items {
		node.Labels[fmt.Sprintf("dice/org-%s", orgDto.Name)] = "true"
		if _, err = tc.ClientSet.CoreV1().Nodes().Update(context.Background(), &node,
			metav1.UpdateOptions{}); err != nil {
			return err
		}
	}

	// check init job, if already exist, return
	if _, err = cs.BatchV1().Jobs(getPlatformNamespace()).Get(context.Background(), generateInitJobName(req.OrgID, req.ClusterName),
		metav1.GetOptions{}); err == nil {
		return nil
	}

	// create init job
	if _, err = cs.BatchV1().Jobs(getPlatformNamespace()).Create(context.Background(), c.generateClusterInitJob(req.OrgID, req.ClusterName, false),
		metav1.CreateOptions{}); err != nil {
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
	cs, err := c.k8s.GetInClusterClient()
	if err != nil {
		return err
	}

	if err = c.checkNamespace(); err != nil {
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
			if err = cs.BatchV1().Jobs(getPlatformNamespace()).Delete(context.Background(), generateInitJobName(orgID,
				req.ClusterName), metav1.DeleteOptions{
				PropagationPolicy: &propagationPolicy,
			}); err != nil {
				// if delete error is job not found, try again
				if !k8serrors.IsNotFound(err) {
					time.Sleep(500 * time.Millisecond)
					continue
				}
				// create job, if create error, tip retry again
				if _, err = cs.BatchV1().Jobs(getPlatformNamespace()).Create(context.Background(),
					c.generateClusterInitJob(orgID, req.ClusterName, true), metav1.CreateOptions{}); err != nil {
					return fmt.Errorf("create retry job error: %v, please try again", err)
				}
				return nil
			}
		}
	}
}

func (c *Clusters) checkNamespace() error {
	cs, err := c.k8s.GetInClusterClient()
	if err != nil {
		return err
	}

	// check namespace
	_, err = cs.CoreV1().Namespaces().Get(context.Background(), conf.ErdaNamespace(), metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		if _, err = cs.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: conf.ErdaNamespace(),
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

	masterCluster := os.Getenv(apistructs.MasterClusterKey)
	if masterCluster == "" {
		return "", fmt.Errorf("can't get master cluster info")
	}

	ci, err := c.bdl.QueryClusterInfo(masterCluster)
	if err != nil {
		return "", err
	}

	version := ci.Get("DICE_VERSION")
	masterClusterDomain := ci.Get("DICE_ROOT_DOMAIN")
	mcProtocol := parseClusterProtocol(ci.Get("DICE_PROTOCOL"))

	rd := RenderDeploy{
		ClusterName:           clusterName,
		MasterClusterDomain:   masterClusterDomain,
		MasterClusterProtocol: mcProtocol,
		PlateFormVersion:      version,
		CustomDomain:          cluster.WildcardDomain,
		InitJobImage:          renderReleaseImageAddr(ModuleClusterInit, version),
		ClusterAgentImage:     renderReleaseImageAddr(ModuleClusterAgent, version),
		ErdaHelmChartVersion:  conf.ErdaHelmChartVersion(),
		DialerPublicAddr:      conf.DialerPublicAddr(),
		ErdaSystem:            conf.ErdaNamespace(),
		OrgName:               orgName,
	}

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
func (c *Clusters) generateClusterInitJob(orgID uint64, clusterName string, reInstall bool) *batchv1.Job {
	jobName := generateInitJobName(orgID, clusterName)
	var backOffLimit int32

	compClusterName := os.Getenv(apistructs.MasterClusterKey)
	if compClusterName == "" {
		return nil
	}

	cci, err := c.bdl.QueryClusterInfo(compClusterName)
	if err != nil {
		return nil
	}

	eci, err := c.bdl.GetCluster(clusterName)
	if err != nil {
		return nil
	}

	platformDomain := cci.Get("DICE_ROOT_DOMAIN")
	platformVersion := cci.Get("DICE_VERSION")
	mcProtocol := parseClusterProtocol(cci.Get("DICE_PROTOCOL"))

	envs := []corev1.EnvVar{
		{
			Name:  "ERDA_CHART_VERSION",
			Value: conf.ErdaHelmChartVersion(),
		},
		{
			Name:  "TARGET_CLUSTER",
			Value: clusterName,
		},
		{
			Name:  "INSTALL_MODE",
			Value: "remote",
		},
		{
			Name:  "REPO_MODE",
			Value: "local",
		},
		{
			Name:  "HELM_NAMESPACE",
			Value: conf.ErdaNamespace(),
		},
		{
			Name: "CHART_ERDA_BASE_VALUES",
			Value: fmt.Sprintf("configmap.clustername=%s,configmap.domain=%s",
				clusterName, eci.WildcardDomain),
		},
		{
			Name:  "CHART_ERDA_ADDONS_VALUES",
			Value: "registry.networkMode=''",
		},
		{
			Name: "CHART_ERDA_VALUES",
			Value: fmt.Sprintf("domain=%s,clusterName=%s,masterCluster.domain=%s,masterCluster.protocol=%s",
				eci.WildcardDomain, clusterName, platformDomain, mcProtocol),
		},
		{
			Name:  "CLUSTER_MANAGER_ADDR",
			Value: discover.ClusterManager(),
		},
		{
			Name:  "REINSTALL",
			Value: strconv.FormatBool(reInstall),
		},
	}

	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: getPlatformNamespace(),
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backOffLimit,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						{
							Name:            jobName,
							Image:           renderReleaseImageAddr(ModuleClusterInit, platformVersion),
							ImagePullPolicy: "Always",
							Command:         []string{"sh", "-c", fmt.Sprintf("/app/%s", ModuleClusterInit)},
							Env:             envs,
						},
					},
				},
			},
		},
	}
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

func parseClusterProtocol(protocol string) string {
	var (
		protocolHttp  = "http"
		protocolHttps = "https"
	)

	if strings.Contains(strings.ToLower(protocol), protocolHttps) {
		return protocolHttps
	}

	return protocolHttp
}

// renderReleaseImageAddr render release image with module name and version
// e.g. registry.erda.cloud/erda:v1.1
func renderReleaseImageAddr(module string, version string) string {
	return fmt.Sprintf("%s/%s:v%s", conf.ReleaseRepo(), module, version)
}

// generateInitJobName generate init job name with orgID and clusterName
func generateInitJobName(orgID uint64, clusterName string) string {
	return fmt.Sprintf("erda-cluster-init-%d-%s", orgID, clusterName)
}
