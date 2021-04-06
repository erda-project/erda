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

/* 由于要实现 多个eventbox实例同时watch相同目录，并且只处理一次，而etcd库中提供的分布式锁在etcd断开连接时候有问题，
eventbox中用事务来实现, 但是需要一个 etcd client， 所以在这个文件实现一下 NewEtcdClient
*/
package etcdclient

import (
	"crypto/tls"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/pkg/errors"
)

const (
	// The short keepalive timeout and interval have been chosen to aggressively
	// detect a failed etcd server without introducing much overhead.
	keepaliveTime    = 30 * time.Second
	keepaliveTimeout = 10 * time.Second
)

var mutex = new(sync.RWMutex)
var instance *clientv3.Client

func NewEtcdClientSingleInstance() (*clientv3.Client, error) {
	if instance != nil {
		return instance, nil
	}
	mutex.Lock()
	defer mutex.Unlock()
	if instance == nil {
		client, err := NewEtcdClient()
		if err != nil {
			return nil, err
		}
		instance = client
	}
	return instance, nil
}

func NewEtcdClient() (*clientv3.Client, error) {
	var endpoints []string
	env := os.Getenv("ETCD_ENDPOINTS")
	if env == "" {
		endpoints = []string{"http://127.0.0.1:2379"}
	} else {
		endpoints = strings.Split(env, ",")
	}

	var tlsConfig *tls.Config
	if len(endpoints) < 1 {
		return nil, errors.New("Invalid Etcd endpoints")
	}
	url, err := url.Parse(endpoints[0])
	if err != nil {
		return nil, errors.Wrap(err, "Invalid Etcd endpoints")
	}
	if url.Scheme == "https" {
		tlsInfo := transport.TLSInfo{
			CertFile:      "/certs/etcd-client.pem",
			KeyFile:       "/certs/etcd-client-key.pem",
			TrustedCAFile: "/certs/etcd-ca.pem",
		}
		tlsConfig, err = tlsInfo.ClientConfig()
		if err != nil {
			return nil, errors.Wrap(err, "Invalid Etcd TLS config")
		}
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:            endpoints,
		DialKeepAliveTime:    keepaliveTime,
		DialKeepAliveTimeout: keepaliveTimeout,
		TLS:                  tlsConfig,
	})
	return cli, err
}
