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

package addon

import (
	"crypto/md5"
	"encoding/hex"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/agiledragon/gomonkey"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/kms/kmstypes"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func InitAddonSQLMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func() error) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open("mysql", sqlDB)
	if err != nil {
		t.Fatalf("Failed to initialize GORM: %v", err)
	}
	return gormDB, mock, func() error {
		return sqlDB.Close()
	}
}

const (
	instanceID      = "e27cc10b88d31483bb901adc75881d589"
	encryptPassword = "MDMyMTBhZTY1MDkxY2U4NDAzNGEzYzdhMjNkZGE2MTgxZWIwMTKc1hV+Bkh4n3UkqOhhIA6SOaqeAYMDx7vcFj1+p+fE7ZqhmFpku+XbKfan6w=="
)

var (
	mockInfluxDBInstance = &dbclient.AddonInstance{
		ID: instanceID,
	}
	mockInfluxDBServiceGroup = &apistructs.ServiceGroup{
		Dice: apistructs.Dice{
			Services: []apistructs.Service{
				{
					Vip: "influxdb.addon-influxdb--cb510e986b12d471c967be012987abb32.svc.cluster.local",
				},
			},
		},
	}
)

func TestDeployStatus(t *testing.T) {
	gormDB, mock, closeFn := InitAddonSQLMock(t)
	defer closeFn()

	a := Addon{
		db: &dbclient.DBClient{
			DBEngine: &dbengine.DBEngine{
				DB: gormDB,
			},
		},
	}

	d := &dbclient.AddonInstanceExtra{
		ID:         a.getRandomId(),
		InstanceID: instanceID,
		Field:      InfluxDBKMSPasswordKey,
		Value:      encryptPassword,
		Deleted:    apistructs.AddonNotDeleted,
	}

	mock.ExpectQuery("SELECT \\* FROM `tb_middle_instance_extra` WHERE \\(instance_id = \\?\\) AND \\(field = \\?\\) AND \\(is_deleted = \\?\\) ORDER BY `tb_middle_instance_extra`\\.`id` ASC LIMIT 1").
		WithArgs(instanceID, InfluxDBKMSPasswordKey, apistructs.AddonNotDeleted).
		WillReturnRows(sqlmock.NewRows([]string{"id", "instance_id", "field", "is_deleted", "value"}).
			AddRow("1", d.InstanceID, d.Field, apistructs.AddonNotDeleted, d.Value))

	_, err := a.InfluxDBDeployStatus(mockInfluxDBInstance, mockInfluxDBServiceGroup)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInfluxDBInitRender(t *testing.T) {
	envMap := make(diceyml.EnvMap)
	gormDB, mock, closeFn := InitAddonSQLMock(t)
	defer closeFn()

	bdl := bundle.New(bundle.WithKMS())

	a := Addon{
		bdl: bdl,
		db: &dbclient.DBClient{
			DBEngine: &dbengine.DBEngine{
				DB: gormDB,
			},
		},
	}

	patches := gomonkey.ApplyMethod(reflect.TypeOf(bdl), "KMSEncrypt", func(_ *bundle.Bundle, req apistructs.KMSEncryptRequest) (*kmstypes.EncryptResponse, error) {
		return &kmstypes.EncryptResponse{
			KeyID:            "fake",
			CiphertextBase64: encryptPassword,
		}, nil
	})
	defer patches.Reset()

	instance := dbclient.AddonInstanceExtra{
		InstanceID: instanceID,
		Field:      InfluxDBKMSPasswordKey,
		Value:      encryptPassword,
		Deleted:    apistructs.AddonNotDeleted,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `tb_middle_instance_extra` (`id`,`instance_id`,`field`,`value`,`is_deleted`,`create_time`,`update_time`) VALUES (?,?,?,?,?,?,?)")).
		WithArgs(sqlmock.AnyArg(), instance.InstanceID, instance.Field, instance.Value, instance.Deleted, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := a.influxDBInitRender(&apistructs.AddonHandlerCreateItem{
		Options: map[string]string{
			InfluxDBParamsBucket: "new_bucket",
		},
	}, mockInfluxDBInstance, envMap)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenInfluxDBOrg(t *testing.T) {
	tests := []struct {
		name        string
		addonIns    *dbclient.AddonInstance
		expectedMD5 string
	}{
		{
			name: "Basic case",
			addonIns: &dbclient.AddonInstance{
				OrgID:     "org1",
				ProjectID: "proj1",
				Workspace: "dev",
			},
			expectedMD5: calculateMD5("org1_proj1_dev"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := genInfluxDBOrg(tt.addonIns)
			if result != tt.expectedMD5 {
				t.Errorf("genInfluxDBOrg(%v) = %s; want %s", tt.addonIns, result, tt.expectedMD5)
			}
		})
	}
}

func calculateMD5(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}
