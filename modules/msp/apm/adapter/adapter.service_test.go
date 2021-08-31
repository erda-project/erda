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

package adapter

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	_ "google.golang.org/grpc"

	"github.com/erda-project/erda-proto-go/msp/apm/adapter/pb"
)

////go:generate mockgen -destination=./adapter_logs_test.go -package exporter github.com/erda-project/erda-infra/base/logs Logger
////go:generate mockgen -destination=./adapter_register_test.go -package exporter github.com/erda-project/erda-infra/pkg/transport Register
func Test_adapterService_GetInstrumentationLibrary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := NewMockLogger(ctrl)
	//register := NewMockRegister(ctrl)
	pro := &provider{
		Cfg: &config{
			Library:    []string{"./../../../../conf/msp/adapter/instrumentationlibrary.yaml"},
			ConfigFile: []string{"./../../../../conf/msp/adapter/config.yaml"},
		},
		Log:            logger,
		Register:       nil,
		adapterService: &adapterService{},
		libraryMap: map[string]interface{}{
			"Java Agent":        []interface{}{"Java"},
			"Apache SkyWalking": []interface{}{"Java"},
		},
		configFile: "./../../../../conf/msp/adapter/config.yaml",
	}
	pro.adapterService.p = pro
	_, err := pro.adapterService.GetInstrumentationLibrary(context.Background(), &pb.GetInstrumentationLibraryRequest{})
	if err != nil {
		t.Errorf("should not throw err")
	}
}

func Test_adapterService_GetInstrumentationLibraryDocs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := NewMockLogger(ctrl)
	//register := NewMockRegister(ctrl)
	pro := &provider{
		Cfg: &config{
			Library:    []string{"./../../../../conf/msp/adapter/instrumentationlibrary.yaml"},
			ConfigFile: []string{"./../../../../conf/msp/adapter/config.yaml"},
		},
		Log:            logger,
		Register:       nil,
		adapterService: &adapterService{},
		libraryMap: map[string]interface{}{
			"Java Agent":        []interface{}{"Java"},
			"Apache SkyWalking": []interface{}{"Java"},
		},
		configFile: "./../../../../conf/msp/adapter/config.yaml",
	}
	pro.adapterService.p = pro
	_, err := pro.adapterService.GetInstrumentationLibraryDocs(context.Background(), &pb.GetInstrumentationLibraryDocsRequest{
		Language: "java",
		Strategy: "javaagent",
	})
	if err != nil {
		t.Errorf("shoult not err")
	}
}
