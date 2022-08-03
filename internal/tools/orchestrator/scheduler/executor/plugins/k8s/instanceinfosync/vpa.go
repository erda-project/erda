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
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func buildVPAEventInfo(bdl *bundle.Bundle, pod *corev1.Pod, errorinfo string, errorinfo_human string, tp string) {
	dedupid := fmt.Sprintf("%s-%s-%s", pod.Labels["DICE_RUNTIME_ID"], pod.Labels["DICE_SERVICE_NAME"], tp)
	if err := bdl.CreateErrorLog(&apistructs.ErrorLogCreateRequest{
		ErrorLog: apistructs.ErrorLog{
			ResourceType:   apistructs.RuntimeError,
			Level:          apistructs.ErrorLevel,
			ResourceID:     pod.Labels["DICE_RUNTIME_ID"],
			OccurrenceTime: strconv.FormatInt(time.Now().Unix(), 10),
			HumanLog:       errorinfo_human,
			PrimevalLog:    errorinfo,
			DedupID:        dedupid,
		},
	}); err != nil {
		logrus.Errorf("createErrorLog: %v", err)
	}

}
