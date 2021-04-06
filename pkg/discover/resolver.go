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

package discover

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type endpointInfo struct {
	endpoint  string
	expiresAt time.Time
}

const (
	defaultTimeout = 60 * time.Second
)

var endpointMap = &map[string]endpointInfo{}
var mutex sync.Mutex

func init() {
	go func() {
		for range time.Tick(defaultTimeout / 2) {
			if len(*endpointMap) == 0 {
				continue
			}
			mutex.Lock()
			newEndpointMap := map[string]endpointInfo{}
			for key, value := range *endpointMap {
				if time.Now().Before(value.expiresAt) {
					newEndpointMap[key] = value
					continue
				}
				endpoint, err := resolveEndpoint(key)
				if err != nil {
					logrus.Errorf("resolve endpoint failed, err:%+v", err)
					continue
				}
				newEndpointMap[key] = endpointInfo{
					endpoint:  endpoint,
					expiresAt: time.Now().Add(defaultTimeout),
				}
			}
			endpointMap = &newEndpointMap
			mutex.Unlock()
		}
	}()
}

func setEndpoint(serviceName, endpoint string) {
	mutex.Lock()
	defer mutex.Unlock()
	newEndpointMap := map[string]endpointInfo{}
	for key, value := range *endpointMap {
		newEndpointMap[key] = value
	}
	newEndpointMap[serviceName] = endpointInfo{
		endpoint:  endpoint,
		expiresAt: time.Now().Add(defaultTimeout),
	}
	endpointMap = &newEndpointMap
}

func getEndpoint(serviceName string) (string, bool) {
	endpoint, ok := (*endpointMap)[serviceName]
	return endpoint.endpoint, ok
}

func resolveEndpoint(serviceName string) (endpoint string, err error) {
	_, addrs, err := net.DefaultResolver.LookupSRV(context.Background(), "", "", serviceName)
	if err != nil {
		err = errors.Wrapf(err, "resolve service endpoint failed, service:%s", serviceName)
		return
	}
	if len(addrs) == 0 || addrs[0] == nil {
		err = errors.Errorf("can't find service endpoint, service:%s", serviceName)
		return
	}
	endpoint = fmt.Sprintf("%s:%d", strings.TrimSuffix(addrs[0].Target, "."), addrs[0].Port)
	logrus.Debugf("service: %s endpoint acquired: %s", serviceName, endpoint)
	return
}

func GetEndpoint(serviceName string) (endpoint string, err error) {
	endpoint, ok := getEndpoint(serviceName)
	if ok {
		return
	}
	endpoint, err = resolveEndpoint(serviceName)
	if err != nil {
		return
	}
	go setEndpoint(serviceName, endpoint)
	return
}
