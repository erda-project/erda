// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package eventsync

import (
	"context"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/user/impl/uc/eventsync/dao"
	"github.com/erda-project/erda/internal/core/user/impl/uc/eventsync/ucclient"
	"github.com/erda-project/erda/pkg/cron"
	"github.com/erda-project/erda/pkg/dlock"
	"github.com/erda-project/erda/pkg/strutil"
)

type Syncer struct {
	db        *dao.DBClient
	uc        *ucclient.UCClient
	cron      *cron.Cron
	bdl       *bundle.Bundle
	receivers []Receiver
	cfg       *config
	log       logs.Logger
}

type Receiver interface {
	SendAudits(ucaudits *apistructs.UCAuditsListResponse) ([]int64, error)
	Name() string
}

type Option func(*Syncer)

func NewSyncer(options ...Option) *Syncer {
	s := &Syncer{
		cron: cron.New(),
	}
	for _, op := range options {
		op(s)
	}
	s.RegistryReceiver(NewAuditReceiver(s.bdl), NewMemberReceiver(s.bdl))
	return s
}

func WithDBClient(db *dao.DBClient) Option {
	return func(syncer *Syncer) {
		syncer.db = db
	}
}

func WithUCClient(uc *ucclient.UCClient) Option {
	return func(syncer *Syncer) {
		syncer.uc = uc
	}
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(syncer *Syncer) {
		syncer.bdl = bdl
	}
}

func WithConfig(cfg *config) Option {
	return func(syncer *Syncer) {
		syncer.cfg = cfg
	}
}

func (s *Syncer) RegistryReceiver(receiver ...Receiver) {
	s.receivers = append(s.receivers, receiver...)
}

// Start starts the scheduled tasks.
func (s *Syncer) Start() {
	if s.cron == nil || s.cfg == nil {
		return
	}
	if err := s.cron.AddFunc(s.cfg.UCAuditorCron, s.syncUcAudit, "uc-event-sync"); err != nil {
		s.log.Errorf("add sync job err: %v", err)
	}
	if err := s.cron.AddFunc(s.cfg.CompensationExecCron, s.compensateUcAudit, "uc-event-compensate"); err != nil {
		s.log.Errorf("add compensate job err: %v", err)
	}
	if err := s.cron.AddFunc(s.cfg.UCSyncRecordCleanCron, s.deleteUCSyncRecord, "uc-event-clean"); err != nil {
		s.log.Errorf("add clean job err: %v", err)
	}
	s.cron.Start()
}

// Close stops the scheduled tasks.
func (s *Syncer) Close() {
	if s.cron != nil {
		s.cron.Close()
	}
}

func (s *Syncer) syncUcAudit() {
	lock, err := getLock("/audit/synclock")
	if err != nil {
		s.log.Errorf("get lock err: %v", err)
		return
	}
	defer func() {
		lock.UnlockAndClose()
	}()

	audits, err := s.PullAudits()
	if err != nil {
		s.log.Error(err)
	}
	if audits == nil || len(audits.Result) == 0 {
		return
	}
	for _, r := range s.receivers {
		go func(r Receiver) {
			ucIDs, err := r.SendAudits(audits)
			if err != nil {
				s.log.Errorf("%v send uc audits err: %v", r.Name(), err)
				if err := s.markSendFailedUCEvent(r.Name(), ucIDs); err != nil {
					s.log.Errorf("%v mark uc audits failed err: %v", r.Name(), err)
				}
			}
		}(r)
	}
}

// compensateUcAudit compensates for failed UC event sync.
func (s *Syncer) compensateUcAudit() {
	lock, err := getLock("/audit/compensatelock")
	if err != nil {
		s.log.Errorf("get lock err: %v", err)
		return
	}
	defer func() {
		lock.UnlockAndClose()
	}()

	var ucIDs []int64
	records, err := s.db.GetFailedRecord(s.cfg.CompensationBatchSize)
	if err != nil {
		s.log.Errorf("get failed record err: %v", err)
		return
	}
	for _, v := range records {
		ucIDs = append(ucIDs, v.UCID)
	}
	audits, err := s.PullAuditsByIDs(ucIDs)
	if err != nil {
		s.log.Errorf("pull uc audits by ids error: %v", err)
		return
	}
	if audits == nil {
		s.log.Info("no data need compensate")
		return
	}
	s.log.Infof("starting compensate %v data", len(audits.Result))
	for _, r := range s.receivers {
		go func(r Receiver) {
			_, err := r.SendAudits(audits)
			if err != nil {
				s.log.Errorf("%v compensate uc audits err: %v", r.Name(), err)
				return
			}
			if err := s.removeSendFailedReceiver(r.Name(), ucIDs); err != nil {
				s.log.Errorf("%v remove failed receiver err: %v", r.Name(), err)
			}
		}(r)
	}
}

