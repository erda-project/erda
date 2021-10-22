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
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	credentialpb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/dbclient"
	"github.com/erda-project/erda/pkg/k8sclient"
)

// GetAccessKey get access key with cluster name.
func (c *Clusters) GetAccessKey(clusterName string) (*credentialpb.QueryAccessKeysResponse, error) {
	return c.credential.QueryAccessKeys(context.Background(), &credentialpb.QueryAccessKeysRequest{
		Status:      credentialpb.StatusEnum_ACTIVATE,
		SubjectType: credentialpb.SubjectTypeEnum_CLUSTER,
		Subject:     clusterName,
	})
}

// GetOrCreateAccessKey get or create access key
func (c *Clusters) GetOrCreateAccessKey(clusterName string) (*credentialpb.AccessKeysItem, error) {
	ak, err := c.GetAccessKey(clusterName)
	if err != nil {
		logrus.Errorf("get access key error when create precheck: %v", err)
		return nil, err
	}

	// if accessKey already exist, return the latest
	if ak.Total != 0 {
		return ak.Data[0], nil
	}

	// Check cluster exist or not
	if err = c.CheckCluster(clusterName); err != nil {
		return nil, err
	}

	// Create accessKey
	res, err := c.credential.CreateAccessKey(context.Background(), &credentialpb.CreateAccessKeyRequest{
		Subject:     clusterName,
		SubjectType: credentialpb.SubjectTypeEnum_CLUSTER,
		Scope:       apistructs.CMPClusterScope,
		ScopeId:     clusterName,
	})

	if err != nil {
		logrus.Errorf("create access key error: %v", err)
		return nil, err
	}

	return res.Data, nil
}

// GetOrCreateAccessKeyWithRecord get or create access key with record
func (c *Clusters) GetOrCreateAccessKeyWithRecord(clusterName, userID, orgID string) (*credentialpb.AccessKeysItem, error) {
	var (
		detailInfo string
		status     = dbclient.StatusTypeSuccess
	)

	res, err := c.GetOrCreateAccessKey(clusterName)
	if err != nil {
		detailInfo = err.Error()
		status = dbclient.StatusTypeFailed
	}

	// Record create credential success or error
	_, recordError := c.db.RecordsWriter().Create(&dbclient.Record{
		RecordType:  dbclient.RecordTypeCreateClusterCredential,
		UserID:      userID,
		OrgID:       orgID,
		ClusterName: clusterName,
		Status:      status,
		Detail:      detailInfo,
	})

	if recordError != nil {
		logrus.Errorf("recorde create cluster credential error: %v", recordError)
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

// DeleteAccessKey Delete access key
func (c *Clusters) DeleteAccessKey(clusterName string) error {
	res, err := c.GetAccessKey(clusterName)
	if err != nil {
		return err
	}

	// Doesn't need delete
	if res.Total == 0 {
		return nil
	}

	var errResp string

	// Clear all access key about clusterName
	for _, item := range res.Data {
		_, err = c.credential.DeleteAccessKey(context.Background(), &credentialpb.DeleteAccessKeyRequest{
			Id: item.Id,
		})
		if err != nil {
			logrus.Errorf("delete accesskey error: %v", err)
			errResp += err.Error()
		}
	}

	if len(errResp) != 0 {
		return fmt.Errorf("delete accessKey abount cluster %v error: %v", clusterName, err)
	}

	return nil
}

// ResetAccessKey reset access key
func (c *Clusters) ResetAccessKey(clusterName string) (*credentialpb.AccessKeysItem, error) {
	// Get configmap and PreCheck cluster connection
	kc, err := k8sclient.NewWithTimeOut(clusterName, getClusterTimeout)
	if err != nil {
		logrus.Errorf("get kubernetes client error when update accesskey, err: %v", err)
		tipErr := fmt.Errorf("connect to cluster: %s error: %v", clusterName, err)
		return nil, tipErr
	}

	// Get kubernetes clientSet
	cs := kc.ClientSet

	return c.ResetAccessKeyWithClientSet(clusterName, cs)
}

// ResetAccessKeyWithClientSet reset access key with specified clientSet
func (c *Clusters) ResetAccessKeyWithClientSet(clusterName string, cs *kubernetes.Clientset) (*credentialpb.AccessKeysItem, error) {
	// Get worker namespace
	workerNs := getWorkerNamespace()

	sec, err := cs.CoreV1().Secrets(workerNs).Get(context.Background(), apistructs.ErdaClusterCredential, v1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			tipErr := fmt.Errorf("get worker cluster config err: %v", err)
			logrus.Error(tipErr)
			return nil, tipErr
		}

		logrus.Info("cluster credential secret doesn't exist, create it.")
		sec, err = cs.CoreV1().Secrets(workerNs).Create(context.Background(), &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name: apistructs.ErdaClusterCredential,
			},
		}, v1.CreateOptions{})
		if err != nil {
			logrus.Errorf("create cluster credential secret error: %v", err)
			return nil, err
		}
	}

	// Delete access key
	if err = c.DeleteAccessKey(clusterName); err != nil {
		logrus.Errorf("delete access key error: %v", err)
		return nil, err
	}

	// Create new accessKey
	newAk, err := c.GetOrCreateAccessKey(clusterName)
	if err != nil {
		logrus.Errorf("create access key error: %v", err)
		return nil, err
	}

	if sec.Data == nil {
		sec.Data = make(map[string][]byte, 0)
	}

	// Update secret
	sec.Data[apistructs.ClusterAccessKey] = []byte(newAk.AccessKey)

	_, err = cs.CoreV1().Secrets(workerNs).Update(context.Background(), sec, v1.UpdateOptions{})
	if err != nil {
		logrus.Errorf("update worker cluster credential error: %v", err)
		return nil, err
	}

	return newAk, nil
}

// ResetAccessKeyWithRecord reset ak with record
func (c *Clusters) ResetAccessKeyWithRecord(clusterName, userID, orgID string) (*credentialpb.AccessKeysItem, error) {
	var (
		detailInfo string
		status     = dbclient.StatusTypeSuccess
	)

	res, err := c.ResetAccessKey(clusterName)
	if err != nil {
		detailInfo = err.Error()
		status = dbclient.StatusTypeFailed
	}

	// Record reset credential success or error
	_, recordError := c.db.RecordsWriter().Create(&dbclient.Record{
		RecordType:  dbclient.RecordTypeResetClusterCredential,
		UserID:      userID,
		OrgID:       orgID,
		ClusterName: clusterName,
		Status:      status,
		Detail:      detailInfo,
	})

	if recordError != nil {
		logrus.Errorf("recorde reset cluster credneital error: %v", recordError)
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

// CheckCluster check cluster
func (c *Clusters) CheckCluster(clusterName string) error {
	// check cluster exist or not
	_, err := c.bdl.GetCluster(clusterName)
	if err != nil {
		logrus.Errorf("check cluster error: %v", err)
		return err
	}
	return nil
}
