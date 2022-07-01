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

package instanceinfosync

import (
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	hpatypes "github.com/erda-project/erda/internal/tools/orchestrator/components/horizontalpodscaler/types"
)

func buildHPAEventInfo(bdl *bundle.Bundle, hpa autoscalingv2beta2.HorizontalPodAutoscaler, errorinfo string, errorinfo_human string, tp string) {
	dedupid := fmt.Sprintf("%s-%s-%s", hpa.Labels[hpatypes.ErdaHPAObjectRuntimeIDLabel], hpa.Labels[hpatypes.ErdaHPAObjectRuntimeServiceNameLabel], tp)
	if err := bdl.CreateErrorLog(&apistructs.ErrorLogCreateRequest{
		ErrorLog: apistructs.ErrorLog{
			ResourceType:   apistructs.RuntimeError,
			Level:          apistructs.ErrorLevel,
			ResourceID:     hpa.Labels[hpatypes.ErdaHPAObjectRuntimeIDLabel],
			OccurrenceTime: strconv.FormatInt(time.Now().Unix(), 10),
			HumanLog:       errorinfo_human,
			PrimevalLog:    errorinfo,
			DedupID:        dedupid,
		},
	}); err != nil {
		logrus.Errorf("createErrorLog: %v", err)
	}

}
