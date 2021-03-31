package toleration

import (
	apiv1 "k8s.io/api/core/v1"
)

func GenTolerations() []apiv1.Toleration {
	return []apiv1.Toleration{
		{
			Key:      "node-role.kubernetes.io/lb",
			Operator: "Exists",
			Effect:   "NoSchedule",
		},
		{
			Key:      "node-role.kubernetes.io/master",
			Operator: "Exists",
			Effect:   "NoSchedule",
		},
	}
}
