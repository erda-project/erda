// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/erda-project/erda/pkg/aliyunclient"
	"go.uber.org/ratelimit"
)

type CloudVendor interface {
	DoReq(request interface{}, response interface{}) error
}

// vendors
type aliyun struct {
	cfg              *AliyunConfig
	limiter          ratelimit.Limiter
	reqLimitDuration time.Duration
	client           *aliyunclient.Client
	ak, sk           string
	bucket           uint64
	done             chan struct{}
}

type AliyunConfig struct {
	Timeout          time.Duration
	MaxQPS           int
	ReqLimit         uint64
	ReqLimitDuration time.Duration
}

func (a *aliyun) limit(timeout time.Duration) error {
	timer := time.After(timeout)
	for {
		select {
		case <-timer:
			return fmt.Errorf("limit timeout")
		default:
		}

		if atomic.LoadUint64(&a.bucket) == 0 {
			continue
		} else {
			atomic.AddUint64(&a.bucket, ^uint64(0))
			break
		}
	}
	a.limiter.Take()
	return nil
}

func (a *aliyun) DoReq(request interface{}, response interface{}) error {
	if err := a.limit(a.cfg.Timeout); err != nil {
		return fmt.Errorf("limit req. err=%s", err)
	}

	errConvert := fmt.Errorf("convert failed")
	req, ok := request.(requests.AcsRequest)
	if !ok {
		return errConvert
	}
	resp, ok := response.(responses.AcsResponse)
	if !ok {
		return errConvert
	}
	return a.client.DoAction(req, resp)
}

func (a *aliyun) init() {
	go func() {
		for {
			select {
			case <-a.done:
				return
			default:
			}
			time.Sleep(a.reqLimitDuration)

			atomic.StoreUint64(&a.bucket, a.cfg.ReqLimit)
		}
	}()

}

func NewAliyunVendor(ak, sk string, cfg AliyunConfig) (v *aliyun, err error) {
	client := &aliyunclient.Client{}
	err = client.InitWithAccessKey("", ak, sk)
	if err != nil {
		return nil, fmt.Errorf("failed create aliyun client, err %s", err)
	}
	v = &aliyun{
		ak:               ak,
		sk:               sk,
		client:           client,
		done:             make(chan struct{}),
		bucket:           cfg.ReqLimit,
		reqLimitDuration: cfg.ReqLimitDuration,
		limiter:          ratelimit.New(cfg.MaxQPS),
		cfg:              &cfg,
	}
	v.init()

	return v, nil
}
