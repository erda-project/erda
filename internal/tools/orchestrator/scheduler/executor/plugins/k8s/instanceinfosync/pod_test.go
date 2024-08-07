package instanceinfosync

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_extractEnvs(t *testing.T) {
	pod := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{},
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"addon.erda.cloud/id":          "10001",
				"core.erda.cloud/cluster-name": "erda-jicheng",
				"core.erda.cloud/org-name":     "development-erda",
				"core.erda.cloud/org-id":       "888",
				"core.erda.cloud/project-name": "Master",
				"core.erda.cloud/project-id":   "8888",
				"core.erda.cloud/app-name":     "go-demo",
				"core.erda.cloud/app-id":       "88888",
				"core.erda.cloud/runtime-id":   "66666",
				"core.erda.cloud/service-name": "go-demo-web",
				"core.erda.cloud/workspace":    "prod",
				"addon.erda.cloud/type":        "mysql",
			},
		},
	}

	extractEnvs(pod)
}
