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

package manager

import (
	"context"
	"fmt"
	"io/ioutil"

	cfgpkg "github.com/recallsong/go-utils/config"
)

func (p *provider) loadRolloverBody() error {
	if len(p.C.RolloverBodyFile) <= 0 {
		return nil
	}
	body, err := ioutil.ReadFile(p.C.RolloverBodyFile)
	if err != nil {
		return fmt.Errorf("fail to load index rollover file: %s", err)
	}
	body = cfgpkg.EscapeEnv(body)
	p.rolloverBody = string(body)
	p.L.Info("load rollover body: \n", p.rolloverBody)
	return nil
}

func (p *provider) doRolloverAlias() {
	indices := p.getIndicesAndWait()
	for addon := range indices {
		p.rolloverAlias(p.C.IndexPrefix + addon)
	}
	p.reload <- struct{}{}
}

func (p *provider) rolloverAlias(alias string) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.C.RequestTimeout)
	defer cancel()
	resp, err := p.client.RolloverIndex(alias).BodyString(p.rolloverBody).Do(ctx)
	if err != nil {
		p.L.Errorf("fail to rollover alias %s : %s", alias, err)
		return err
	}
	p.L.Infof("rollover alias %s from %s to %s, %v", alias, resp.OldIndex, resp.NewIndex, resp.Acknowledged)
	return nil
}
