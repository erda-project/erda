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

package gormutil

import (
	"context"
	"os"
	"time"

	"gorm.io/gorm/logger"
)

type SQLCollector struct {
	logger.Interface
	filename string
}

func NewSQLCollector(filename string, baseLogger logger.Interface) (*SQLCollector, error) {
	if baseLogger == nil {
		baseLogger = &logger.Recorder
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return &SQLCollector{
		Interface: baseLogger,
		filename:  filename,
	}, nil
}

func (c SQLCollector) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	c.Interface.Trace(ctx, begin, fc, err)
	sql, _ := fc()
	c.collect(sql)
}

func (c SQLCollector) collect(sql string) {
	if file, err := os.OpenFile(c.filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644); err == nil {
		_, _ = file.WriteString(sql + "  /*-LINE END-*/\n")
	}
}
