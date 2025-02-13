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

package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	kedav1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	autoscaling "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	vpatypes "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	papb "github.com/erda-project/erda-proto-go/orchestrator/podscaler/pb"
	"github.com/erda-project/erda/apistructs"
	pstypes "github.com/erda-project/erda/internal/tools/orchestrator/components/podscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/types"
	"github.com/erda-project/erda/pkg/strutil"
)

func (k *Kubernetes) createErdaHPARules(spec interface{}) (interface{}, error) {
	hpaObjects := make(map[string]pstypes.ErdaHPAObject)
	sg, err := ValidateRuntime(spec, "TaskScale")

	if err != nil {
		return nil, err
	}

	if !IsGroupStateful(sg) && sg.ProjectNamespace != "" {
		k.setProjectServiceName(sg)
	}

	if IsGroupStateful(sg) {
		// statefulset application
		// Judge the group from the label, each group is a statefulset
		groups, err := groupStatefulset(sg)
		if err != nil {
			logrus.Infof(err.Error())
			return nil, err
		}

		for _, groupedSG := range groups {
			// 每个  groupedSG 对应一个 statefulSet，其中 Services 数量表示副本数
			// Scale statefulset
			sts, err := k.getStatefulSetAbstract(groupedSG)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			hpaObjects[groupedSG.Services[0].Name] = pstypes.ErdaHPAObject{
				TypeMeta: metav1.TypeMeta{
					APIVersion: sts.APIVersion,
					Kind:       sts.Kind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: sts.Namespace,
					Name:      sts.Name,
				},
			}
		}
	} else {
		// stateless application
		for index, svc := range sg.Services {
			switch svc.WorkLoad {
			case types.ServicePerNode:
				logrus.Errorf("svc %s in sg %+v is daemonset, can not scale", svc.Name, sg)
				errs := fmt.Errorf("svc %s in sg %+v is daemonset, can not scale", svc.Name, sg)
				logrus.Error(errs)
				return nil, errs
			default:
				// Scale deployment
				dp, err := k.getDeploymentAbstract(sg, index)
				if err != nil {
					logrus.Error(err)
					return nil, err
				}
				hpaObjects[sg.Services[index].Name] = pstypes.ErdaHPAObject{
					TypeMeta: metav1.TypeMeta{
						APIVersion: dp.APIVersion,
						Kind:       dp.Kind,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: dp.Namespace,
						Name:      dp.Name,
					},
				}
			}
		}
	}

	return hpaObjects, nil
}

func (k *Kubernetes) cancelErdaHPARules(sg apistructs.ServiceGroup) (interface{}, error) {
	for svc, sc := range sg.Extra {
		scaledObject := papb.ScaledConfig{}
		err := json.Unmarshal([]byte(sc), &scaledObject)
		if err != nil {
			return sg, errors.Errorf("cancel hpa for serviceGroup service %s failed: %v", svc, err)
		}

		if scaledObject.RuleName == "" || scaledObject.RuleNameSpace == "" {
			return sg, errors.Errorf("cancel hpa for sg %#v service %s failed: [name: %s] or [namespace: %s] not set ", sg, svc, scaledObject.RuleName, scaledObject.RuleNameSpace)
		}

		_, err = k.scaledObject.Get(scaledObject.RuleNameSpace, scaledObject.RuleName+"-"+strutil.ToLower(scaledObject.ScaleTargetRef.Kind)+"-"+scaledObject.ScaleTargetRef.Name)
		if err == k8serror.ErrNotFound {
			logrus.Warnf("No need to cancel hpa rule for svc %s, not found scaledObjects for this service", svc)
			continue
		}

		if err = k.scaledObject.Delete(scaledObject.RuleNameSpace, scaledObject.RuleName+"-"+strutil.ToLower(scaledObject.ScaleTargetRef.Kind)+"-"+scaledObject.ScaleTargetRef.Name); err != nil {
			return sg, err
		}
	}

	return sg, nil
}

