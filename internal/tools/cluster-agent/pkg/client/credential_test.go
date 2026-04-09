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
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/cluster-agent/config"
	"github.com/erda-project/erda/pkg/k8sclient"
)

func TestWatchClusterCredential_RebuildsWatcherAfterChannelClose(t *testing.T) {
	secretEvent := credentialSecretEvent(t, watch.Modified, "updated-ak")
	firstWatcher := newFakeCredentialWatcher()
	secondWatcher := newFakeCredentialWatcher()

	var (
		mu         sync.Mutex
		buildCount int
		watchers   = []credentialWatcher{firstWatcher, secondWatcher}
	)

	c := newCredentialWatchTestClient()
	c.newCredentialWatcher = func(_ context.Context, _ kubernetes.Interface, _ string) (credentialWatcher, error) {
		mu.Lock()
		defer mu.Unlock()
		if buildCount >= len(watchers) {
			return nil, fmt.Errorf("unexpected watcher build %d", buildCount)
		}
		w := watchers[buildCount]
		buildCount++
		return w, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- c.watchClusterCredential(ctx)
	}()

	firstWatcher.close()
	secondWatcher.send(secretEvent)

	require.Eventually(t, func() bool {
		return c.getAccessKey() == "updated-ak"
	}, time.Second, 10*time.Millisecond)

	cancel()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("watchClusterCredential did not stop")
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 2, buildCount)
}

func TestWatchClusterCredential_RetriesWatcherCreation(t *testing.T) {
	secretEvent := credentialSecretEvent(t, watch.Modified, "retried-ak")
	retryWatcher := newFakeCredentialWatcher()

	var (
		mu         sync.Mutex
		buildCount int
	)

	c := newCredentialWatchTestClient()
	c.newCredentialWatcher = func(_ context.Context, _ kubernetes.Interface, _ string) (credentialWatcher, error) {
		mu.Lock()
		defer mu.Unlock()
		buildCount++
		if buildCount == 1 {
			return nil, fmt.Errorf("temporary watch creation failure")
		}
		return retryWatcher, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- c.watchClusterCredential(ctx)
	}()

	retryWatcher.send(secretEvent)

	require.Eventually(t, func() bool {
		return c.getAccessKey() == "retried-ak"
	}, time.Second, 10*time.Millisecond)

	cancel()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("watchClusterCredential did not stop")
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 2, buildCount)
}

func TestConsumeClusterCredentialEvents_IgnoresNonUpdatingEvents(t *testing.T) {
	tests := []struct {
		name             string
		initialAccessKey string
		event            watch.Event
	}{
		{
			name:             "watch error",
			initialAccessKey: "initial-ak",
			event: watch.Event{
				Type: watch.Error,
				Object: &metav1.Status{
					Status: metav1.StatusFailure,
					Code:   500,
				},
			},
		},
		{
			name:             "deleted secret",
			initialAccessKey: "initial-ak",
			event:            credentialSecretEvent(t, watch.Deleted, "deleted-ak"),
		},
		{
			name:             "missing access key",
			initialAccessKey: "initial-ak",
			event: watch.Event{
				Type: watch.Modified,
				Object: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:            apistructs.ErdaClusterCredential,
						Namespace:       metav1.NamespaceDefault,
						ResourceVersion: "2",
					},
				},
			},
		},
		{
			name:             "unchanged access key",
			initialAccessKey: "initial-ak",
			event:            credentialSecretEvent(t, watch.Modified, "initial-ak"),
		},
		{
			name:             "illegal secret object",
			initialAccessKey: "initial-ak",
			event: watch.Event{
				Type:   watch.Modified,
				Object: &metav1.Status{Status: metav1.StatusSuccess},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCredentialWatchTestClient()
			c.setAccessKey(tt.initialAccessKey)

			reconnectCtx, reconnectCancel := context.WithCancel(context.Background())
			t.Cleanup(reconnectCancel)
			c.setActiveConnectCancel(reconnectCancel)

			err := c.consumeClusterCredentialEvents(context.Background(), newFakeCredentialWatcher(tt.event))
			require.ErrorIs(t, err, errCredentialWatcherClosed)
			assert.Equal(t, tt.initialAccessKey, c.getAccessKey())
			assertContextNotCanceled(t, reconnectCtx)
		})
	}
}

