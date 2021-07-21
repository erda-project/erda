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

package logic

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/reference"
	"k8s.io/kubectl/pkg/scheme"
)

const (
	DefaultEventsLimit = 5
)

var metadataAccessor = meta.NewAccessor()

// SearchEvents finds events about the specified object.
func SearchEvents(client corev1client.EventsGetter, objOrRef runtime.Object) (*corev1.EventList, error) {
	ref, err := reference.GetReference(scheme.Scheme, objOrRef)
	if err != nil {
		return nil, err
	}
	stringRefKind := string(ref.Kind)
	var refKind *string
	if len(stringRefKind) > 0 {
		refKind = &stringRefKind
	}
	stringRefUID := string(ref.UID)
	var refUID *string
	if len(stringRefUID) > 0 {
		refUID = &stringRefUID
	}

	e := client.Events(ref.Namespace)
	fieldSelector := e.GetFieldSelector(&ref.Name, &ref.Namespace, refKind, refUID)
	initialOpts := metav1.ListOptions{FieldSelector: fieldSelector.String(), Limit: DefaultEventsLimit}
	eventList := &corev1.EventList{}
	err = FollowContinue(&initialOpts,
		func(options metav1.ListOptions) (runtime.Object, error) {
			newEvents, err := e.List(context.TODO(), options)
			if err != nil {
				return nil, err
			}
			eventList.Items = append(eventList.Items, newEvents.Items...)
			return newEvents, nil
		})
	return eventList, err
}

func FollowContinue(initialOpts *metav1.ListOptions,
	listFunc func(metav1.ListOptions) (runtime.Object, error)) error {
	opts := initialOpts
	for {
		list, err := listFunc(*opts)
		if err != nil {
			return err
		}
		nextContinueToken, _ := metadataAccessor.Continue(list)
		if len(nextContinueToken) == 0 {
			return nil
		}
		opts.Continue = nextContinueToken
	}
}
