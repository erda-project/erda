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

package schema

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/cassandra"
	mutex "github.com/erda-project/erda-infra/providers/etcd-mutex"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

const (
	impossibleOrgNum = math.MaxInt64
	gcGraceSeconds   = 86400
)

var bdl = bundle.New(bundle.WithCoreServices(), bundle.WithDOP())

type LogSchema interface {
	Name() string
	RunDaemon(ctx context.Context, interval time.Duration, muInf mutex.Interface)
	CreateDefault() error
}

type CassandraSchema struct {
	Logger         logs.Logger
	cass           cassandra.Interface
	defaultSession *gocql.Session
	lastOrgList    []string
	mutexKey       string
}

type Option func(cs *CassandraSchema)

func WithMutexKey(key string) Option {
	return func(cs *CassandraSchema) {
		cs.mutexKey = key
	}
}

func NewCassandraSchema(cass cassandra.Interface, l logs.Logger, ops ...Option) (*CassandraSchema, error) {
	cs := &CassandraSchema{}
	cs.cass = cass
	sysSession, err := cs.cass.Session(&cassandra.SessionConfig{Keyspace: *defaultKeyspaceConfig("system"), Consistency: "LOCAL_ONE"})
	if err != nil {
		return nil, err
	}
	cs.defaultSession = sysSession
	cs.lastOrgList = []string{}
	cs.Logger = l
	cs.mutexKey = "logs_store"

	for _, op := range ops {
		op(cs)
	}

	return cs, nil
}

func (cs *CassandraSchema) Name() string {
	return "schema with Cassandra"
}

func (cs *CassandraSchema) RunDaemon(ctx context.Context, interval time.Duration, muInf mutex.Interface) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	mu, err := muInf.New(ctx, cs.mutexKey)
	if err != nil {
		if err != context.Canceled {
			cs.Logger.Errorf("create mu failed, err: %s", err)
		}
		return
	}

	err = mu.Lock(ctx)
	if err != nil {
		if err != context.Canceled {
			cs.Logger.Errorf("lock failed, err: %s", err)
		}
		return
	}

	defer func() {
		if mu != nil {
			mu.Unlock(context.TODO())
			mu.Close()
		}
	}()

	for {
		err = cs.compareOrUpdate()
		if err != nil {
			cs.Logger.Errorf("refresh org info or keyspaces failed. err: %s", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (cs *CassandraSchema) compareOrUpdate() error {
	orgs, err := cs.listOrgNames()
	if err != nil {
		return err
	}
	if reflect.DeepEqual(orgs, cs.lastOrgList) {
		return nil
	}
	cs.lastOrgList = orgs

	for _, org := range orgs {
		keyspace := KeyspaceWithOrgName(org)
		keyspaceExisted, tableExisted := cs.existedCheck(keyspace)
		if !keyspaceExisted {
			if err := cs.cass.CreateKeyspaces(defaultKeyspaceConfig(keyspace)); err != nil {
				return errors.Wrapf(err, "create keyspace %s failed", keyspace)
			}
		}
		if !tableExisted {
			if err := cs.createTableWithKC(defaultKeyspaceConfig(keyspace)); err != nil {
				return errors.Wrapf(err, "create table failed of %s", keyspace)
			}
		}
	}
	return nil
}

func (cs *CassandraSchema) existedCheck(keyspace string) (keyspaceExisted bool, tableExisted bool) {
	m, err := cs.defaultSession.KeyspaceMetadata(keyspace)
	// keyspace existed check
	if err != nil {
		return false, false
	}

	keyspaceExisted = true
	// table existed check
	tableExisted = true
	for _, table := range []string{"base_log"} {
		_, ok := m.Tables[table]
		if !ok {
			tableExisted = false
			break
		}
	}
	return
}

func (cs *CassandraSchema) listOrgNames() (res []string, err error) {
	res = []string{}
	resp, err := bdl.ListDopOrgs(&apistructs.OrgSearchRequest{PageNo: 1, PageSize: impossibleOrgNum})
	if err != nil {
		// return res, nil
		return nil, fmt.Errorf("get orglist failed. err: %s", err)
	}
	for _, item := range resp.List {
		res = append(res, item.Name)
	}
	return
}

func (cs *CassandraSchema) createTableWithKC(item *cassandra.KeyspaceConfig) error {
	stmts := []string{
		fmt.Sprintf(BaseLogCreateTable, item.Name, gcGraceSeconds),
		fmt.Sprintf(BaseLogAlterTableGCGraceSeconds, item.Name, gcGraceSeconds),
		fmt.Sprintf(BaseLogCreateIndex, item.Name),
	}
	for _, stmt := range stmts {
		if err := cs.createTable(stmt); err != nil {
			return fmt.Errorf("create table failed. stmt=%s, err=%s", stmt, err)
		}
		cs.Logger.Infof("cassandra init cql: %s", stmt)
	}
	return nil
}

func (cs *CassandraSchema) CreateDefault() error {
	for _, stmt := range []string{
		fmt.Sprintf(BaseLogCreateTable, DefaultKeySpace, gcGraceSeconds),
		fmt.Sprintf(BaseLogAlterTableGCGraceSeconds, DefaultKeySpace, gcGraceSeconds),
		fmt.Sprintf(BaseLogCreateIndex, DefaultKeySpace),
		fmt.Sprintf(LogMetaCreateTable, DefaultKeySpace, gcGraceSeconds),
		fmt.Sprintf(LogMetaCreateIndex, DefaultKeySpace),
	} {
		err := cs.createTable(stmt)
		if err != nil {
			return fmt.Errorf("create default tables failed. stmt=%s, err: %w", stmt, err)
		}
	}
	return nil
}

func (cs *CassandraSchema) createTable(stmt string) error {
	q := cs.defaultSession.Query(stmt).Consistency(gocql.All).RetryPolicy(nil)
	err := q.Exec()
	q.Release()
	if err != nil {
		return fmt.Errorf("create tables failed. err: %s", err)
	}
	return nil
}

func defaultKeyspaceConfig(keysapce string) *cassandra.KeyspaceConfig {
	return &cassandra.KeyspaceConfig{
		Name: keysapce,
		Auto: false,
		Replication: cassandra.KeyspaceReplicationConfig{
			Class:  "SimpleStrategy",
			Factor: 2,
		},
	}
}
