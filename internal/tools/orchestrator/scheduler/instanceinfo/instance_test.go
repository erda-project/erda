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

package instanceinfo

import (
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/database/dbengine"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gdb, err := gorm.Open("mysql", db)
	assert.NoError(t, err)
	gdb.LogMode(false)

	return gdb, mock, func() {
		db.Close()
	}
}

func TestInstanceReader_Do(t *testing.T) {
	gdb, mock, cleanup := setupMockDB(t)
	defer cleanup()

	dbEngine := &dbengine.DBEngine{DB: gdb}
	client := &Client{db: dbEngine}

	now := time.Now()

	testcases := []struct {
		name       string
		builder    func(*InstanceReader) *InstanceReader
		args       []driver.Value
		queryRegex string
	}{
		{
			name: "ByCluster",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByCluster("cluster-1")
			},
			args:       []driver.Value{"cluster-1"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(cluster = \\?\\) ORDER BY started_at desc",
		},
		{
			name: "ByNamespace",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByNamespace("ns")
			},
			args:       []driver.Value{"ns"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(namespace = \\?\\) ORDER BY started_at desc",
		},
		{
			name: "ByName",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByName("name")
			},
			args:       []driver.Value{"name"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(name = \\?\\) ORDER BY started_at desc",
		},
		{
			name: "ByOrgID and ByOrgName",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByOrgID("1").ByOrgName("org")
			},
			args:       []driver.Value{"1", "org"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(org_id = \\?\\) AND \\(org_name = \\?\\) ORDER BY started_at desc",
		},
		{
			name: "ByPhases",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByPhases("Running", "Failed")
			},
			args:       []driver.Value{"Running", "Failed"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(phase IN \\(\\?,\\?\\)\\) ORDER BY started_at desc",
		},
		{
			name: "ByFinishedTime",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByFinishedTime(5)
			},
			args:       []driver.Value{"5"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(finished_at < now\\(\\) - interval \\? day\\) ORDER BY started_at desc",
		},
		{
			name: "ByUpdatedTime",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByUpdatedTime(10)
			},
			args:       []driver.Value{"10"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(updated_at < '.*' - interval \\? second\\) ORDER BY started_at desc",
		},
		{
			name: "ByInstanceIP",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByInstanceIP("10.0.0.1", "10.0.0.2")
			},
			args:       []driver.Value{"10.0.0.1", "10.0.0.2"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(container_ip IN \\(\\?,\\?\\)\\) ORDER BY started_at desc",
		},
		{
			name: "ByMetaLike",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByMetaLike("version")
			},
			args:       []driver.Value{"%version%"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(meta LIKE \\?\\) ORDER BY started_at desc",
		},
		{
			name: "WithLimit",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByName("abc").Limit(1)
			},
			args:       []driver.Value{"abc"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(name = \\?\\) ORDER BY started_at desc LIMIT 1",
		},
		{
			name: "ByProjectName and ByProjectID",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByProjectName("project").ByProjectID("1")
			},
			args:       []driver.Value{"project", "1"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(project_name = \\?\\) AND \\(project_id = \\?\\) ORDER BY started_at desc",
		},
		{
			name: "ByApplicationName and ByApplicationID",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByApplicationName("app").ByApplicationID("1")
			},
			args:       []driver.Value{"app", "1"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(application_name = \\?\\) AND \\(application_id = \\?\\) ORDER BY started_at desc",
		},
		{
			name: "ByRuntimeName and ByRuntimeID",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByRuntimeName("runtime").ByRuntimeID("1")
			},
			args:       []driver.Value{"runtime", "1"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(runtime_name = \\?\\) AND \\(runtime_id = \\?\\) ORDER BY started_at desc",
		},
		{
			name: "ByService and ByServiceType",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByService("service").ByServiceType("type")
			},
			args:       []driver.Value{"service", "type"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(service_name = \\?\\) AND \\(service_type = \\?\\) ORDER BY started_at desc",
		},
		{
			name: "ByWorkspace",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByWorkspace("prod")
			},
			args:       []driver.Value{"prod"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(workspace = \\?\\) ORDER BY started_at desc",
		},
		{
			name: "ByContainerID",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByContainerID("container-1")
			},
			args:       []driver.Value{"container-1"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(container_id = \\?\\) ORDER BY started_at desc",
		},
		{
			name: "ByTaskID",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByTaskID("task-1")
			},
			args:       []driver.Value{"task-1"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(task_id = \\?\\) ORDER BY started_at desc",
		},
		{
			name: "ByNotTaskID",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByNotTaskID("task-1")
			},
			args:       []driver.Value{"task-1"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(task_id <> \\?\\) ORDER BY started_at desc",
		},
		{
			name: "ByAddonID",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByAddonID("addon-1")
			},
			args:       []driver.Value{"addon-1"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(addon_id = \\?\\) ORDER BY started_at desc",
		},
		{
			name: "ByHostIP",
			builder: func(r *InstanceReader) *InstanceReader {
				return r.ByHostIP("10.0.0.1", "10.0.0.2")
			},
			args:       []driver.Value{"10.0.0.1", "10.0.0.2"},
			queryRegex: "SELECT \\* FROM `s_instance_info` WHERE \\(host_ip IN \\(\\?,\\?\\)\\) ORDER BY started_at desc",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"id", "cluster", "started_at"}).
				AddRow(uint64(1), "cluster-1", now)

			mock.ExpectQuery(tc.queryRegex).
				WithArgs(tc.args...).
				WillReturnRows(rows)

			reader := client.InstanceReader()
			results, err := tc.builder(reader).Do()
			assert.NoError(t, err)
			if len(results) > 0 {
				assert.Equal(t, uint64(1), results[0].ID)
			}
		})
	}
}

