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

package cmp_dashboard_workloads

import (
	"testing"

	"github.com/rancher/wrangler/pkg/data"
)

func TestParseWorkloadStatus(t *testing.T) {
	fields := make([]string, 8, 8)
	fields[2], fields[3] = "1", "1"
	deployment := data.Object{
		"kind": "Deployment",
		"metadata": map[string]interface{}{
			"fields": fields,
		},
	}
	status, color, err := ParseWorkloadStatus(deployment)
	if err != nil {
		t.Error(err)
	}
	if status != "Active" || color != "green" {
		t.Errorf("test failed, deployment status is unexpected")
	}
	fields[2], fields[3] = "0", "1"
	deployment = data.Object{
		"kind": "Deployment",
		"metadata": map[string]interface{}{
			"fields": fields,
		},
	}
	status, color, err = ParseWorkloadStatus(deployment)
	if err != nil {
		t.Error(err)
	}
	if status != "Error" || color != "red" {
		t.Errorf("test failed, deployment status is unexpected")
	}

	fields = make([]string, 11, 11)
	fields[1], fields[3] = "1", "1"
	daemonset := data.Object{
		"kind": "DaemonSet",
		"metadata": map[string]interface{}{
			"fields": fields,
		},
	}
	status, color, err = ParseWorkloadStatus(daemonset)
	if err != nil {
		t.Error(err)
	}
	if status != "Active" || color != "green" {
		t.Errorf("test failed, daemonset status is unexpected")
	}
	fields[1], fields[3] = "0", "1"
	daemonset = data.Object{
		"kind": "DaemonSet",
		"metadata": map[string]interface{}{
			"fields": fields,
		},
	}
	status, color, err = ParseWorkloadStatus(daemonset)
	if err != nil {
		t.Error(err)
	}
	if status != "Error" || color != "red" {
		t.Errorf("test failed, daemonset status is unexpected")
	}

	fields = make([]string, 5, 5)
	fields[1] = "1/1"
	statefulset := data.Object{
		"kind": "StatefulSet",
		"metadata": map[string]interface{}{
			"fields": fields,
		},
	}
	status, color, err = ParseWorkloadStatus(statefulset)
	if err != nil {
		t.Error(err)
	}
	if status != "Active" || color != "green" {
		t.Errorf("test failed, statefulset status is unexpected")
	}
	fields[1] = "0/1"
	statefulset = data.Object{
		"kind": "StatefulSet",
		"metadata": map[string]interface{}{
			"fields": fields,
		},
	}
	status, color, err = ParseWorkloadStatus(statefulset)
	if err != nil {
		t.Error(err)
	}
	if status != "Error" || color != "red" {
		t.Errorf("test failed, statefulset status is unexpected")
	}

	fields = make([]string, 7, 7)
	job := data.Object{
		"kind": "Job",
		"metadata": map[string]interface{}{
			"fields": fields,
		},
		"status": map[string]interface{}{
			"active": 1,
			"field":  0,
		},
	}
	status, color, err = ParseWorkloadStatus(job)
	if err != nil {
		t.Error(err)
	}
	if status != "Active" || color != "green" {
		t.Errorf("test failed, job status is unexpected")
	}
	job = data.Object{
		"kind": "Job",
		"metadata": map[string]interface{}{
			"fields": fields,
		},
		"status": map[string]interface{}{
			"active": 0,
			"failed": 1,
		},
	}
	status, color, err = ParseWorkloadStatus(job)
	if err != nil {
		t.Error(err)
	}
	if status != "Failed" || color != "red" {
		t.Errorf("test failed, job status is unexpected")
	}
	job = data.Object{
		"kind": "Job",
		"metadata": map[string]interface{}{
			"fields": fields,
		},
		"status": map[string]interface{}{
			"succeeded": 1,
		},
	}
	status, color, err = ParseWorkloadStatus(job)
	if err != nil {
		t.Error(err)
	}
	if status != "Succeeded" || color != "steelBlue" {
		t.Errorf("test failed, job status is unexpected")
	}

	fields = make([]string, 7, 7)
	cronjob := data.Object{
		"kind": "CronJob",
		"metadata": map[string]interface{}{
			"fields": fields,
		},
	}
	status, color, err = ParseWorkloadStatus(cronjob)
	if err != nil {
		t.Error(err)
	}
	if status != "Active" || color != "green" {
		t.Errorf("test failed, cronjob status is unexpected")
	}
}
