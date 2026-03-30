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

package archive

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	auditmodel "github.com/erda-project/erda/internal/apps/ai-proxy/models/audit"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

var archiveCSVHeader = []string{
	"id",
	"created_at",
	"updated_at",
	"deleted_at",
	"request_at",
	"response_at",
	"auth_key",
	"status",
	"prompt",
	"completion",
	"request_body",
	"response_body",
	"actual_request_body",
	"actual_response_body",
	"user_agent",
	"x_request_id",
	"call_id",
	"client_id",
	"model_id",
	"session_id",
	"username",
	"email",
	"source",
	"operation_id",
	"res_func_call_name",
	"metadata",
}

func (s *Service) Run(ctx context.Context) error {
	if !s.Config.Enable {
		return nil
	}
	if err := s.Config.Normalize(); err != nil {
		return err
	}
	if err := s.markInterruptedIfNeeded(ctx); err != nil {
		s.logf("mark interrupted archive day failed: %v", err)
	}
	if err := s.tick(ctx); err != nil {
		s.logf("initial archive tick failed: %v", err)
	}

	ticker := time.NewTicker(s.Config.LoopInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.tick(ctx); err != nil {
				s.logf("archive tick failed: %v", err)
			}
		}
	}
}

func (s *Service) markInterruptedIfNeeded(ctx context.Context) error {
	latestStart, err := s.EventClient.LatestByEvent(ctx, EventArchiveDayStart)
	if err != nil || latestStart == nil {
		return err
	}
	latestEnd, err := s.EventClient.LatestByEvent(ctx, EventArchiveDayEnd)
	if err != nil {
		return err
	}
	if !isAfter(latestStart, latestEnd) {
		return nil
	}
	if _, err = s.EventClient.Create(ctx, EventArchiveDayInterrupted, latestStart.Detail); err != nil {
		return err
	}
	_, err = s.EventClient.Create(ctx, EventArchiveDayEnd, latestStart.Detail)
	return err
}

func (s *Service) tick(ctx context.Context) error {
	status, err := s.GetStatus(ctx)
	if err != nil {
		return err
	}
	if !status.Enabled || !status.Started {
		return nil
	}

	cutoff := archiveDayStart(time.Now()).AddDate(0, 0, -s.Config.RetentionDays)
	successEvent, err := s.EventClient.LatestByEvent(ctx, EventArchiveDaySuccess)
	if err != nil {
		return err
	}

	if successEvent != nil {
		successDay, err := parseArchiveDay(successEvent.Detail)
		if err != nil {
			return err
		}
		objectExists, err := s.objectExists(ctx, successDay)
		if err != nil {
			return err
		}
		if !objectExists {
			return s.archiveDay(ctx, successDay)
		}

		hasRows, err := s.AuditClient.HasRowsInRange(ctx, successDay, successDay.Add(24*time.Hour))
		if err != nil {
			return err
		}
		if hasRows {
			_, err = s.AuditClient.DeleteArchiveBatch(ctx, successDay, successDay.Add(24*time.Hour), s.Config.BatchSize)
			return err
		}

		nextDay := successDay.Add(24 * time.Hour)
		if !nextDay.Before(cutoff) {
			return nil
		}
		return s.archiveDay(ctx, nextDay)
	}

	oldestDay, ok, err := s.AuditClient.OldestDayBefore(ctx, cutoff)
	if err != nil || !ok {
		return err
	}
	return s.archiveDay(ctx, oldestDay)
}

func (s *Service) archiveDay(ctx context.Context, day time.Time) error {
	dayStr := day.Format("2006-01-02")
	if _, err := s.EventClient.Create(ctx, EventArchiveDayStart, dayStr); err != nil {
		return err
	}

	if err := s.exportDay(ctx, day); err != nil {
		_ = s.writeEndEvents(ctx, EventArchiveDayFailed, dayStr)
		return err
	}

	exists, err := s.objectExists(ctx, day)
	if err != nil {
		_ = s.writeEndEvents(ctx, EventArchiveDayFailed, dayStr)
		return err
	}
	if !exists {
		_ = s.writeEndEvents(ctx, EventArchiveDayFailed, dayStr)
		return fmt.Errorf("archive object not found for day %s", dayStr)
	}

	if err := s.writeEndEvents(ctx, EventArchiveDaySuccess, dayStr); err != nil {
		return err
	}

	_, err = s.AuditClient.DeleteArchiveBatch(ctx, day, day.Add(24*time.Hour), s.Config.BatchSize)
	return err
}