func TestConsumeClusterCredentialEvents_UpdatesAccessKeyAndRequestsReconnect(t *testing.T) {
	tests := []struct {
		name         string
		runOnConnect bool
	}{
		{
			name:         "during connect attempt",
			runOnConnect: false,
		},
		{
			name:         "during active connection",
			runOnConnect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCredentialWatchTestClient()
			reconnectCtx, reconnectCancel := context.WithCancel(context.Background())
			defer reconnectCancel()
			c.setActiveConnectCancel(reconnectCancel)

			var connectDone <-chan error
			if tt.runOnConnect {
				done := make(chan error, 1)
				connectDone = done
				go func() {
					done <- c.onConnect(reconnectCtx, nil)
				}()
			}

			err := c.consumeClusterCredentialEvents(context.Background(), newFakeCredentialWatcher(credentialSecretEvent(t, watch.Modified, "next-ak")))
			require.ErrorIs(t, err, errCredentialWatcherClosed)
			assert.Equal(t, "next-ak", c.getAccessKey())
			assertContextCanceled(t, reconnectCtx)

			if connectDone != nil {
				select {
				case err := <-connectDone:
					require.NoError(t, err)
				case <-time.After(time.Second):
					t.Fatal("onConnect did not return after reconnect request")
				}
			}
		})
	}
}

func TestConsumeClusterCredentialEvents_ReturnsNilOnContextCancellation(t *testing.T) {
	c := newCredentialWatchTestClient()
	watcher := newFakeCredentialWatcher()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := c.consumeClusterCredentialEvents(ctx, watcher)
	require.NoError(t, err)
}

func TestConsumeClusterCredentialEvents_DoesNotLogAccessKey(t *testing.T) {
	c := newCredentialWatchTestClient()

	var logs bytes.Buffer
	restore := swapLogrusOutput(&logs)
	defer restore()

	err := c.consumeClusterCredentialEvents(context.Background(), newFakeCredentialWatcher(credentialSecretEvent(t, watch.Modified, "sensitive-ak")))
	require.ErrorIs(t, err, errCredentialWatcherClosed)
	assert.NotContains(t, logs.String(), "sensitive-ak")
}

func TestCredentialWatchRetryDelay(t *testing.T) {
	tests := []struct {
		name          string
		retryInterval time.Duration
		attempt       int
		want          time.Duration
	}{
		{
			name:          "uses default interval when zero",
			retryInterval: 0,
			attempt:       0,
			want:          time.Second,
		},
		{
			name:          "negative attempt treated as zero",
			retryInterval: 2 * time.Second,
			attempt:       -1,
			want:          2 * time.Second,
		},
		{
			name:          "applies exponential backoff",
			retryInterval: time.Second,
			attempt:       3,
			want:          8 * time.Second,
		},
		{
			name:          "caps maximum delay",
			retryInterval: 2 * time.Second,
			attempt:       10,
			want:          maxCredentialWatchRetryInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.watchRetryInterval = tt.retryInterval
			assert.Equal(t, tt.want, c.credentialWatchRetryDelay(tt.attempt))
		})
	}
}

func TestGetRetryWatcher_PropagatesContextToListAndWatch(t *testing.T) {
	ctxKey := testContextKey("marker")
	ctx := context.WithValue(context.Background(), ctxKey, "value")

	baseClient := fakeclientset.NewSimpleClientset()
	secretWatcher := watch.NewFake()
	recordingSecrets := &recordingSecretInterface{
		listResult: &corev1.SecretList{
			ListMeta: metav1.ListMeta{
				ResourceVersion: "17",
			},
			Items: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      apistructs.ErdaClusterCredential,
						Namespace: metav1.NamespaceDefault,
					},
					Data: map[string][]byte{
						apistructs.ClusterAccessKey: []byte("initial-ak"),
					},
				},
			},
		},
		watchResult: secretWatcher,
	}
	coreV1Recorder := &recordingCoreV1{
		CoreV1Interface: baseClient.CoreV1(),
		secrets:         recordingSecrets,
	}
	c := New()
	c.setAccessKey("")

	retryWatcher, err := c.getRetryWatcher(ctx, &recordingKubeClient{
		Clientset: baseClient,
		coreV1:    coreV1Recorder,
	}, metav1.NamespaceDefault)
	require.NoError(t, err)
	defer retryWatcher.Stop()

	require.Eventually(t, func() bool {
		_, watchCtx, _, _ := recordingSecrets.snapshot()
		return watchCtx != nil
	}, time.Second, 10*time.Millisecond)
	secretWatcher.Stop()

	listCtx, watchCtx, listOpts, watchOpts := recordingSecrets.snapshot()
	require.NotNil(t, listCtx)
	require.NotNil(t, watchCtx)
	assert.Equal(t, "value", listCtx.Value(ctxKey))
	assert.Equal(t, "value", watchCtx.Value(ctxKey))
	assert.Equal(t, metav1.NamespaceDefault, coreV1Recorder.namespaceValue())
	assert.Equal(t, "metadata.name="+apistructs.ErdaClusterCredential, listOpts.FieldSelector)
	assert.Equal(t, "metadata.name="+apistructs.ErdaClusterCredential, watchOpts.FieldSelector)
	assert.Equal(t, "17", watchOpts.ResourceVersion)
	assert.Equal(t, "initial-ak", c.getAccessKey())
}