// deleteUCSyncRecord periodically cleans up UC sync records.
func (s *Syncer) deleteUCSyncRecord() {
	cutoff := time.Now().AddDate(0, 0, -s.cfg.CleanRecordDays)
	if err := s.db.DeleteRecordByTime(cutoff); err != nil {
		s.log.Errorf("clean uc sync record err: %v", err)
	}
}

// PullAudits pulls UC audit events.
func (s *Syncer) PullAudits() (*apistructs.UCAuditsListResponse, error) {
	ucAuditReq, byID, err := s.genUCAuditReq()
	if err != nil {
		return nil, err
	}

	var resp *apistructs.UCAuditsListResponse
	if byID {
		s.log.Infof("last ucid is %v", ucAuditReq.LastID)
		resp, err = s.uc.ListUCAuditsByLastID(*ucAuditReq)
		if err != nil {
			return nil, err
		}
	} else {
		resp, err = s.uc.ListUCAuditsByEventTime(*ucAuditReq)
		if err != nil {
			return nil, err
		}
	}

	var records []*dao.UCSyncRecord
	for _, v := range resp.Result {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", time.Unix(v.EventTime/1e3, 0).Format("2006-01-02 15:04:05"), time.Local)
		if err != nil {
			return nil, err
		}
		records = append(records, &dao.UCSyncRecord{UCID: v.ID, UCEventTime: t})
	}

	if err := s.db.BatchCreateUCSyncRecord(records); err != nil {
		return nil, err
	}

	return resp, nil
}

// PullAuditsByIDs fetches UC audit events by IDs.
func (s *Syncer) PullAuditsByIDs(ucIDs []int64) (*apistructs.UCAuditsListResponse, error) {
	l := len(ucIDs)
	if l == 0 {
		return nil, nil
	}

	req := &apistructs.UCAuditsListRequest{LastID: ucIDs[0] - 1, Size: 1}
	ucAudits := make([]apistructs.UCAudit, 0, l)
	for i := 0; i < l; i++ {
		if i+1 < l && ucIDs[i+1]-ucIDs[i] == 1 {
			req.Size++
			continue
		}

		resp, err := s.uc.ListUCAuditsByLastID(*req)
		if err != nil {
			return nil, err
		}
		req.LastID = ucIDs[i] - 1
		req.Size = 1
		ucAudits = append(ucAudits, resp.Result...)
	}

	return &apistructs.UCAuditsListResponse{Result: ucAudits}, nil
}

func (s *Syncer) genUCAuditReq() (*apistructs.UCAuditsListRequest, bool, error) {
	lastRecord, err := s.db.GetLastNRecord(1)
	if err != nil {
		return nil, false, err
	}

	// On first run or if last event is older than 3 days, fetch from 10 minutes ago.
	if len(lastRecord) == 0 || lastRecord[0].UCEventTime.AddDate(0, 0, 3).Unix() < time.Now().Unix() {
		return &apistructs.UCAuditsListRequest{
			Size:      s.cfg.UCAuditorPullSize,
			EventTime: time.Now().Add(-10*time.Minute).Unix() * 1000,
		}, false, nil
	}

	return &apistructs.UCAuditsListRequest{
		LastID: lastRecord[0].UCID,
		Size:   s.cfg.UCAuditorPullSize,
	}, true, nil
}

// markSendFailedUCEvent marks UC events that failed to send.
func (s *Syncer) markSendFailedUCEvent(unReceiver string, ucIDs []int64) error {
	if len(ucIDs) == 0 {
		return nil
	}
	records, err := s.db.GetRecordByUCIDs(ucIDs)
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.UnReceiver == "" {
			record.UnReceiver = unReceiver
		} else {
			record.UnReceiver = record.UnReceiver + "," + unReceiver
		}

		if err := s.db.UpdateRecord(&record); err != nil {
			s.log.Errorf("mark uc sync record send failed err: %v", err)
		}
	}

	return nil
}

// removeSendFailedReceiver removes the failed receiver mark after successful retry.
func (s *Syncer) removeSendFailedReceiver(receiver string, ucIDs []int64) error {
	records, err := s.db.GetRecordByUCIDs(ucIDs)
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.UnReceiver == "" {
			continue
		}
		receivers := strutil.SplitIfEmptyString(record.UnReceiver, ",")
		l := len(receivers)
		for i, v := range receivers {
			if v == receiver {
				if i == l-1 {
					receivers = receivers[:i]
				} else {
					receivers = append(receivers[:i], receivers[i+1:]...)
				}
			}
		}
		record.UnReceiver = strings.Join(receivers, ",")
		if err := s.db.UpdateRecord(&record); err != nil {
			s.log.Errorf("remove uc sync record send failed receiver err: %v", err)
		}
	}

	return nil
}

func getLock(key string) (*dlock.DLock, error) {
	lock, err := dlock.New(key, func() {})
	if err != nil {
		return nil, err
	}
	if err := lock.Lock(context.Background()); err != nil {
		return nil, err
	}

	return lock, nil
}