// only call by when delete runtime
func (k *Kubernetes) cancelErdaPARules(sg apistructs.ServiceGroup) error {
	hpaSg := make(map[string]string)
	vpaSg := make(map[string]string)

	for svcName, sc := range sg.Extra {
		if strings.HasPrefix(svcName, pstypes.ErdaHPAPrefix) {
			subStr := strings.Split(svcName, pstypes.ErdaHPAServiceSepStr)
			if len(subStr) == 2 {
				hpaSg[subStr[1]] = sc
				continue
			}
		}

		if strings.HasPrefix(svcName, pstypes.ErdaVPAPrefix) {
			subStr := strings.Split(svcName, pstypes.ErdaVPAServiceSepStr)
			if len(subStr) == 2 {
				vpaSg[subStr[1]] = sc
			}
		}
	}

	sg.Extra = hpaSg
	_, err := k.cancelErdaHPARules(sg)
	if err != nil {
		logrus.Infof("delete HPA error: %v", err)
		return err
	}

	sg.Extra = vpaSg
	_, err = k.cancelErdaVPARules(sg)
	if err != nil {
		logrus.Infof("delete VPA error: %v", err)
		return err
	}

	return nil
}

func (k *Kubernetes) reApplyErdaHPARules(sg apistructs.ServiceGroup) (interface{}, error) {
	for svc, sc := range sg.Extra {
		scaledObject := papb.ScaledConfig{}
		err := json.Unmarshal([]byte(sc), &scaledObject)
		if err != nil {
			return sg, errors.Errorf("reapply hpa for sg %#v service %s failed: %v", sg, svc, err)
		}

		if scaledObject.RuleName == "" || scaledObject.RuleNameSpace == "" {
			return sg, errors.Errorf("reapply hpa for sg %#v service %s failed: [name: %s] or [namespace: %s] not set ", sg, svc, scaledObject.RuleName, scaledObject.RuleNameSpace)
		}

		scaledObj := convertToKedaScaledObject(scaledObject)

		if scaledObj == nil {
			return nil, errors.New("keda scaled object is nil")
		}

		old := &kedav1alpha1.ScaledObject{}

		err = k.k8sClient.CRClient.Get(context.Background(), client.ObjectKey{
			Namespace: scaledObj.Namespace,
			Name:      scaledObj.Name,
		}, old)

		if err != nil {
			return nil, err
		}

		spec := types.ScaledPatchStruct{}

		spec.Spec.ScaleTargetRef = scaledObj.Spec.ScaleTargetRef
		if scaledObj.Spec.MinReplicaCount != nil && scaledObj.Spec.MaxReplicaCount != nil {

			var minCount = *scaledObj.Spec.MinReplicaCount
			var maxCount = *scaledObj.Spec.MaxReplicaCount

			if minCount > 0 && maxCount > 0 && minCount < maxCount {
				spec.Spec.MinReplicaCount = scaledObj.Spec.MinReplicaCount
				spec.Spec.MaxReplicaCount = scaledObj.Spec.MaxReplicaCount
			}
		}

		if len(scaledObj.Spec.Triggers) > 0 {
			spec.Spec.Triggers = scaledObj.Spec.Triggers
		}

		patchBytes, err := json.Marshal(spec)
		if err != nil {
			return nil, err
		}
		patch := client.RawPatch(apitypes.MergePatchType, patchBytes)

		err = k.k8sClient.CRClient.Patch(context.Background(), old, patch)
		if err != nil {
			return sg, err
		}
	}

	return sg, nil
}

func (k *Kubernetes) manualScale(ctx context.Context, spec interface{}) (interface{}, error) {
	sg, err := ValidateRuntime(spec, "TaskScale")

	if err != nil {
		return nil, err
	}

	if !IsGroupStateful(sg) && sg.ProjectNamespace != "" {
		k.setProjectServiceName(sg)
	}

	// only support scale one service resources
	if len(sg.Services) != 1 {
		logrus.Infof("the scaling service count is not equal 1 for sg.Services: %#v", sg.Services)
		//	return nil, fmt.Errorf("the scaling service count is not equal 1")
	}

	// scale operator use addon update
	operator, ok := sg.Labels["USE_OPERATOR"]
	if ok {
		op, err := k.whichOperator(operator)
		if err != nil {
			return nil, fmt.Errorf("not found addonoperator: %v", operator)
		}
		return sg, addon.Update(op, sg)
	}

	if IsGroupStateful(sg) {
		// statefulset application
		// Judge the group from the label, each group is a statefulset
		groups, err := groupStatefulset(sg)
		if err != nil {
			logrus.Infof(err.Error())
			return sg, err
		}

		for _, groupedSG := range groups {
			// 每个  groupedSG 对应一个 statefulSet，其中 Services 数量表示副本数
			if err = k.scaleStatefulSet(ctx, groupedSG); err != nil {
				logrus.Error(err)
				return sg, err
			}
		}
	} else {
		// stateless application
		for index, svc := range sg.Services {
			switch svc.WorkLoad {
			case types.ServicePerNode:
				logrus.Errorf("svc %s in sg %+v is daemonset, can not scale", svc.Name, sg)
				errs := fmt.Errorf("svc %s in sg %+v is daemonset, can not scale", svc.Name, sg)
				logrus.Error(errs)
				return sg, errs
			default:
				// Scale deployment
				if err = k.scaleDeployment(ctx, sg, index); err != nil {
					logrus.Error(err)
					return sg, err
				}
			}
		}
	}

	return sg, nil
}