func TestInstanceReader_Do_Error(t *testing.T) {
	gdb, mock, cleanup := setupMockDB(t)
	defer cleanup()

	dbEngine := &dbengine.DBEngine{DB: gdb}
	client := &Client{db: dbEngine}

	mock.ExpectQuery("SELECT \\* FROM `s_instance_info` WHERE \\(cluster = \\?\\) ORDER BY started_at desc").
		WithArgs("error-cluster").
		WillReturnError(assert.AnError)

	reader := client.InstanceReader()
	_, err := reader.ByCluster("error-cluster").Do()
	assert.Error(t, err)
}

func TestInstanceWriter(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		gdb, mock, cleanup := setupMockDB(t)
		defer cleanup()

		dbEngine := &dbengine.DBEngine{DB: gdb}
		client := &Client{db: dbEngine}

		instance := &InstanceInfo{
			Cluster:   "test-cluster",
			Namespace: "test-ns",
			Name:      "test-name",
		}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `s_instance_info` .*").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		writer := client.InstanceWriter()
		err := writer.Create(instance)
		assert.NoError(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		gdb, mock, cleanup := setupMockDB(t)
		defer cleanup()

		dbEngine := &dbengine.DBEngine{DB: gdb}
		client := &Client{db: dbEngine}

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM `s_instance_info` WHERE \\(id IN \\(\\?,\\?\\)\\)").
			WithArgs(1, 2).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		writer := client.InstanceWriter()
		err := writer.Delete(1, 2)
		assert.NoError(t, err)
	})
}

func TestInstanceInfo_Metadata(t *testing.T) {
	tests := []struct {
		name   string
		meta   string
		key    string
		want   string
		wantOk bool
	}{
		{
			name:   "valid key-value pair",
			meta:   "key1=value1,key2=value2",
			key:    "key1",
			want:   "value1",
			wantOk: true,
		},
		{
			name:   "non-existent key",
			meta:   "key1=value1,key2=value2",
			key:    "key3",
			want:   "",
			wantOk: false,
		},
		{
			name:   "empty meta",
			meta:   "",
			key:    "key1",
			want:   "",
			wantOk: false,
		},
		{
			name:   "invalid format",
			meta:   "key1=value1,invalid",
			key:    "key1",
			want:   "value1",
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := InstanceInfo{Meta: tt.meta}
			got, ok := i.Metadata(tt.key)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}
