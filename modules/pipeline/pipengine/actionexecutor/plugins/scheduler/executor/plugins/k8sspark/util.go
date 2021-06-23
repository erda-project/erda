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

package k8sspark

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	AliyunPullSecret = "aliyun-registry"

	jobKind                 = "SparkApplication"
	sparkVersion            = "2.4.0"
	jobAPIVersion           = "sparkoperator.k8s.io/v1beta2"
	rbacAPIVersion          = "rbac.authorization.k8s.io/v1"
	rbacAPIGroup            = "rbac.authorization.k8s.io"
	sparkServiceAccountName = "spark"
	sparkRoleName           = "spark-role"
	sparkRoleBindingName    = "spark-role-binding"
	imagePullPolicyAlways   = "Always"
	prefetechVolumeName     = "pre-fetech-volume"

	K8SLabelPrefix = "dice/"
)

func stringptr(s string) *string {
	return &s
}

func int32ptr(n int32) *int32 {
	return &n
}

func int64ptr(n int64) *int64 {
	return &n
}

func float32ptr(n float32) *float32 {
	return &n
}

func float64ptr(n float64) *float64 {
	return &n
}

func addMainApplicationFile(conf *apistructs.BigdataConf) (string, error) {
	var appResource = conf.Spec.Resource

	if strings.HasPrefix(appResource, "local://") || strings.HasPrefix(appResource, "http://") {
		return appResource, nil
	}

	if strings.HasPrefix(appResource, "/") {
		return strutil.Concat("local://", appResource), nil
	}

	return "", errors.Errorf("invalid job spec, resource %s", appResource)
}

func addLabels() map[string]string {
	labels := make(map[string]string)

	// "dice/job": ""
	jobKey := strutil.Concat(K8SLabelPrefix, apistructs.TagJob)
	labels[jobKey] = ""

	// "dice/bigdata": ""
	bigdataKey := strutil.Concat(K8SLabelPrefix, apistructs.TagBigdata)
	labels[bigdataKey] = ""

	return labels
}