func TestNewClusterCredentialRetryWatcher_UsesProvidedContext(t *testing.T) {
	ctxKey := testContextKey("marker")
	ctx := context.WithValue(context.Background(), ctxKey, "value")

	baseClient := fakeclientset.NewSimpleClientset()
	secretWatcher := watch.NewFake()
	recordingSecrets := &recordingSecretInterface{
		watchResult: secretWatcher,
	}
	coreV1Recorder := &recordingCoreV1{
		CoreV1Interface: baseClient.CoreV1(),
		secrets:         recordingSecrets,
	}

	retryWatcher, err := newClusterCredentialRetryWatcher(
		ctx,
		&recordingKubeClient{
			Clientset: baseClient,
			coreV1:    coreV1Recorder,
		},
		metav1.NamespaceDefault,
		mustClusterCredentialSelector(t),
		"23",
	)
	require.NoError(t, err)
	defer retryWatcher.Stop()

	require.Eventually(t, func() bool {
		_, watchCtx, _, _ := recordingSecrets.snapshot()
		return watchCtx != nil
	}, time.Second, 10*time.Millisecond)
	secretWatcher.Stop()

	_, watchCtx, _, watchOpts := recordingSecrets.snapshot()
	require.NotNil(t, watchCtx)
	assert.Equal(t, "value", watchCtx.Value(ctxKey))
	assert.Equal(t, "metadata.name="+apistructs.ErdaClusterCredential, watchOpts.FieldSelector)
	assert.Equal(t, "23", watchOpts.ResourceVersion)
}

func TestGetRetryWatcher_ListErrorScenarios(t *testing.T) {
	tests := []struct {
		name    string
		listErr error
		wantErr string
	}{
		{
			name:    "not found returns list error",
			listErr: k8serrors.NewNotFound(schema.GroupResource{Resource: "secrets"}, apistructs.ErdaClusterCredential),
			wantErr: "get init secret list error: secrets \"" + apistructs.ErdaClusterCredential + "\" not found",
		},
		{
			name:    "other list error returns immediately",
			listErr: errors.New("list failed"),
			wantErr: "get init secret list error: list failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseClient := fakeclientset.NewSimpleClientset()
			recordingSecrets := &recordingSecretInterface{
				listErr: tt.listErr,
			}
			coreV1Recorder := &recordingCoreV1{
				CoreV1Interface: baseClient.CoreV1(),
				secrets:         recordingSecrets,
			}
			c := New()

			_, err := c.getRetryWatcher(context.Background(), &recordingKubeClient{
				Clientset: baseClient,
				coreV1:    coreV1Recorder,
			}, metav1.NamespaceDefault)

			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestAccessKeyConcurrentAccess(t *testing.T) {
	c := newCredentialWatchTestClient()

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 2000; j++ {
				c.setAccessKey(fmt.Sprintf("ak-%d", j))
			}
		}()
	}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 2000; j++ {
				_ = c.getAccessKey()
			}
		}()
	}

	wg.Wait()
}

func newCredentialWatchTestClient() *Client {
	c := New(WithConfig(&config.Config{
		ErdaNamespace: metav1.NamespaceDefault,
	}))
	c.newInClusterClient = func(...k8sclient.Option) (*k8sclient.K8sClient, error) {
		return &k8sclient.K8sClient{
			ClientSet: fakeclientset.NewSimpleClientset(),
		}, nil
	}
	c.watchRetryInterval = time.Millisecond
	return c
}

func mustClusterCredentialSelector(t *testing.T) fields.Selector {
	t.Helper()

	selector, err := clusterCredentialSelector()
	require.NoError(t, err)
	return selector
}

type testContextKey string

type recordingKubeClient struct {
	*fakeclientset.Clientset
	coreV1 *recordingCoreV1
}

func (c *recordingKubeClient) CoreV1() corev1client.CoreV1Interface {
	return c.coreV1
}

type recordingCoreV1 struct {
	mu sync.Mutex
	corev1client.CoreV1Interface
	secrets   *recordingSecretInterface
	namespace string
}

