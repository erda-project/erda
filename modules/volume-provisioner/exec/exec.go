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

package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type CmdExecutor struct {
	config    *restclient.Config
	client    kubernetes.Interface
	namespace string
}

func NewCmdExecutor(config *restclient.Config, client kubernetes.Interface, namespace string) *CmdExecutor {
	return &CmdExecutor{config, client, namespace}
}

// OnLocal execute 'cmd' on the specified node
func (c *CmdExecutor) OnLocal(cmd string) error {
	cc := exec.Command("/bin/sh", "-c", cmd)
	if _, err := cc.Output(); err != nil {
		return err
	}

	return nil
}

// OnPods execute 'cmd' on the specified pods
func (c *CmdExecutor) OnPods(cmd string, listOption metav1.ListOptions) error {
	pods, err := c.client.CoreV1().Pods(c.namespace).List(context.Background(), listOption)
	if err != nil {
		return err
	}
	for i := range pods.Items {
		if err := c.OnPod(cmd, &pods.Items[i]); err != nil {
			return err
		}
	}
	return nil
}

// OnNodesPods execute 'cmd' on the specified pods of the specified nodes
func (c *CmdExecutor) OnNodesPods(cmd string, nodeListOption, podListOption metav1.ListOptions) error {
	nodes, err := c.client.CoreV1().Nodes().List(context.Background(), nodeListOption)
	if err != nil {
		return err
	}
	for _, node := range nodes.Items {
		podOpt := podListOption
		podOpt.FieldSelector = fmt.Sprintf("spec.nodeName=%s", node.Name)
		if err := c.OnPods(cmd, podOpt); err != nil {
			return err
		}
	}
	return err
}

// OnPod execute 'cmd' on 'pod'
func (c *CmdExecutor) OnPod(cmd string, pod *v1.Pod) error {
	logrus.Infof("namespace: %s, pod: %s, cmd: %v", pod.Namespace, pod.Name, cmd)
	req := c.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		Param("container", pod.Spec.Containers[0].Name).
		Param("stdout", "true").
		Param("stderr", "true")
	for _, c := range []string{"/bin/sh", "-c", cmd} {
		req.Param("command", c)
	}
	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		logrus.Errorf("Failed to create NewSPDYExecutor: %v", err)
		return err
	}
	var b bytes.Buffer
	var bErr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &b,
		Stderr: &bErr,
		Tty:    false,
	})
	logrus.Infof("Result stdout: %s", b.String())
	logrus.Infof("Result stderr: %s", bErr.String())
	if err != nil {
		logrus.Errorf("Failed to create Stream: %v", err)
		return err
	}
	return nil
}