func (k *Kubernetes) applyErdaHPARules(sg apistructs.ServiceGroup) (interface{}, error) {
	for svc, sc := range sg.Extra {
		scaledObject := papb.ScaledConfig{}
		err := json.Unmarshal([]byte(sc), &scaledObject)
		if err != nil {
			return sg, errors.Errorf("apply hpa for serviceGroup service %s failed: %v", svc, err)
		}

		if scaledObject.RuleName == "" || scaledObject.RuleNameSpace == "" || scaledObject.ScaleTargetRef.Name == "" {
			return sg, errors.Errorf("apply hpa for serviceGroup service %s failed: [rule name: %s] or [namespace: %s] or [targetRef.Name:%s] not set ", svc, scaledObject.RuleName, scaledObject.RuleNameSpace, scaledObject.ScaleTargetRef.Name)
		}

		scaledObj := convertToKedaScaledObject(scaledObject)
		if err = k.k8sClient.CRClient.Create(context.Background(), scaledObj); err != nil {
			return sg, err
		}
	}

	return sg, nil
}

func (k *Kubernetes) applyErdaVPARules(sg apistructs.ServiceGroup) (interface{}, error) {
	for svc, sc := range sg.Extra {
		scaledObject := papb.RuntimeServiceVPAConfig{}
		err := json.Unmarshal([]byte(sc), &scaledObject)
		if err != nil {
			return sg, errors.Errorf("apply vpa for serviceGroup service %s failed: %v", svc, err)
		}

		if scaledObject.RuleName == "" || scaledObject.RuleNameSpace == "" || scaledObject.ScaleTargetRef.Name == "" {
			return sg, errors.Errorf("apply vpa for serviceGroup service %s failed: [rule name: %s] or [namespace: %s] or [targetRef.Name:%s] not set ", svc, scaledObject.RuleName, scaledObject.RuleNameSpace, scaledObject.ScaleTargetRef.Name)
		}

		scaledObj := convertToVPAObject(scaledObject)
		if err = k.scaledObject.CreateVPA(scaledObj); err != nil {
			return sg, err
		}
	}

	return sg, nil
}

func (k *Kubernetes) cancelErdaVPARules(sg apistructs.ServiceGroup) (interface{}, error) {
	for svc, sc := range sg.Extra {
		scaledObject := papb.RuntimeServiceVPAConfig{}
		err := json.Unmarshal([]byte(sc), &scaledObject)
		if err != nil {
			return sg, errors.Errorf("cancel vpa for serviceGroup service %s failed: %v", svc, err)
		}

		if scaledObject.RuleName == "" || scaledObject.RuleNameSpace == "" {
			return sg, errors.Errorf("cancel vpa for sg %#v service %s failed: [name: %s] or [namespace: %s] not set ", sg, svc, scaledObject.RuleName, scaledObject.RuleNameSpace)
		}

		_, err = k.scaledObject.GetVPA(scaledObject.RuleNameSpace, scaledObject.RuleName+"-"+strutil.ToLower(scaledObject.ScaleTargetRef.Kind)+"-"+scaledObject.ScaleTargetRef.Name)
		if err == k8serror.ErrNotFound {
			logrus.Warnf("No need to cancel vpa rule for svc %s, not found scaledObjects for this service", svc)
			continue
		}

		if err = k.scaledObject.DeleteVPA(scaledObject.RuleNameSpace, scaledObject.RuleName+"-"+strutil.ToLower(scaledObject.ScaleTargetRef.Kind)+"-"+scaledObject.ScaleTargetRef.Name); err != nil {
			return sg, err
		}
	}

	return sg, nil
}