func (c *recordingCoreV1) Secrets(ns string) corev1client.SecretInterface {
	c.mu.Lock()
	c.namespace = ns
	c.mu.Unlock()
	return c.secrets
}

func (c *recordingCoreV1) namespaceValue() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.namespace
}

type recordingSecretInterface struct {
	mu          sync.Mutex
	listCtx     context.Context
	watchCtx    context.Context
	listOpts    metav1.ListOptions
	watchOpts   metav1.ListOptions
	listErr     error
	listResult  *corev1.SecretList
	watchResult watch.Interface
}

func (s *recordingSecretInterface) Create(context.Context, *corev1.Secret, metav1.CreateOptions) (*corev1.Secret, error) {
	panic("unexpected Create call")
}

func (s *recordingSecretInterface) Update(context.Context, *corev1.Secret, metav1.UpdateOptions) (*corev1.Secret, error) {
	panic("unexpected Update call")
}

func (s *recordingSecretInterface) Delete(context.Context, string, metav1.DeleteOptions) error {
	panic("unexpected Delete call")
}

func (s *recordingSecretInterface) DeleteCollection(context.Context, metav1.DeleteOptions, metav1.ListOptions) error {
	panic("unexpected DeleteCollection call")
}

func (s *recordingSecretInterface) Get(context.Context, string, metav1.GetOptions) (*corev1.Secret, error) {
	panic("unexpected Get call")
}

func (s *recordingSecretInterface) List(ctx context.Context, opts metav1.ListOptions) (*corev1.SecretList, error) {
	s.mu.Lock()
	s.listCtx = ctx
	s.listOpts = opts
	listErr := s.listErr
	s.mu.Unlock()
	if listErr != nil {
		return nil, listErr
	}
	if s.listResult != nil {
		return s.listResult, nil
	}
	panic("unexpected List fallback")
}

func (s *recordingSecretInterface) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	s.mu.Lock()
	s.watchCtx = ctx
	s.watchOpts = opts
	s.mu.Unlock()
	if s.watchResult != nil {
		return s.watchResult, nil
	}
	panic("unexpected Watch fallback")
}

func (s *recordingSecretInterface) snapshot() (context.Context, context.Context, metav1.ListOptions, metav1.ListOptions) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listCtx, s.watchCtx, s.listOpts, s.watchOpts
}

func (s *recordingSecretInterface) Patch(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) (*corev1.Secret, error) {
	panic("unexpected Patch call")
}

func (s *recordingSecretInterface) Apply(context.Context, *corev1apply.SecretApplyConfiguration, metav1.ApplyOptions) (*corev1.Secret, error) {
	panic("unexpected Apply call")
}

func credentialSecretEvent(t *testing.T, eventType watch.EventType, accessKey string) watch.Event {
	t.Helper()

	return watch.Event{
		Type: eventType,
		Object: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:            apistructs.ErdaClusterCredential,
				Namespace:       metav1.NamespaceDefault,
				ResourceVersion: "2",
			},
			Data: map[string][]byte{
				apistructs.ClusterAccessKey: []byte(accessKey),
			},
		},
	}
}

func assertContextCanceled(t *testing.T, ctx context.Context) {
	t.Helper()

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("expected context cancellation")
	}
}

func assertContextNotCanceled(t *testing.T, ctx context.Context) {
	t.Helper()

	select {
	case <-ctx.Done():
		t.Fatal("unexpected context cancellation")
	default:
	}
}

func swapLogrusOutput(buf *bytes.Buffer) func() {
	logger := logrus.StandardLogger()
	originalOut := logger.Out
	logger.SetOutput(buf)
	return func() {
		logger.SetOutput(originalOut)
	}
}

type fakeCredentialWatcher struct {
	result chan watch.Event
	once   sync.Once
}

func newFakeCredentialWatcher(events ...watch.Event) *fakeCredentialWatcher {
	size := len(events)
	if size == 0 {
		size = 1
	}
	w := &fakeCredentialWatcher{
		result: make(chan watch.Event, size),
	}
	for _, event := range events {
		w.result <- event
	}
	if len(events) > 0 {
		w.close()
	}
	return w
}

func (w *fakeCredentialWatcher) send(event watch.Event) {
	w.result <- event
}

func (w *fakeCredentialWatcher) ResultChan() <-chan watch.Event {
	return w.result
}

func (w *fakeCredentialWatcher) Stop() {
	w.close()
}

func (w *fakeCredentialWatcher) close() {
	w.once.Do(func() {
		close(w.result)
	})
}
