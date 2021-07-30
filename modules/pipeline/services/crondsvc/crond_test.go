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

package crondsvc

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/cron"
)

// Result:
// 1000  个约 0.03s
// 10000 个约 1.7s
func TestReloadSpeed(t *testing.T) {
	d := cron.New()
	d.Start()
	for i := 0; i < 10; i++ {
		if err := d.AddFunc("*/1 * * * *", func() {
			fmt.Println("hello world")
		}); err != nil {
			panic(err)
		}
	}
	time.Sleep(time.Second * 2)
}

func TestCrondSvc_ListenCrond(t *testing.T) {

	c := CrondSvc{}
	c.cronChan = make(chan string, 10)
	var client = &dbclient.Client{}
	var cr = &cron.Cron{}

	c.dbClient = client
	c.crond = cr

	patch := monkey.PatchInstanceMethod(reflect.TypeOf(client), "GetPipelineCron", func(client *dbclient.Client, id interface{}) (cron spec.PipelineCron, err error) {
		return spec.PipelineCron{ID: 1, Enable: &[]bool{true}[0], CronExpr: "* * * * * *"}, nil
	})

	patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(cr), "Remove", func(cr *cron.Cron, name string) error {
		assert.Equal(t, name, makePipelineCronName(1), "AddFunc")
		return nil
	})

	patch2 := monkey.PatchInstanceMethod(reflect.TypeOf(cr), "AddFunc", func(cr *cron.Cron, spec string, cmd func(), names ...string) error {
		assert.NotZero(t, names)
		assert.Equal(t, names[0], makePipelineCronName(1), "AddFunc")
		return nil
	})

	// todo refactor bad test
	go c.ListenCrond(func(id uint64) {})
	time.Sleep(2 * time.Second)

	err := c.AddIntoPipelineCrond(1)
	assert.NoError(t, err)
	time.Sleep(2 * time.Second)
	err = c.DeletePipelineCrond(1)
	assert.NoError(t, err)
	time.Sleep(2 * time.Second)

	defer patch.Unpatch()
	defer patch1.Unpatch()
	defer patch2.Unpatch()
}