func (k *Kubernetes) reApplyErdaVPARules(sg apistructs.ServiceGroup) (interface{}, error) {
	for svc, sc := range sg.Extra {
		scaledObject := papb.RuntimeServiceVPAConfig{}
		err := json.Unmarshal([]byte(sc), &scaledObject)
		if err != nil {
			return sg, errors.Errorf("re-apply vpa for sg %#v service %s failed: %v", sg, svc, err)
		}

		if scaledObject.RuleName == "" || scaledObject.RuleNameSpace == "" {
			return sg, errors.Errorf("re-apply vpa for sg %#v service %s failed: [name: %s] or [namespace: %s] not set ", sg, svc, scaledObject.RuleName, scaledObject.RuleNameSpace)
		}

		scaledObj := convertToVPAObject(scaledObject)

		err = k.scaledObject.PatchVPA(scaledObj.Namespace, scaledObj.Name, scaledObj)
		if err != nil {
			return sg, err
		}
	}

	return sg, nil
}

func convertToVPAObject(scaledObject papb.RuntimeServiceVPAConfig) *vpatypes.VerticalPodAutoscaler {
	orgID := fmt.Sprintf("%d", scaledObject.OrgID)
	updateMode := vpatypes.UpdateModeAuto
	switch scaledObject.UpdateMode {
	case pstypes.ErdaVPAUpdaterModeRecreate:
		updateMode = vpatypes.UpdateModeRecreate
	case pstypes.ErdaVPAUpdaterModeInitial:
		updateMode = vpatypes.UpdateModeInitial
	case pstypes.ErdaVPAUpdaterModeOff:
		updateMode = vpatypes.UpdateModeOff
	default:
		updateMode = vpatypes.UpdateModeAuto
	}

	maxCpu := fmt.Sprintf("%.fm", scaledObject.MaxResources.Cpu*1000)
	maxMemory := fmt.Sprintf("%.dMi", scaledObject.MaxResources.Mem)

	minCpu := fmt.Sprintf("%.fm", pstypes.ErdaVPAMinResourceCPU*1000)
	minMemory := fmt.Sprintf("%.dMi", pstypes.ErdaVPAMinResourceMemory)

	crps := make([]vpatypes.ContainerResourcePolicy, 0)
	crps = append(crps, vpatypes.ContainerResourcePolicy{
		ContainerName: "*",
		MinAllowed: apiv1.ResourceList{
			apiv1.ResourceCPU:    resource.MustParse(minCpu),
			apiv1.ResourceMemory: resource.MustParse(minMemory),
		},
		MaxAllowed: apiv1.ResourceList{
			apiv1.ResourceCPU:    resource.MustParse(maxCpu),
			apiv1.ResourceMemory: resource.MustParse(maxMemory),
		},
		ControlledResources: &[]apiv1.ResourceName{apiv1.ResourceCPU, apiv1.ResourceMemory},
	})

	return &vpatypes.VerticalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VerticalPodAutoscaler",
			APIVersion: "autoscaling.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      scaledObject.RuleName + "-" + strutil.ToLower(scaledObject.ScaleTargetRef.Kind) + "-" + scaledObject.ScaleTargetRef.Name,
			Namespace: scaledObject.RuleNameSpace,
			Labels: map[string]string{
				pstypes.ErdaPAObjectRuntimeServiceNameLabel: scaledObject.ServiceName,
				pstypes.ErdaPAObjectRuntimeIDLabel:          fmt.Sprintf("%d", scaledObject.RuntimeID),
				pstypes.ErdaPAObjectRuleIDLabel:             scaledObject.RuleID,
				pstypes.ErdaPAObjectOrgIDLabel:              orgID,
				pstypes.ErdaPALabelKey:                      "yes",
			},
		},
		Spec: vpatypes.VerticalPodAutoscalerSpec{
			TargetRef: &autoscaling.CrossVersionObjectReference{
				Name:       scaledObject.ScaleTargetRef.Name,
				APIVersion: scaledObject.ScaleTargetRef.ApiVersion,
				Kind:       scaledObject.ScaleTargetRef.Kind,
			},
			UpdatePolicy: &vpatypes.PodUpdatePolicy{
				UpdateMode: &updateMode,
			},
			ResourcePolicy: &vpatypes.PodResourcePolicy{
				ContainerPolicies: crps,
			},
		},
	}
}

