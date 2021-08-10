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

package cql

import (
	"fmt"

	"github.com/gocql/gocql"

	"github.com/erda-project/erda-infra/base/logs"
)

// Cql .
type Cql struct {
	*gocql.Session
	AlertHistory AlertHistoryCql
}

// New .
func New(session *gocql.Session) *Cql {
	return &Cql{
		Session:      session,
		AlertHistory: AlertHistoryCql{session},
	}
}

// Init .
func (cql *Cql) Init(log logs.Logger, gcGraceSeconds int) error {
	for _, stmt := range []string{
		fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS  alert_history (
		group_id	text,
		timestamp	bigint,
		alert_state	text,
		title	text,
		content	text,
		display_url	text,
		PRIMARY KEY ((group_id),timestamp)
	) WITH CLUSTERING ORDER BY (timestamp DESC)
		AND compression={'chunk_length_in_kb':'64','class':'org.apache.cassandra.io.compress.LZ4Compressor'}
		AND caching={'keys':'ALL','rows_per_partition':'NONE'}
		AND compaction={'max_threshold':'32','min_threshold':'4','class':'org.apache.cassandra.db.compaction.SizeTieredCompactionStrategy'}
		AND gc_grace_seconds = %d;`, gcGraceSeconds),
	} {
		q := cql.Session.Query(stmt).Consistency(gocql.All).RetryPolicy(nil)
		err := q.Exec()
		q.Release()
		if err != nil {
			return err
		}
		log.Infof("cassandra init cql: %s", stmt)
	}
	return nil
}
