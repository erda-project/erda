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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/erda-project/erda/internal/tools/cluster-agent/config"
	"github.com/erda-project/erda/pkg/k8sclient"
)

func TestLoadClusterInfo_UsesReferencedServiceAccountSecret(t *testing.T) {
	c := newLoadClusterInfoTestClient(fakeclientset.NewSimpleClientset(
		newTestServiceAccount("cluster-agent", "cluster-agent-token-mvp6d"),
		newSecret("cluster-agent-token-mvp6d", map[string][]byte{
			caCrtKey:    []byte("fake ca data"),
			tokenSecKey: []byte("fake token data\n"),
		}),
	))

	clusterInfo, err := c.loadClusterInfo(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "https://kubernetes.default.svc", clusterInfo.Address)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("fake ca data")), clusterInfo.CACert)
	assert.Equal(t, "fake token data", clusterInfo.Token)
}

func TestLoadClusterInfo_UsesExistingTokenSecretWhenServiceAccountHasNoSecrets(t *testing.T) {
	c := newLoadClusterInfoTestClient(fakeclientset.NewSimpleClientset(
		newTestServiceAccount("cluster-agent"),
		newSecret("cluster-agent-token", map[string][]byte{
			caCrtKey:    []byte("existing ca"),
			tokenSecKey: []byte("existing token"),
		}),
	))

	clusterInfo, err := c.loadClusterInfo(context.Background())
	require.NoError(t, err)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("existing ca")), clusterInfo.CACert)
	assert.Equal(t, "existing token", clusterInfo.Token)
}

func TestLoadClusterInfo_CreatesTokenSecretAndRetriesUntilReady(t *testing.T) {
	clientSet := fakeclientset.NewSimpleClientset(newTestServiceAccount("cluster-agent"))

	var (
		createCount int
		getCount    int
	)

	clientSet.PrependReactor("create", "secrets", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		createCount++
		create := action.(clientgotesting.CreateAction)
		secret := create.GetObject().(*corev1.Secret).DeepCopy()
		return true, secret, nil
	})
	clientSet.PrependReactor("get", "secrets", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		get := action.(clientgotesting.GetAction)
		if get.GetName() != "cluster-agent-token" {
			return false, nil, nil
		}

		getCount++
		switch getCount {
		case 1:
			return true, nil, k8serrors.NewNotFound(schema.GroupResource{Resource: "secrets"}, get.GetName())
		case 2:
			return true, newSecret("cluster-agent-token", nil), nil
		default:
			return true, newSecret("cluster-agent-token", map[string][]byte{
				caCrtKey:    []byte("ready ca"),
				tokenSecKey: []byte("ready token"),
			}), nil
		}
	})

	c := newLoadClusterInfoTestClient(clientSet)
	clusterInfo, err := c.loadClusterInfo(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, createCount)
	assert.GreaterOrEqual(t, getCount, 3)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("ready ca")), clusterInfo.CACert)
	assert.Equal(t, "ready token", clusterInfo.Token)
}

func TestLoadClusterInfo_MissingRequiredSecretData(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string][]byte
		wantErr string
	}{
		{
			name: "missing ca data",
			data: map[string][]byte{
				tokenSecKey: []byte("token"),
			},
			wantErr: "failed to load CA data from secret cluster-agent-token-mvp6d",
		},
		{
			name: "missing token",
			data: map[string][]byte{
				caCrtKey: []byte("ca"),
			},
			wantErr: "failed to load token from secret cluster-agent-token-mvp6d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newLoadClusterInfoTestClient(fakeclientset.NewSimpleClientset(
				newTestServiceAccount("cluster-agent", "cluster-agent-token-mvp6d"),
				newSecret("cluster-agent-token-mvp6d", tt.data),
			))

			_, err := c.loadClusterInfo(context.Background())
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestLoadClusterInfo_ReturnsInClusterClientError(t *testing.T) {
	c := New(WithConfig(&config.Config{
		CollectClusterInfo:        true,
		ErdaNamespace:             metav1.NamespaceDefault,
		K8SApiServerAddr:          "https://kubernetes.default.svc",
		ServiceAccount:            "cluster-agent",
		ServiceAccountTokenSecret: "cluster-agent-token",
	}))
	c.newInClusterClient = func(...k8sclient.Option) (*k8sclient.K8sClient, error) {
		return nil, errors.New("new in-cluster client failed")
	}

	_, err := c.loadClusterInfo(context.Background())
	require.EqualError(t, err, "new in-cluster client failed")
}

