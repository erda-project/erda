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

package alert

import (
	"bou.ke/monkey"
	"context"
	"fmt"
	monitor "github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	"github.com/erda-project/erda-proto-go/msp/apm/alert/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/monitor/utils"
	"github.com/jinzhu/gorm"
	"testing"

	"github.com/golang/mock/gomock"
)

////go:generate mockgen -destination=./alert_register_test.go -package alert github.com/erda-project/erda-infra/pkg/transport Register
////go:generate mockgen -destination=./alert_monitor_test.go -package alert github.com/erda-project/erda-proto-go/core/monitor/alert/pb AlertServiceServer
func Test_alertService_CreateCustomizeAlert(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	pro := &provider{
		C:                      &config{},
		DB:                     &gorm.DB{},
		Register:               NewMockRegister(ctrl),
		Perm:                   nil,
		MPerm:                  nil,
		alertService:           &alertService{},
		Monitor:                NewMockAlertServiceServer(ctrl),
		authDb:                 nil,
		mspDb:                  nil,
		bdl:                    &bundle.Bundle{},
		microServiceFilterTags: nil,
	}
	pro.alertService.p = pro
	defer monkey.UnpatchAll()
	monkey.Patch((*alertService).getTKByTenant, func(_ *alertService, tenantGroup string) (string, error) {
		return "e49b5fe96c144069c24f029f2df6559f", nil
	})
	monkey.Patch(utils.NewContextWithHeader, func(ctx context.Context) context.Context {
		return context.Background()
	})
	monkey.Patch(checkCustomMetric, func(customMetrics *monitor.CustomizeMetrics, alert *monitor.CustomizeAlertDetail) error {
		return nil
	})
	monkey.Patch(monitor.AlertServiceServer.CreateCustomizeAlert, func(_ context.Context, _ *monitor.CreateCustomizeAlertRequest) (*monitor.CreateCustomizeAlertResponse, error) {
		return &monitor.CreateCustomizeAlertResponse{
			Data: 18,
		}, nil
	})
	_, err := pro.alertService.CreateCustomizeAlert(context.Background(), &pb.CreateCustomizeAlertRequest{
		Name: "erda_RRdddeeeeesss",
		Rules: []*monitor.CustomizeAlertRule{
			{
				Metric: "status_page",
				Window: 1,
				Functions: []*monitor.CustomizeAlertRuleFunction{
					{
						Field:      "code",
						Alias:      "dds",
						Aggregator: "max",
						Operator:   "eq",
						Value:      nil,
						DataType:   "",
						Unit:       "",
					},
				},
				Filters:             nil,
				Group:               nil,
				Outputs:             nil,
				Select:              nil,
				Attributes:          nil,
				ActivedMetricGroups: nil,
				CreateTime:          0,
				UpdateTime:          0,
			},
		},
		Notifies: nil,
	})
	if err != nil {
		fmt.Println("should not err")
	}
}
