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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/cluster-agent/config"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/k8sclient"
)

const (
	caCrtKey    = "ca.crt"
	tokenSecKey = "token"
)

var ServiceAccountTokenNotReady = errors.New("service account token not ready")

var defaultRetry = wait.Backoff{
	Steps:    5,
	Duration: 200 * time.Millisecond,
	Factor:   2.0,
	Jitter:   0.1,
}

type KubernetesClusterInfo struct {
	Address string `json:"address"`
	Token   string `json:"token"`
	CACert  string `json:"caCert"`
}

type Option func(*Client)

type Client struct {
	sync.Mutex
	cfg        *config.Config
	accessKey  string
	connected  bool
	disconnect chan struct{}
}

func New(ops ...Option) *Client {
	c := Client{
		disconnect: make(chan struct{}),
	}
	for _, op := range ops {
		op(&c)
	}
	return &c
}

func WithConfig(cfg *config.Config) Option {
	return func(c *Client) {
		c.cfg = cfg
	}
}

func (c *Client) DisConnect() {
	if !c.IsConnected() {
		return
	}
	c.disconnect <- struct{}{}
}

func (c *Client) Start(ctx context.Context) error {
	headers := http.Header{
		"X-Erda-Cluster-Key": {c.cfg.ClusterKey},
	}

	if c.cfg.CollectClusterInfo {
		clusterInfo, err := c.loadClusterInfo(ctx)
		if err != nil {
			return err
		}
		bytes, err := json.Marshal(clusterInfo)
		if err != nil {
			return err
		}
		headers["X-Erda-Cluster-Info"] = []string{base64.StdEncoding.EncodeToString(bytes)}
	}

	ep, err := parseDialerEndpoint(c.cfg.ClusterManagerEndpoint)
	if err != nil {
		logrus.Errorf("failed to parse dial endpoint: %v", err)
		return err
	}

	// If specified cluster access key, preferred to use it.
	if c.cfg.ClusterAccessKey == "" {
		go func() {
			if err := c.watchClusterCredential(ctx); err != nil {
				logrus.Errorf("watch cluster info error: %v", err)
				return
			}
		}()
	} else {
		// Set access key values default
		c.setAccessKey(c.cfg.ClusterAccessKey)
		logrus.Infof("use specified cluster access key: %v", c.cfg.ClusterAccessKey)
	}

	for {
		if c.getAccessKey() == "" {
			continue
		}

		headers.Set("Authorization", c.getAccessKey())
		_ = remotedialer.ClientConnect(ctx, ep, headers, nil, func(proto, address string) bool {
			switch proto {
			case "tcp":
				return true
			case "unix":
				return address == "/var/run/docker.sock"
			case "npipe":
				return address == "//./pipe/docker_engine"
			}
			return false
		}, c.onConnect)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(c.cfg.ConRetryInterval) * time.Second):
		}
	}
}

// onConnect
func (c *Client) onConnect(ctx context.Context, _ *remotedialer.Session) error {
	defer func() {
		c.setConnected(false)
	}()

	c.setConnected(true)

	// Or passThrough cancel() function
	select {
	case <-c.disconnect:
		return fmt.Errorf("cluster credential reload")
	case <-ctx.Done():
		return nil
	}
}

func (c *Client) setConnected(b bool) {
	c.Lock()
	defer c.Unlock()
	c.connected = b
}

func (c *Client) IsConnected() bool {
	// data race
	c.Lock()
	defer c.Unlock()
	return c.connected
}

func (c *Client) loadClusterInfo(ctx context.Context) (*KubernetesClusterInfo, error) {
	clusterInfo := &KubernetesClusterInfo{
		Address: c.cfg.K8SApiServerAddr,
	}

	k, err := k8sclient.NewForInCluster()
	if err != nil {
		return nil, err
	}

	ns := c.cfg.ErdaNamespace
	serviceAccountName := c.cfg.ServiceAccount
	tokenSecretName := c.cfg.ServiceAccountTokenSecret

	// Retrieve the service account
	serviceAccount, err := k.ClientSet.CoreV1().ServiceAccounts(ns).Get(ctx, serviceAccountName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Determine which secret to use based on available secrets in the service account
	var secret *corev1.Secret
	if len(serviceAccount.Secrets) != 0 {
		secretName := serviceAccount.Secrets[0].Name
		logrus.Infof("Loading auth info from secret %s", secretName)
		secret, err = k.ClientSet.CoreV1().Secrets(ns).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	} else {
		// Create or retrieve the service account token secret
		secret, err = k.ClientSet.CoreV1().Secrets(ns).Get(ctx, tokenSecretName, metav1.GetOptions{})
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				return nil, err
			}
			// Create the token secret if it does not exist
			_, err = k.ClientSet.CoreV1().Secrets(ns).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: tokenSecretName,
					Annotations: map[string]string{
						corev1.ServiceAccountNameKey: serviceAccountName,
					},
				},
				Type: corev1.SecretTypeServiceAccountToken,
			}, metav1.CreateOptions{})
			if err != nil {
				return nil, err
			}

			if err = retry.OnError(defaultRetry, func(err error) bool {
				return errors.Is(err, ServiceAccountTokenNotReady)
			}, func() error {
				gotSecret, err := k.ClientSet.CoreV1().Secrets(ns).Get(ctx, tokenSecretName, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if len(gotSecret.Data) != 0 {
					secret = gotSecret
					return nil
				}
				return ServiceAccountTokenNotReady
			}); err != nil {
				return nil, err
			}
		}
	}

	logrus.Infof("gonna to load data from secret %s", secret.Name)

	// Retrieve and process CA certificate
	caData, ok := secret.Data[caCrtKey]
	if !ok {
		return nil, fmt.Errorf("failed to load CA data from secret %s", secret.Name)
	}
	clusterInfo.CACert = base64.StdEncoding.EncodeToString(caData)

	// Retrieve and process token
	token, ok := secret.Data[tokenSecKey]
	if !ok {
		return nil, fmt.Errorf("failed to load token from secret %s", secret.Name)
	}
	clusterInfo.Token = strings.TrimSpace(string(token))

	logrus.Debugf("loaded cluster info: %#+v", clusterInfo)
	return clusterInfo, nil
}

func parseDialerEndpoint(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	//inCluster, visit dialer inner service first.
	if os.Getenv(string(apistructs.DICE_IS_EDGE)) == "false" && discover.ClusterDialer() != "" {
		return "ws://" + discover.ClusterDialer() + u.Path, nil
	}

	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	}

	return u.String(), nil
}
