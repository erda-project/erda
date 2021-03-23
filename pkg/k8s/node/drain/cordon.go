package drain

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
)

// CordonHelper wraps functionality to cordon/uncordon nodes
type CordonHelper struct {
	node    *corev1.Node
	desired bool
}

// NewCordonHelper returns a new CordonHelper
func NewCordonHelper(node *corev1.Node) *CordonHelper {
	return &CordonHelper{
		node: node,
	}
}

// UpdateIfRequired returns true if c.node.Spec.Unschedulable isn't already set,
// or false when no change is needed.
func (c *CordonHelper) UpdateIfRequired(desired bool) bool {
	c.desired = desired

	return c.node.Spec.Unschedulable != c.desired
}

// PatchOrReplace uses given clientset to update the node status, either by patching or
// updating the given node object; it may return error if the object cannot be encoded as
// JSON, or if either patch or update calls fail; it will also return a second error
// whenever creating a patch has failed.
func (c *CordonHelper) PatchOrReplace(ctx context.Context, clientset kubernetes.Interface) (error, error) {
	client := clientset.CoreV1().Nodes()

	oldData, err := json.Marshal(c.node)
	if err != nil {
		return err, nil
	}

	c.node.Spec.Unschedulable = c.desired

	newData, err := json.Marshal(c.node)
	if err != nil {
		return err, nil
	}

	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldData, newData, c.node)
	if patchErr == nil {
		_, err = client.Patch(ctx, c.node.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	} else {
		_, err = client.Update(ctx, c.node, metav1.UpdateOptions{})
	}
	return err, patchErr
}