func convertToKedaScaledObject(scaledObject papb.ScaledConfig) *kedav1alpha1.ScaledObject {
	var stabilizationWindowSeconds int32 = 300
	selectPolicy := autoscalingv2.MaxChangePolicySelect

	orgID := fmt.Sprintf("%d", scaledObject.OrgID)
	triggers := make([]kedav1alpha1.ScaleTriggers, 0)
	for _, trigger := range scaledObject.Triggers {
		triggers = append(triggers, kedav1alpha1.ScaleTriggers{
			Type: trigger.Type,
			//Name:     "",
			Metadata: trigger.Metadata,
			// TODO: Retains for Authentication
			/*
				AuthenticationRef: &kedav1alpha1.ScaledObjectAuthRef{
					Name: "",
					Kind: "",
				},
				MetricType: "",
			*/
		})
	}

	return &kedav1alpha1.ScaledObject{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ScaledObject",
			APIVersion: "keda.sh/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      scaledObject.RuleName + "-" + strutil.ToLower(scaledObject.ScaleTargetRef.Kind) + "-" + scaledObject.ScaleTargetRef.Name,
			Namespace: scaledObject.RuleNameSpace,
			Labels: map[string]string{
				pstypes.ErdaPAObjectRuntimeServiceNameLabel: scaledObject.ServiceName,
				pstypes.ErdaPAObjectRuntimeIDLabel:          fmt.Sprintf("%d", scaledObject.RuntimeID),
				pstypes.ErdaPAObjectRuleIDLabel:             scaledObject.RuleID,
				pstypes.ErdaPAObjectOrgIDLabel:              orgID,
				pstypes.ErdaPALabelKey:                      "yes",
			},
		},
		Spec: kedav1alpha1.ScaledObjectSpec{
			ScaleTargetRef: &kedav1alpha1.ScaleTarget{
				Name:       scaledObject.ScaleTargetRef.Name,
				APIVersion: scaledObject.ScaleTargetRef.ApiVersion,
				Kind:       scaledObject.ScaleTargetRef.Kind,
				//EnvSourceContainerName:  "containerName",
			},
			// +optional
			PollingInterval: &scaledObject.PollingInterval,
			// +optional
			CooldownPeriod: &scaledObject.CooldownPeriod,
			// +optional
			//IdleReplicaCount: &scaledObject.PollingInterval,
			MinReplicaCount: &scaledObject.MinReplicaCount,
			MaxReplicaCount: &scaledObject.MaxReplicaCount,
			Advanced: &kedav1alpha1.AdvancedConfig{
				// +optional
				HorizontalPodAutoscalerConfig: &kedav1alpha1.HorizontalPodAutoscalerConfig{
					Behavior: &autoscalingv2.HorizontalPodAutoscalerBehavior{
						ScaleUp: &autoscalingv2.HPAScalingRules{
							StabilizationWindowSeconds: &stabilizationWindowSeconds,
							SelectPolicy:               &selectPolicy,
							Policies: []autoscalingv2.HPAScalingPolicy{
								{
									Type:          autoscalingv2.PodsScalingPolicy,
									Value:         2,
									PeriodSeconds: 30,
								},
								{
									Type:          autoscalingv2.PercentScalingPolicy,
									Value:         50,
									PeriodSeconds: 30,
								},
							},
						},
						ScaleDown: &autoscalingv2.HPAScalingRules{
							StabilizationWindowSeconds: &stabilizationWindowSeconds,
							SelectPolicy:               &selectPolicy,
							Policies: []autoscalingv2.HPAScalingPolicy{
								{
									Type:          autoscalingv2.PodsScalingPolicy,
									Value:         2,
									PeriodSeconds: 30,
								},
								{
									Type:          autoscalingv2.PercentScalingPolicy,
									Value:         50,
									PeriodSeconds: 30,
								},
							},
						},
					},
				},
				// +optional
				RestoreToOriginalReplicaCount: scaledObject.Advanced.RestoreToOriginalReplicaCount,
			},

			Triggers: triggers,
			Fallback: &kedav1alpha1.Fallback{
				FailureThreshold: 3,
				Replicas:         scaledObject.Fallback.Replicas,
			},
		},
	}
}
