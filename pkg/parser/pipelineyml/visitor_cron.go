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

package pipelineyml

import (
	"strings"
	"time"

	"github.com/erda-project/erda/pkg/cron"
)

const (
	defaultListNextScheduleCount = 5
	maxListNextScheduleCount     = 1000
)

var DefaultCronCompensator = CronCompensator{
	Enable:               false,
	LatestFirst:          true,
	StopIfLatterExecuted: true,
}

type CronVisitor struct {
	// request
	cronStartTime *time.Time
	cronEndTime   *time.Time
	count         int

	// result
	nextTimes []time.Time
	isCron    bool
}

func NewCronVisitor(ops ...CronVisitorOption) *CronVisitor {
	var v CronVisitor
	v.cronStartTime = nil
	v.cronEndTime = nil
	v.count = defaultListNextScheduleCount

	for _, op := range ops {
		op(&v)
	}

	return &v
}

type CronVisitorOption func(*CronVisitor)

func WithCronStartEndTime(cronStartTime, cronEndTime *time.Time) CronVisitorOption {
	return func(v *CronVisitor) {
		v.cronStartTime = cronStartTime
		v.cronEndTime = cronEndTime
	}
}

func WithListNextScheduleCount(count int) CronVisitorOption {
	return func(v *CronVisitor) {
		v.count = count
	}
}

func (v *CronVisitor) Visit(s *Spec) {
	if s.Cron == "" {
		s.CronCompensator = nil
		v.isCron = false
		return
	}

	v.isCron = true

	var (
		schedule cron.Schedule
		err      error
	)

	switch fields := strings.Fields(s.Cron); len(fields) {
	case 7:
		fieldsWithoutYear := fields[:len(fields)-1]
		s.Cron = strings.Join(fieldsWithoutYear, " ")
		fallthrough
	case 6:
		schedule, err = cron.Parse(s.Cron)
	default:
		schedule, err = cron.ParseStandard(s.Cron)
	}
	if err != nil {
		s.appendError(err)
		return
	}

	now := time.Unix(time.Now().Unix(), 0)
	scheduleFrom := now
	if v.cronStartTime != nil {
		scheduleFrom = *v.cronStartTime
	}

	for {
		if len(v.nextTimes) >= maxListNextScheduleCount {
			break
		}
		if v.count >= 0 && len(v.nextTimes) >= v.count {
			break
		}
		nextTime := schedule.Next(scheduleFrom)
		if v.cronEndTime != nil && (*v.cronEndTime).Before(nextTime) {
			break
		}
		v.nextTimes = append(v.nextTimes, nextTime)
		scheduleFrom = nextTime
	}

	// compensator
	if s.CronCompensator == nil {
		s.CronCompensator = &DefaultCronCompensator
	}
}

func ListNextCronTime(cronExpr string, ops ...CronVisitorOption) ([]time.Time, error) {
	s := Spec{Cron: cronExpr}
	v := NewCronVisitor(ops...)
	s.Accept(v)
	return v.nextTimes, s.mergeErrors()
}

func IsCron(s *Spec) bool {
	v := NewCronVisitor(nil, nil)
	s.Accept(v)
	return v.isCron
}
