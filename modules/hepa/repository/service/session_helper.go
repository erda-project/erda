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

package service

import (
	"github.com/pkg/errors"
	"github.com/xormplus/xorm"

	"github.com/erda-project/erda/modules/hepa/repository/orm"
)

type SessionHelper struct {
	session   *xorm.Session
	completed bool
	closed    bool
}

func NewSessionHelper() (*SessionHelper, error) {
	engine, err := orm.GetSingleton()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	session := engine.NewSession()
	err = session.Begin()
	if err != nil {
		session.Close()
		return nil, errors.WithStack(err)
	}
	return &SessionHelper{session, false, false}, nil
}

func (impl *SessionHelper) GetSessionHelper() *SessionHelper {
	return impl
}

func (impl *SessionHelper) Session() *xorm.Session { return impl.session }

func (impl *SessionHelper) Begin() error {
	if impl.closed {
		return errors.New("session already closed")
	}
	if !impl.completed {
		return errors.New("last trans not completed")
	}
	err := impl.session.Begin()
	if err != nil {
		impl.session.Close()
		return errors.WithStack(err)
	}
	impl.completed = false
	return nil
}

func (impl *SessionHelper) Commit() error {
	if impl.completed {
		return nil
	}
	err := impl.session.Commit()
	if err != nil {
		return errors.WithStack(err)
	}
	impl.completed = true
	return nil
}

func (impl *SessionHelper) Rollback() error {
	if impl.completed {
		return nil
	}
	err := impl.session.Rollback()
	if err != nil {
		return errors.WithStack(err)
	}
	impl.completed = true
	return nil
}

func (impl *SessionHelper) Close() {
	if impl.closed {
		return
	}
	impl.session.Close()
	impl.completed = true
	impl.closed = true
}
