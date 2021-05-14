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

package demo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
)

func init() {
	executortypes.Register("DEMO", func(name executortypes.Name, clustername string, options map[string]string, optionsPlus interface{}) (executortypes.Executor, error) {
		return &Demo{
			name: name,
		}, nil
	})
}

type Demo struct {
	name executortypes.Name
}

func (d *Demo) Kind() executortypes.Kind {
	return "DEMO"
}

func (d *Demo) Name() executortypes.Name {
	return d.name
}

func (d *Demo) Create(ctx context.Context, specObj interface{}) (interface{}, error) {
	logrus.Infof("demo create ...")

	time.Sleep(time.Second * 5)
	return nil, nil
}

func (d *Demo) Destroy(ctx context.Context, spec interface{}) error {
	logrus.Infof("demo destroy ...")

	time.Sleep(time.Second * 5)
	return nil
}

func (d *Demo) Status(ctx context.Context, specObj interface{}) (apistructs.StatusDesc, error) {
	logrus.Infof("demo status ...")

	time.Sleep(time.Second * 5)
	return apistructs.StatusDesc{
		Status: apistructs.StatusStoppedOnOK,
	}, nil
}

func (d *Demo) Remove(ctx context.Context, spec interface{}) error {
	logrus.Infof("demo remove ...")

	time.Sleep(time.Second * 5)
	return nil
}

func (d *Demo) Update(ctx context.Context, specObj interface{}) (interface{}, error) {
	logrus.Infof("demo update ...")

	time.Sleep(time.Second * 5)
	return nil, nil
}

func (d *Demo) Inspect(ctx context.Context, spec interface{}) (interface{}, error) {
	logrus.Infof("demo inspect ...")

	time.Sleep(time.Second * 5)
	return nil, nil
}

func (d *Demo) Cancel(ctx context.Context, spec interface{}) (interface{}, error) {
	logrus.Infof("demo cancel ...")

	time.Sleep(time.Second * 5)
	return nil, nil
}
func (d *Demo) Precheck(ctx context.Context, specObj interface{}) (apistructs.ServiceGroupPrecheckData, error) {
	return apistructs.ServiceGroupPrecheckData{Status: "ok"}, nil
}

func (d *Demo) SetNodeLabels(setting executortypes.NodeLabelSetting, hosts []string, labels map[string]string) error {
	return errors.New("SetNodeLabels not implemented in Demo")
}

func (d *Demo) CapacityInfo() apistructs.CapacityInfoData {
	return apistructs.CapacityInfoData{}
}
func (d *Demo) ResourceInfo(brief bool) (apistructs.ClusterResourceInfoData, error) {
	return apistructs.ClusterResourceInfoData{}, fmt.Errorf("resourceinfo not support for demo")
}
func (k *Demo) CleanUpBeforeDelete() {}
func (k *Demo) JobVolumeCreate(ctx context.Context, spec interface{}) (string, error) {
	return "", fmt.Errorf("not support for demo")
}
func (*Demo) KillPod(podname string) error {
	return fmt.Errorf("not support for demo")
}