func TestLoadClusterInfo_ReturnsRetryErrorWhenTokenSecretNeverReady(t *testing.T) {
	originalRetry := defaultRetry
	defaultRetry = wait.Backoff{
		Steps:    2,
		Duration: time.Millisecond,
		Factor:   1,
		Jitter:   0,
	}
	t.Cleanup(func() {
		defaultRetry = originalRetry
	})

	clientSet := fakeclientset.NewSimpleClientset(newTestServiceAccount("cluster-agent"))
	getCount := 0
	clientSet.PrependReactor("get", "secrets", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		get := action.(clientgotesting.GetAction)
		if get.GetName() != "cluster-agent-token" {
			return false, nil, nil
		}
		getCount++
		if getCount == 1 {
			return true, nil, k8serrors.NewNotFound(schema.GroupResource{Resource: "secrets"}, get.GetName())
		}
		return true, newSecret("cluster-agent-token", nil), nil
	})
	clientSet.PrependReactor("create", "secrets", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		create := action.(clientgotesting.CreateAction)
		secret := create.GetObject().(*corev1.Secret).DeepCopy()
		return true, secret, nil
	})

	c := newLoadClusterInfoTestClient(clientSet)
	_, err := c.loadClusterInfo(context.Background())
	require.ErrorIs(t, err, ServiceAccountTokenNotReady)
}

func TestLoadClusterInfo_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*fakeclientset.Clientset)
		wantErr string
	}{
		{
			name: "service account get error",
			prepare: func(clientSet *fakeclientset.Clientset) {
				clientSet.PrependReactor("get", "serviceaccounts", func(clientgotesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("service account get failed")
				})
			},
			wantErr: "service account get failed",
		},
		{
			name: "referenced secret get error",
			prepare: func(clientSet *fakeclientset.Clientset) {
				clientSet.PrependReactor("get", "secrets", func(action clientgotesting.Action) (bool, runtime.Object, error) {
					get := action.(clientgotesting.GetAction)
					if get.GetName() != "cluster-agent-token-mvp6d" {
						return false, nil, nil
					}
					return true, nil, errors.New("referenced secret get failed")
				})
			},
			wantErr: "referenced secret get failed",
		},
		{
			name: "existing token secret get error",
			prepare: func(clientSet *fakeclientset.Clientset) {
				clientSet.PrependReactor("get", "secrets", func(action clientgotesting.Action) (bool, runtime.Object, error) {
					get := action.(clientgotesting.GetAction)
					if get.GetName() != "cluster-agent-token" {
						return false, nil, nil
					}
					return true, nil, errors.New("existing token secret get failed")
				})
			},
			wantErr: "existing token secret get failed",
		},
		{
			name: "token secret create error",
			prepare: func(clientSet *fakeclientset.Clientset) {
				clientSet.PrependReactor("get", "secrets", func(action clientgotesting.Action) (bool, runtime.Object, error) {
					get := action.(clientgotesting.GetAction)
					if get.GetName() != "cluster-agent-token" {
						return false, nil, nil
					}
					return true, nil, k8serrors.NewNotFound(schema.GroupResource{Resource: "secrets"}, get.GetName())
				})
				clientSet.PrependReactor("create", "secrets", func(clientgotesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("token secret create failed")
				})
			},
			wantErr: "token secret create failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseObjects := []runtime.Object{
				newTestServiceAccount("cluster-agent", "cluster-agent-token-mvp6d"),
				newSecret("cluster-agent-token-mvp6d", map[string][]byte{
					caCrtKey:    []byte("ca"),
					tokenSecKey: []byte("token"),
				}),
			}
			if tt.name == "existing token secret get error" || tt.name == "token secret create error" {
				baseObjects = []runtime.Object{
					newTestServiceAccount("cluster-agent"),
				}
			}

			clientSet := fakeclientset.NewSimpleClientset(baseObjects...)
			tt.prepare(clientSet)

			c := newLoadClusterInfoTestClient(clientSet)
			_, err := c.loadClusterInfo(context.Background())
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestDisConnect_DoesNotBlockWithoutReceiver(t *testing.T) {
	c := New()

	done := make(chan struct{})
	go func() {
		c.DisConnect()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("DisConnect blocked without an active connect receiver")
	}
}

func TestOnConnect_ReturnsWhenDisConnectRequested(t *testing.T) {
	c := New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.setActiveConnectCancel(cancel)

	done := make(chan error, 1)
	go func() {
		done <- c.onConnect(ctx, nil)
	}()

	c.DisConnect()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("onConnect did not return after disconnect request")
	}
}

func newLoadClusterInfoTestClient(clientSet *fakeclientset.Clientset) *Client {
	c := New(WithConfig(&config.Config{
		CollectClusterInfo:        true,
		ErdaNamespace:             metav1.NamespaceDefault,
		K8SApiServerAddr:          "https://kubernetes.default.svc",
		ServiceAccount:            "cluster-agent",
		ServiceAccountTokenSecret: "cluster-agent-token",
	}))
	c.newInClusterClient = func(...k8sclient.Option) (*k8sclient.K8sClient, error) {
		return &k8sclient.K8sClient{ClientSet: clientSet}, nil
	}
	return c
}

func newTestServiceAccount(name string, secretNames ...string) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
	}
	for _, secretName := range secretNames {
		sa.Secrets = append(sa.Secrets, corev1.ObjectReference{Name: secretName})
	}
	return sa
}

func newSecret(name string, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Data: data,
	}
}