func (s *Service) writeEndEvents(ctx context.Context, resultEvent, day string) error {
	if _, err := s.EventClient.Create(ctx, resultEvent, day); err != nil {
		return err
	}
	_, err := s.EventClient.Create(ctx, EventArchiveDayEnd, day)
	return err
}

func (s *Service) exportDay(ctx context.Context, day time.Time) error {
	start := day
	end := day.Add(24 * time.Hour)

	tmpFile, err := os.CreateTemp("", "ai-proxy-audit-archive-*.csv.gz")
	if err != nil {
		return err
	}
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	gzipWriter := gzip.NewWriter(tmpFile)
	csvWriter := csv.NewWriter(gzipWriter)
	if err := csvWriter.Write(archiveCSVHeader); err != nil {
		return err
	}

	var afterCreatedAt *time.Time
	var afterID string
	for {
		list, err := s.AuditClient.ListArchiveBatch(ctx, start, end, afterCreatedAt, afterID, s.Config.BatchSize)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			break
		}
		for _, rec := range list {
			if err := csvWriter.Write(toArchiveCSVRow(rec)); err != nil {
				return err
			}
		}
		last := list[len(list)-1]
		lastCreatedAt := last.CreatedAt
		afterCreatedAt = &lastCreatedAt
		afterID = last.ID.String
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		return err
	}
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return s.putObject(day, tmpFile)
}

func (s *Service) objectKey(day time.Time) string {
	return fmt.Sprintf("ai-proxy/%s/audit/archive/%04d/%02d/audit-%s.csv.gz",
		s.Config.Name, day.Year(), int(day.Month()), day.Format("2006-01-02"))
}

func (s *Service) objectExists(_ context.Context, day time.Time) (bool, error) {
	bucket, err := s.newBucket()
	if err != nil {
		return false, err
	}
	return bucket.IsObjectExist(s.objectKey(day))
}

func (s *Service) putObject(day time.Time, reader io.Reader) error {
	bucket, err := s.newBucket()
	if err != nil {
		return err
	}
	return bucket.PutObject(s.objectKey(day), reader)
}

func (s *Service) newBucket() (*oss.Bucket, error) {
	client, err := oss.New(s.Config.OSS.Endpoint, s.Config.OSS.AccessKeyID, s.Config.OSS.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	return client.Bucket(s.Config.OSS.Bucket)
}

func parseArchiveDay(v string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", strings.TrimSpace(v), time.Local)
}

func toArchiveCSVRow(rec *auditmodel.Audit) []string {
	return []string{
		rec.ID.String,
		formatTime(rec.CreatedAt),
		formatTime(rec.UpdatedAt),
		formatDeletedAt(rec.DeletedAt.Time, rec.DeletedAt.Valid),
		formatTime(rec.RequestAt),
		formatTime(rec.ResponseAt),
		rec.AuthKey,
		fmt.Sprintf("%d", rec.Status),
		rec.Prompt,
		rec.Completion,
		rec.RequestBody,
		rec.ResponseBody,
		rec.ActualRequestBody,
		rec.ActualResponseBody,
		rec.UserAgent,
		rec.XRequestID,
		rec.CallID,
		rec.ClientID,
		rec.ModelID,
		rec.SessionID,
		rec.Username,
		rec.Email,
		rec.BizSource,
		rec.OperationID,
		rec.ResponseFunctionCallName,
		mustMetadataJSON(rec.Metadata),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339Nano)
}

func formatDeletedAt(t time.Time, valid bool) string {
	if !valid || t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339Nano)
}

func mustMetadataJSON(m metadata.Metadata) string {
	s, err := m.ToJson()
	if err != nil {
		return ""
	}
	return s
}

func (s *Service) logf(format string, args ...any) {
	if s.Logger != nil {
		s.Logger.Infof(format, args...)
	}
}
