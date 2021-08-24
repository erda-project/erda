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

package aoptypes

import "github.com/sirupsen/logrus"

// TuneChain 表示一组有序 TunePoint
type TuneChain []TunePoint

// Handle 根据上下文调用 TuneChain
func (chain TuneChain) Handle(ctx *TuneContext) error {
	if len(chain) == 0 {
		return nil
	}
	for _, point := range chain {
		logrus.Debugf("begin handle tune point, type: %s, trigger: %s, name: %s", point.Type(), ctx.SDK.TuneTrigger, point.Name())
		if err := point.Handle(ctx); err != nil {
			logrus.Errorf("end handle tune point, type: %s, trigger: %s, name: %s, failed, err: %v", point.Type(), ctx.SDK.TuneTrigger, point.Name(), err)
		} else {
			logrus.Debugf("end handle tune point, type: %s, trigger: %s, name: %s, success", point.Type(), ctx.SDK.TuneTrigger, point.Name())
		}
	}
	return nil
}

func (chain TuneChain) Len() int {
	return len(chain)
}

func (chain TuneChain) Less(i, j int) bool {
	//return chain[i].Rank() < chain[j].Rank()
	return true
}

func (chain TuneChain) Swap(i, j int) { chain[i], chain[j] = chain[j], chain[i] }
