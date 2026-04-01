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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	auditmodel "github.com/erda-project/erda/internal/apps/ai-proxy/models/audit"
	eventmodel "github.com/erda-project/erda/internal/apps/ai-proxy/models/event"
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

const archiveSplitCheckStride = 100

var archiveObjectExists = func(s *Service, _ context.Context, objectKey string) (bool, error) {
	bucket, err := s.newBucket()
	if err != nil {
		return false, err
	}
	return bucket.IsObjectExist(objectKey)
}

var archivePutObject = func(s *Service, objectKey string, reader io.Reader) error {
	bucket, err := s.newBucket()
	if err != nil {
		return err
	}
	return bucket.PutObject(objectKey, reader)
}

var archiveDeleteBatch = func(s *Service, ctx context.Context, day time.Time) (int64, error) {
	return s.AuditClient.DeleteArchiveBatch(ctx, day, day.Add(24*time.Hour), s.Config.BatchSize)
}

type archiveEventDetail struct {
	Day                 string              `json:"day"`
	RowCount            int64               `json:"row_count"`
	RawSizeBytes        int64               `json:"raw_size_bytes"`
	CompressedSizeBytes int64               `json:"compressed_size_bytes"`
	Parts               []archivePartDetail `json:"parts,omitempty"`
	Error               string              `json:"error,omitempty"`
}

type archivePartDetail struct {
	Index               int    `json:"index"`
	ObjectKey           string `json:"object_key"`
	RowCount            int64  `json:"row_count"`
	RawSizeBytes        int64  `json:"raw_size_bytes"`
	CompressedSizeBytes int64  `json:"compressed_size_bytes"`
}

type archiveExportStats struct {
	Day                 string
	RowCount            int64
	RawSizeBytes        int64
	CompressedSizeBytes int64
	Parts               []archivePartDetail
}

