package bundle

//import (
//	"encoding/json"
//	"testing"
//	"time"
//
//	apps "k8s.io/api/apps/v1"
//	v1 "k8s.io/api/core/v1"
//	yaml2 "k8s.io/apimachinery/pkg/util/yaml"
//
//	"github.com/erda-project/erda/apistructs"
//	"github.com/erda-project/erda/pkg/http/httpclient"
//)
//
//func TestSteveRequest(t *testing.T) {
//	bundleOpts := []Option{
//		WithHTTPClient(
//			httpclient.New(
//				httpclient.WithTimeout(30*time.Second, time.Second*60),
//			)),
//		WithCMP(),
//	}
//	bdl := New(bundleOpts...)
//
//	// Test get node
//	res, err := bdl.GetSteveResource(&apistructs.SteveRequest{
//		Type:          apistructs.K8SNode,
//		ClusterName:   "terminus-dev",
//		Name:          "node-010000006198",
//	})
//	if err != nil {
//		t.Errorf("get node failed: %v", err)
//	}
//
//	data, err := json.Marshal(res.K8SResource)
//	if err != nil {
//		t.Errorf("json marshal failed in get node: %v", err)
//	}
//	var node v1.Node
//	if err = json.Unmarshal(data, &node); err != nil {
//		t.Errorf("json unmarshal failed in get node: %v", err)
//	}
//
//	// Test update node
//	node.Labels["test"] = "test"
//	res, err = bdl.UpdateSteveResource(&apistructs.SteveRequest{
//		Type:          apistructs.K8SNode,
//		ClusterName:   "terminus-dev",
//		Name:          node.Name,
//		Obj:           &node,
//	})
//	if err != nil {
//		t.Errorf("update node failed: %v", err)
//	}
//
//	res, err = bdl.GetSteveResource(&apistructs.SteveRequest{
//		Type:          apistructs.K8SNode,
//		ClusterName:   "terminus-dev",
//		Name:          "node-010000006198",
//	})
//	if err != nil {
//		t.Errorf("get node failed: %v", err)
//	}
//
//	data, err = json.Marshal(res.K8SResource)
//	if err != nil {
//		t.Errorf("json marshal failed in get newNode: %v", err)
//	}
//
//	var newNode v1.Node
//	if err = json.Unmarshal(data, &newNode); err != nil {
//		t.Errorf("json unmarshal failed in get newNode: %v", err)
//	}
//
//	delete(newNode.Labels, "test")
//	res, err = bdl.UpdateSteveResource(&apistructs.SteveRequest{
//		Type:          apistructs.K8SNode,
//		ClusterName:   "terminus-dev",
//		Name:          newNode.Name,
//		Obj:           &newNode,
//	})
//	if err != nil {
//		t.Errorf("update newNode failed: %v", err)
//	}
//
//	// Test get node 404
//	res, err = bdl.GetSteveResource(&apistructs.SteveRequest{
//		Type:          apistructs.K8SNode,
//		ClusterName:   "terminus-dev",
//		Name:          "node-xxx",
//	})
//	if err == nil {
//		t.Errorf("expect 404 error but not returned")
//	}
//
//	// Test list nodes
//	collection, err := bdl.ListSteveResource(&apistructs.SteveRequest{
//		Type:          apistructs.K8SNode,
//		ClusterName:   "terminus-dev",
//	})
//	if err != nil {
//		t.Errorf("list nodes failed: %v", err)
//	}
//
//	var nodes []*v1.Node
//	for _, res := range collection.Data {
//		var node v1.Node
//		data, err = json.Marshal(res.K8SResource)
//		if err != nil {
//			t.Errorf("json marshal failed in list nodes: %v", err)
//		}
//		if err = json.Unmarshal(data, &node); err != nil {
//			t.Errorf("json unmarshal failed in list nodes: %v", err)
//		}
//		nodes = append(nodes, &node)
//	}
//
//	// Test list nodes by selector
//	collection, err = bdl.ListSteveResource(&apistructs.SteveRequest{
//		Type:          apistructs.K8SNode,
//		ClusterName:   "terminus-dev",
//		LabelSelector: []string{
//			"dice/job=true",
//		},
//	})
//	if err != nil {
//		t.Errorf("list nodes by selector failed: %v", err)
//	}
//
//	nodes = nil
//	for _, res := range collection.Data {
//		var node v1.Node
//		data, err = json.Marshal(res.K8SResource)
//		if err != nil {
//			t.Errorf("json marshal failed in list node by selector: %v", err)
//		}
//		if err = json.Unmarshal(data, &node); err != nil {
//			t.Errorf("json unmarshal failed in list node by selector: %v", err)
//		}
//		nodes = append(nodes, &node)
//	}
//
//	// Test create deployment
//	yml := `
//apiVersion: apps/v1
//kind: Deployment
//metadata:
// labels:
//   run: debug
// namespace: default
// name: debug-test
//spec:
// replicas: 1
// selector:
//   matchLabels:
//     run: debug
// template:
//   metadata:
//     labels:
//       run: debug
//   spec:
//     containers:
//     - command:
//       - sleep
//       - "99999"
//       image: craigmchen/debug:v1
//       imagePullPolicy: IfNotPresent
//       name: debug
//`
//	var deploy apps.Deployment
//	js, err:= yaml2.ToJSON([]byte(yml))
//	if err != nil {
//		t.Errorf("yaml to json failed in create deploy: %v", err)
//	}
//
//	if err := json.Unmarshal(js, &deploy); err != nil {
//		t.Errorf("json unmarshal failed in create deploy: %v", err)
//	}
//
//	res, err = bdl.CreateSteveResource(&apistructs.SteveRequest{
//		Type:          apistructs.K8SDeployment,
//		ClusterName:   "terminus-dev",
//		Obj:           &deploy,
//	})
//	_ = res
//	if err != nil {
//		t.Errorf("create deploy failed: %v", err)
//	}
//
//	// Test delete deployment
//	if err = bdl.DeleteSteveResource(&apistructs.SteveRequest{
//		Type:          apistructs.K8SDeployment,
//		ClusterName:   "terminus-dev",
//		Name:          "debug-test",
//		Namespace:     "default",
//	}); err != nil {
//		t.Errorf("delete deploy failed: %v", err)
//	}
//}
