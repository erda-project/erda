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

package datomic

import (
	"context"
	"path/filepath"
	"strconv"
	"time"

	"github.com/erda-project/erda/pkg/jsonstore/etcd"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

const (
	dir = "/datomic/"
	TTL = 600
)

type DInt struct {
	ctx             context.Context
	es              *etcd.Store
	key             string
	leaseID         clientv3.LeaseID
	cancelKeepalive context.CancelFunc
}

func New(key string) (*DInt, error) {
	es, err := etcd.New()
	if err != nil {
		return nil, err
	}
	path := mkEtcdPath(key)
	ec := es.GetClient()
	lease := clientv3.NewLease(ec)
	l, err := lease.Grant(context.Background(), TTL)
	if err != nil {
		return nil, err
	}
	leaseID := l.ID

	resp, err := es.GetClient().
		Txn(context.Background()).
		If(clientv3.Compare(clientv3.Version(path), "=", 0)).
		Then(clientv3.OpPut(path, strconv.FormatInt(0, 10), clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet(path)).
		Commit()
	if err != nil {
		return nil, err
	}
	if !resp.Succeeded {
		if _, err := lease.Revoke(context.Background(), leaseID); err != nil {
			return nil, err
		}
		leaseID = clientv3.LeaseID(resp.Responses[0].GetResponseRange().Kvs[0].Lease)
	}

	keepalivectx, cancelKeepalive := context.WithCancel(context.Background())
	if leaseID != 0 {
		go keepaliveLease(keepalivectx, lease, leaseID)
	}

	return &DInt{es: es, key: path, leaseID: leaseID, cancelKeepalive: cancelKeepalive}, nil
}

// quit loop reason:
// 1. ctx cancel
// 2. lease expired
func keepaliveLease(ctx context.Context, lease clientv3.Lease, id clientv3.LeaseID) {
	for {
		keepalive, err := lease.KeepAlive(ctx, id)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		for range keepalive {
		}
		r, err := lease.TimeToLive(context.Background(), id)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		if r.TTL <= 0 {
			break
		}
		select {
		case <-ctx.Done():
			return
		}
	}
}

func (d *DInt) Add(delta int64) (old uint64, new uint64, err error) {
	add := func(stm concurrency.STM) error {
		old, err = strconv.ParseUint(stm.Get(d.key), 10, 64)
		if err != nil {
			return err
		}
		if delta >= 0 {
			new = old + uint64(delta)
		} else {
			new = old - uint64(-delta)
		}
		stm.Put(d.key, strconv.FormatUint(new, 10), clientv3.WithLease(d.leaseID))
		return nil
	}
	if _, err = concurrency.NewSTM(d.es.GetClient(), add); err != nil {
		return
	}
	return
}

// Store will set 'new' value if 'cond' is satisfied, if cond is true, return (true,err)
func (d *DInt) Store(cond func(old uint64) bool, new uint64) (bool, error) {
	var condResult bool
	store := func(stm concurrency.STM) error {
		old, err := strconv.ParseUint(stm.Get(d.key), 10, 64)
		if err != nil {
			return err
		}
		if cond(old) {
			stm.Put(d.key, strconv.FormatUint(new, 10), clientv3.WithLease(d.leaseID))
			condResult = true
			return nil
		}
		condResult = false
		return nil
	}
	if _, err := concurrency.NewSTM(d.es.GetClient(), store); err != nil {
		return condResult, err
	}
	return condResult, nil
}

// clear kv in etcd
func (d *DInt) Clear() error {
	if _, err := d.es.GetClient().Txn(context.Background()).
		If(clientv3.Compare(clientv3.Version(d.key), "!=", 0)).
		Then(clientv3.OpDelete(d.key)).Commit(); err != nil {
		return err
	}
	d.cancelKeepalive()
	return nil
}

func mkEtcdPath(key string) string {
	return filepath.Join(dir, key)
}