type countingWriter struct {
	w io.Writer
	n int64
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.n += int64(n)
	return n, err
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
	// try to acquire leadership via optimistic lock on heartbeat event
	ok, err := s.EventClient.TryAcquireLeaderLease(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

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
	dryRunDayEvent := (*eventmodel.Event)(nil)
	if status.DryRun {
		dryRunDayEvent, err = s.EventClient.LatestByEvent(ctx, EventArchiveDayDryRun)
		if err != nil {
			return err
		}
	}

	if successEvent != nil {
		successDay, err := parseArchiveDay(successEvent.Detail)
		if err != nil {
			return err
		}
		objectExists, err := s.objectExists(ctx, successDay, successEvent.Detail)
		if err != nil {
			return err
		}
		if !objectExists {
			if status.DryRun {
				return s.dryRunDay(ctx, successDay, "would re-export because archive object is missing")
			}
			return s.archiveDay(ctx, successDay)
		}

		hasRows, err := s.AuditClient.HasRowsInRange(ctx, successDay, successDay.Add(24*time.Hour))
		if err != nil {
			return err
		}
		if hasRows {
			if status.DryRun {
				return s.dryRunDay(ctx, successDay, "would delete archived audit rows")
			}
			return s.deleteArchivedDayRows(ctx, successDay)
		}

		nextDay := successDay.Add(24 * time.Hour)
		if status.DryRun && dryRunDayEvent != nil {
			dryRunDay, err := parseArchiveDay(dryRunDayEvent.Detail)
			if err != nil {
				return err
			}
			if dryRunDay.After(successDay) {
				nextDay = dryRunDay.Add(24 * time.Hour)
			}
		}
		if !nextDay.Before(cutoff) {
			return nil
		}
		if status.DryRun {
			return s.dryRunDay(ctx, nextDay, "would export archive CSV and upload to OSS")
		}
		return s.archiveDay(ctx, nextDay)
	}

	oldestDay, ok, err := s.AuditClient.OldestDayBefore(ctx, cutoff)
	if err != nil || !ok {
		return err
	}
	if status.DryRun && dryRunDayEvent != nil {
		dryRunDay, err := parseArchiveDay(dryRunDayEvent.Detail)
		if err != nil {
			return err
		}
		nextDay := dryRunDay.Add(24 * time.Hour)
		if nextDay.Before(cutoff) {
			return s.dryRunDay(ctx, nextDay, "would export archive CSV and upload to OSS")
		}
		return nil
	}
	if status.DryRun {
		return s.dryRunDay(ctx, oldestDay, "would export archive CSV and upload to OSS")
	}
	return s.archiveDay(ctx, oldestDay)
}

func (s *Service) archiveDay(ctx context.Context, day time.Time) (err error) {
	dayStr := day.Format("2006-01-02")
	defer func() {
		if r := recover(); r != nil {
			reason := fmt.Sprintf("panic: %v", r)
			s.logf("archive day panic day=%s reason=%s stack=%s", dayStr, reason, strings.TrimSpace(string(debug.Stack())))
			err = s.failDay(ctx, dayStr, archiveExportStats{}, fmt.Errorf("%s", reason))
		}
	}()
	if _, err = s.EventClient.Create(ctx, EventArchiveDayStart, dayStr); err != nil {
		return err
	}

	stats, err := s.exportDay(ctx, day)
	if err != nil {
		return s.failDay(ctx, dayStr, stats, err)
	}

	var exists bool
	exists, err = s.objectExists(ctx, day, mustArchiveEventDetailJSON(archiveEventDetail{
		Day:                 dayStr,
		RowCount:            stats.RowCount,
		RawSizeBytes:        stats.RawSizeBytes,
		CompressedSizeBytes: stats.CompressedSizeBytes,
		Parts:               stats.Parts,
	}))
	if err != nil {
		return s.failDay(ctx, dayStr, stats, err)
	}
	if !exists {
		return s.failDay(ctx, dayStr, stats, fmt.Errorf("archive object not found for day %s", dayStr))
	}

	if err = s.deleteArchivedDayRows(ctx, day); err != nil {
		return s.failDay(ctx, dayStr, stats, err)
	}

	if err := s.writeEndEvents(ctx, EventArchiveDaySuccess, mustArchiveEventDetailJSON(archiveEventDetail{
		Day:                 dayStr,
		RowCount:            stats.RowCount,
		RawSizeBytes:        stats.RawSizeBytes,
		CompressedSizeBytes: stats.CompressedSizeBytes,
		Parts:               stats.Parts,
	})); err != nil {
		return err
	}
	return nil
}

func (s *Service) dryRunDay(ctx context.Context, day time.Time, action string) error {
	dayStr := day.Format("2006-01-02")
	if _, err := s.EventClient.Create(ctx, EventArchiveDayStart, dayStr); err != nil {
		return err
	}

	rowCount, err := s.scanArchiveDay(ctx, day)
	if err != nil {
		_ = s.failDay(ctx, dayStr, archiveExportStats{}, err)
		return err
	}

	s.logf("audit archive dry-run day=%s action=%s object_key_pattern=%s rows=%d", dayStr, action, s.objectKeyPattern(day), rowCount)
	if err := s.writeEndEvents(ctx, EventArchiveDayDryRun, dayStr); err != nil {
		return err
	}
	return nil
}

func (s *Service) writeEndEvents(ctx context.Context, resultEvent, day string) error {
	if _, err := s.EventClient.Create(ctx, resultEvent, day); err != nil {
		return err
	}
	_, err := s.EventClient.Create(ctx, EventArchiveDayEnd, archiveDayFromDetail(day))
	return err
}

func (s *Service) failDay(ctx context.Context, day string, stats archiveExportStats, cause error) error {
	detail := archiveEventDetail{Day: archiveDayFromDetail(day)}
	if cause != nil {
		detail.Error = strings.TrimSpace(cause.Error())
	}
	detail.RowCount = stats.RowCount
	detail.RawSizeBytes = stats.RawSizeBytes
	detail.CompressedSizeBytes = stats.CompressedSizeBytes
	detail.Parts = stats.Parts
	if err := s.writeEndEvents(ctx, EventArchiveDayFailed, mustArchiveEventDetailJSON(detail)); err != nil {
		if cause == nil {
			return err
		}
		return fmt.Errorf("%w; write failed event: %v", cause, err)
	}
	return cause
}

func (s *Service) exportDay(ctx context.Context, day time.Time) (archiveExportStats, error) {
	start := day
	end := day.Add(24 * time.Hour)
	stats := archiveExportStats{Day: day.Format("2006-01-02")}

	var afterCreatedAt *time.Time
	var afterID string
	partIndex := 1
	part, err := newArchivePartFile()
	if err != nil {
		return stats, err
	}
	defer func() { part.cleanup() }()
	list, err := s.AuditClient.ListArchiveBatch(ctx, start, end, afterCreatedAt, afterID, s.Config.BatchSize)
	if err != nil {
		return stats, err
	}
	for len(list) > 0 {
		splitAfterBatch := false
		for i, rec := range list {
			if err := part.writeRow(toArchiveCSVRow(rec)); err != nil {
				return stats, err
			}
			if part.rowCount%archiveSplitCheckStride != 0 {
				continue
			}
			_, compressedSize, err := part.syncStats()
			if err != nil {
				return stats, err
			}
			if compressedSize < s.Config.MaxCompressedFileSizeBytes {
				continue
			}
			if i < len(list)-1 {
				oldPart := part
				partStats, err := s.finalizeArchivePart(day, partIndex, true, oldPart)
				if err != nil {
					return stats, err
				}
				oldPart.cleanup()
				stats.addPart(partStats)
				partIndex++
				part, err = newArchivePartFile()
				if err != nil {
					return stats, err
				}
				continue
			}
			splitAfterBatch = true
		}

		last := list[len(list)-1]
		lastCreatedAt := last.CreatedAt
		afterCreatedAt = &lastCreatedAt
		afterID = last.ID.String

		nextList, err := s.AuditClient.ListArchiveBatch(ctx, start, end, afterCreatedAt, afterID, s.Config.BatchSize)
		if err != nil {
			return stats, err
		}

		if splitAfterBatch {
			oldPart := part
			partStats, err := s.finalizeArchivePart(day, partIndex, true, oldPart)
			if err != nil {
				return stats, err
			}
			oldPart.cleanup()
			stats.addPart(partStats)
			partIndex++
			part, err = newArchivePartFile()
			if err != nil {
				return stats, err
			}
		}
		list = nextList
	}

	if len(stats.Parts) == 0 || part.rowCount > 0 {
		partStats, err := s.finalizeArchivePart(day, partIndex, len(stats.Parts) > 0, part)
		if err != nil {
			return stats, err
		}
		stats.addPart(partStats)
	}
	return stats, nil
}

func (s *Service) scanArchiveDay(ctx context.Context, day time.Time) (int64, error) {
	start := day
	end := day.Add(24 * time.Hour)

	var (
		rowCount       int64
		afterCreatedAt *time.Time
		afterID        string
	)
	for {
		list, err := s.AuditClient.ListArchiveBatch(ctx, start, end, afterCreatedAt, afterID, s.Config.BatchSize)
		if err != nil {
			return 0, err
		}
		if len(list) == 0 {
			break
		}
		rowCount += int64(len(list))
		last := list[len(list)-1]
		lastCreatedAt := last.CreatedAt
		afterCreatedAt = &lastCreatedAt
		afterID = last.ID.String
	}
	return rowCount, nil
}

func (s *Service) objectKey(day time.Time, partIndex int, multipart bool) string {
	if !multipart {
		return fmt.Sprintf("ai-proxy/%s/audit/archive/%04d/%02d/audit-%s.csv.gz",
			s.Config.Name, day.Year(), int(day.Month()), day.Format("2006-01-02"))
	}
	return fmt.Sprintf("ai-proxy/%s/audit/archive/%04d/%02d/audit-%s_%d.csv.gz",
		s.Config.Name, day.Year(), int(day.Month()), day.Format("2006-01-02"), partIndex)
}

func (s *Service) objectKeyPattern(day time.Time) string {
	return fmt.Sprintf("ai-proxy/%s/audit/archive/%04d/%02d/audit-%s*.csv.gz",
		s.Config.Name, day.Year(), int(day.Month()), day.Format("2006-01-02"))
}

func (s *Service) objectKeysFromDetail(day time.Time, detail string) []string {
	payload, ok := parseArchiveEventDetail(detail)
	if ok && len(payload.Parts) > 0 {
		keys := make([]string, 0, len(payload.Parts))
		for _, part := range payload.Parts {
			if strings.TrimSpace(part.ObjectKey) == "" {
				continue
			}
			keys = append(keys, strings.TrimSpace(part.ObjectKey))
		}
		if len(keys) > 0 {
			return keys
		}
	}
	return []string{s.objectKey(day, 1, false)}
}

func (s *Service) objectExists(ctx context.Context, day time.Time, detail string) (bool, error) {
	keys := s.objectKeysFromDetail(day, detail)
	for _, objectKey := range keys {
		exists, err := archiveObjectExists(s, ctx, objectKey)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}
	return true, nil
}

func (s *Service) putObject(objectKey string, reader io.Reader) error {
	return archivePutObject(s, objectKey, reader)
}

func (s *Service) deleteArchivedDayRows(ctx context.Context, day time.Time) error {
	for {
		rows, err := archiveDeleteBatch(s, ctx, day)
		if err != nil {
			return err
		}
		if rows == 0 {
			return nil
		}
	}
}

func (s *Service) newBucket() (*oss.Bucket, error) {
	if err := s.validateOSSConfig(); err != nil {
		return nil, err
	}
	client, err := oss.New(s.Config.OSS.Endpoint, s.Config.OSS.AccessKeyID, s.Config.OSS.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	return client.Bucket(s.Config.OSS.Bucket)
}

func (s *Service) validateOSSConfig() error {
	var missing []string
	if strings.TrimSpace(s.Config.OSS.Endpoint) == "" {
		missing = append(missing, "AI_PROXY_AUDIT_ARCHIVE_OSS_ENDPOINT")
	}
	if strings.TrimSpace(s.Config.OSS.AccessKeyID) == "" {
		missing = append(missing, "AI_PROXY_AUDIT_ARCHIVE_OSS_ACCESS_KEY_ID")
	}
	if strings.TrimSpace(s.Config.OSS.AccessKeySecret) == "" {
		missing = append(missing, "AI_PROXY_AUDIT_ARCHIVE_OSS_ACCESS_KEY_SECRET")
	}
	if strings.TrimSpace(s.Config.OSS.Bucket) == "" {
		missing = append(missing, "AI_PROXY_AUDIT_ARCHIVE_OSS_BUCKET")
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("missing OSS archive config: %s", strings.Join(missing, ", "))
}

func parseArchiveDay(v string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", archiveDayFromDetail(v), time.Local)
}

func mustArchiveEventDetailJSON(detail archiveEventDetail) string {
	raw, err := json.Marshal(detail)
	if err != nil {
		return fmt.Sprintf("{\"day\":%q,\"error\":%q}", detail.Day, fmt.Sprintf("marshal archive detail failed: %v", err))
	}
	return string(raw)
}

func parseArchiveEventDetail(detail string) (archiveEventDetail, bool) {
	var payload archiveEventDetail
	if err := json.Unmarshal([]byte(strings.TrimSpace(detail)), &payload); err != nil {
		return archiveEventDetail{}, false
	}
	if strings.TrimSpace(payload.Day) == "" {
		return archiveEventDetail{}, false
	}
	return payload, true
}

type archivePartFile struct {
	tmpFile    *os.File
	gzipWriter *gzip.Writer
	rawCounter *countingWriter
	csvWriter  *csv.Writer
	rowCount   int64
}

func newArchivePartFile() (*archivePartFile, error) {
	tmpFile, err := os.CreateTemp("", "ai-proxy-audit-archive-*.csv.gz")
	if err != nil {
		return nil, err
	}
	gzipWriter := gzip.NewWriter(tmpFile)
	rawCounter := &countingWriter{w: gzipWriter}
	csvWriter := csv.NewWriter(rawCounter)
	if err := csvWriter.Write(archiveCSVHeader); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return nil, err
	}
	return &archivePartFile{
		tmpFile:    tmpFile,
		gzipWriter: gzipWriter,
		rawCounter: rawCounter,
		csvWriter:  csvWriter,
	}, nil
}

func (p *archivePartFile) writeRow(row []string) error {
	if err := p.csvWriter.Write(row); err != nil {
		return err
	}
	p.rowCount++
	return nil
}

func (p *archivePartFile) syncStats() (int64, int64, error) {
	p.csvWriter.Flush()
	if err := p.csvWriter.Error(); err != nil {
		return 0, 0, err
	}
	if err := p.gzipWriter.Flush(); err != nil {
		return 0, 0, err
	}
	info, err := p.tmpFile.Stat()
	if err != nil {
		return 0, 0, err
	}
	return p.rawCounter.n, info.Size(), nil
}

func (p *archivePartFile) close() (int64, int64, error) {
	p.csvWriter.Flush()
	if err := p.csvWriter.Error(); err != nil {
		return 0, 0, err
	}
	rawSize := p.rawCounter.n
	if err := p.gzipWriter.Close(); err != nil {
		return 0, 0, err
	}
	info, err := p.tmpFile.Stat()
	if err != nil {
		return 0, 0, err
	}
	return rawSize, info.Size(), nil
}

func (p *archivePartFile) cleanup() {
	if p == nil || p.tmpFile == nil {
		return
	}
	_ = p.tmpFile.Close()
	_ = os.Remove(p.tmpFile.Name())
}

func (s *Service) finalizeArchivePart(day time.Time, partIndex int, multipart bool, part *archivePartFile) (archivePartDetail, error) {
	rawSize, compressedSize, err := part.close()
	if err != nil {
		return archivePartDetail{}, err
	}
	if _, err := part.tmpFile.Seek(0, io.SeekStart); err != nil {
		return archivePartDetail{}, err
	}
	objectKey := s.objectKey(day, partIndex, multipart)
	if err := s.putObject(objectKey, part.tmpFile); err != nil {
		return archivePartDetail{}, err
	}
	return archivePartDetail{
		Index:               partIndex,
		ObjectKey:           objectKey,
		RowCount:            part.rowCount,
		RawSizeBytes:        rawSize,
		CompressedSizeBytes: compressedSize,
	}, nil
}

func (s *archiveExportStats) addPart(part archivePartDetail) {
	s.Parts = append(s.Parts, part)
	s.RowCount += part.RowCount
	s.RawSizeBytes += part.RawSizeBytes
	s.CompressedSizeBytes += part.CompressedSizeBytes
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
