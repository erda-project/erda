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

package logic

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
)

const (
	fakeSecretName = "ut-regcred"
	fakeTargetNs   = "pipeline-0000000000"
	fakeSourceNs   = "erda-system"
)

func Test_CreateInnerSecretIfNotExist(t *testing.T) {
	sourceSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fakeSecretName,
			Namespace: fakeSourceNs,
		},
	}
	targetSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fakeSecretName,
			Namespace: fakeTargetNs,
		},
	}

	type args struct {
		cs         kubernetes.Interface
		targetNs   string
		secretName string
	}
	tests := []struct {
		name    string
		args    args
		isMount bool
		want    bool
		wantErr bool
	}{
		{
			name: "secret already exists in target namespace",
			args: args{
				cs:         fakeclientset.NewSimpleClientset(targetSecret),
				targetNs:   fakeTargetNs,
				secretName: fakeSecretName,
			},
			want: true,
		},
		{
			name: "secret already exists in target namespace, and source is not empty",
			args: args{
				cs:         fakeclientset.NewSimpleClientset(sourceSecret, targetSecret),
				targetNs:   fakeTargetNs,
				secretName: fakeSecretName,
			},
			want: true,
		},
		{
			name: "secret does not exist in target namespace",
			args: args{
				cs:         fakeclientset.NewSimpleClientset(sourceSecret),
				targetNs:   fakeTargetNs,
				secretName: fakeSecretName,
			},
			want: true,
		},
		{
			name: "doesn't need mount secret",
			args: args{
				cs:         fakeclientset.NewSimpleClientset(),
				targetNs:   fakeTargetNs,
				secretName: fakeSecretName,
			},
			want: false,
		},
		{
			name: "clientset is nil",
			args: args{
				cs:         nil,
				targetNs:   fakeTargetNs,
				secretName: fakeSecretName,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				err error
				got bool
			)

			if got, err = CreateInnerSecretIfNotExist(tt.args.cs, fakeSourceNs, tt.args.targetNs, tt.args.secretName); got != tt.want {
				t.Errorf("CreateInnerSecretIfNotExist() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateInnerSecretIfNotExist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
