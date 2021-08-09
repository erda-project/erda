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

package apis

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda-proto-go/msp/apm/checker/pb"
	projectpb "github.com/erda-project/erda-proto-go/msp/tenant/project/pb"
	"github.com/erda-project/erda/modules/msp/apm/checker/storage/cache"
	"github.com/erda-project/erda/modules/msp/apm/checker/storage/db"
	"github.com/erda-project/erda/pkg/common/errors"
)

type checkerV1Service struct {
	projectDB     *db.ProjectDB
	metricDB      *db.MetricDB
	cache         *cache.Cache
	metricq       metricpb.MetricServiceServer
	projectServer projectpb.ProjectServiceServer
}

func (s *checkerV1Service) CreateCheckerV1(ctx context.Context, req *pb.CreateCheckerV1Request) (*pb.CreateCheckerV1Response, error) {
	if req.Data == nil {
		return nil, errors.NewMissingParameterError("data")
	}
	if req.Data.ProjectID <= 0 {
		return nil, errors.NewMissingParameterError("projectId")
	}
	extra := strconv.FormatInt(req.Data.ProjectID, 10)
	now := time.Now()
	m := &db.Metric{
		ProjectID:  req.Data.ProjectID,
		Env:        req.Data.Env,
		Name:       req.Data.Name,
		Mode:       req.Data.Mode,
		Extra:      extra,
		URL:        req.Data.Url,
		CreateTime: now,
		UpdateTime: now,
	}
	if err := s.metricDB.Create(m); err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	checker := s.ConvertToChecker(ctx, m, req.Data.ProjectID)
	if checker != nil {
		err := s.cache.Put(checker)
		if err != nil {
			return nil, err
		}
	}
	return &pb.CreateCheckerV1Response{Data: m.ID}, nil
}

func (s *checkerV1Service) ConvertToChecker(ctx context.Context, m *db.Metric, projectID int64) *pb.Checker {
	ck := &pb.Checker{
		Id:   m.ID,
		Name: m.Name,
		Type: m.Mode,
		Config: map[string]string{
			"url": m.URL,
		},
		Tags: map[string]string{
			"_metric_scope": "micro_service",
			"project_id":    strconv.FormatInt(projectID, 10),
			"env":           m.Env,
			"metric":        strconv.FormatInt(m.ID, 10),
			"metric_name":   m.Name,
		},
	}
	response, err := s.projectServer.GetProject(ctx, &projectpb.GetProjectRequest{ProjectID: strconv.FormatInt(projectID, 10)})
	if err == nil && response != nil {
		project := response.Data
		if project != nil && len(project.Relationship) > 0 {
			var relationship *projectpb.TenantRelationship
			for _, r := range project.Relationship {
				if r.Workspace == m.Env {
					relationship = r
				}
			}
			if relationship != nil {
				s.addTenantTags(ck, relationship.TenantID, project.Name)
				return ck
			}
		}
		return nil
	}

	scopeInfo, err := s.metricDB.QueryScopeInfo(projectID, m.Env)
	if err == nil && scopeInfo != nil {
		s.addTenantTags(ck, scopeInfo.ScopeID, scopeInfo.ProjectName)
	}
	return ck
}

func (s *checkerV1Service) addTenantTags(ck *pb.Checker, tenantId, projectName string) {
	ck.Tags["_metric_scope_id"] = tenantId
	ck.Tags["terminus_key"] = tenantId
	ck.Tags["project_name"] = projectName
}

func (s *checkerV1Service) UpdateCheckerV1(ctx context.Context, req *pb.UpdateCheckerV1Request) (*pb.UpdateCheckerV1Response, error) {
	if req.Data == nil {
		return nil, errors.NewMissingParameterError("data")
	}
	metric, err := s.metricDB.GetByID(req.Id)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	if metric == nil {
		return nil, errors.NewNotFoundError(fmt.Sprintf("metric/%d", req.Id))
	}
	metric.URL = req.Data.Url
	metric.Name = req.Data.Name
	if err := s.metricDB.Update(metric); err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	checker := s.ConvertToChecker(ctx, metric, metric.ProjectID)
	if checker != nil {
		err := s.cache.Put(checker)
		if err != nil {
			return nil, err
		}
	}
	return &pb.UpdateCheckerV1Response{Data: req.Id}, nil
}

func (s *checkerV1Service) DeleteCheckerV1(ctx context.Context, req *pb.DeleteCheckerV1Request) (*pb.DeleteCheckerV1Response, error) {
	metric, err := s.metricDB.GetByID(req.Id)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	if metric == nil {
		return &pb.DeleteCheckerV1Response{}, nil
	}

	var projectID int64
	proj, err := s.projectDB.GetByID(metric.ProjectID)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	if proj != nil {
		projectID = proj.ProjectID
	}

	err = s.metricDB.Delete(req.Id)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	err = s.cache.Remove(req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteCheckerV1Response{Data: &pb.CheckerV1{
		Name:      metric.Name,
		Mode:      metric.Mode,
		Url:       metric.URL,
		ProjectID: projectID,
		Env:       metric.Env,
	}}, nil
}

func (s *checkerV1Service) GetCheckerV1(ctx context.Context, req *pb.GetCheckerV1Request) (*pb.GetCheckerV1Response, error) {
	metric, err := s.metricDB.GetByID(req.Id)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	if metric == nil {
		return &pb.GetCheckerV1Response{}, nil
	}
	return &pb.GetCheckerV1Response{
		Data: &pb.CheckerV1{
			Name:      metric.Name,
			Mode:      metric.Mode,
			Url:       metric.URL,
			ProjectID: metric.ProjectID,
			Env:       metric.Env,
		},
	}, nil
}

func (s *checkerV1Service) DescribeCheckersV1(ctx context.Context, req *pb.DescribeCheckersV1Request) (*pb.DescribeCheckersV1Response, error) {
	proj, err := s.projectDB.GetByProjectID(req.ProjectID)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	var list []*db.Metric
	if proj != nil {
		// history record
		oldCheckers, err := s.metricDB.ListByProjectIDAndEnv(proj.ID, req.Env)
		for _, checker := range oldCheckers {
			if checker.Extra == "" {
				list = append(list, checker)
			}
		}
		if err != nil {
			return nil, errors.NewDatabaseError(err)
		}
		newCheckers, err := s.metricDB.ListByProjectIDAndEnv(req.ProjectID, req.Env)
		if err != nil {
			return nil, errors.NewDatabaseError(err)
		}
		for _, checker := range newCheckers {
			if checker.Extra != "" {
				extra, err := strconv.ParseInt(checker.Extra, 10, 64)
				if err != nil {
					return nil, errors.NewDatabaseError(err)
				}
				if checker.ProjectID == extra {
					list = append(list, checker)
				}
			}
		}
	} else {
		list, err = s.metricDB.ListByProjectIDAndEnv(req.ProjectID, req.Env)
		if err != nil {
			return nil, errors.NewDatabaseError(err)
		}
	}

	results := make(map[int64]*pb.DescribeItemV1)
	for _, item := range list {
		result := &pb.DescribeItemV1{
			Name:   item.Name,
			Mode:   item.Mode,
			Url:    item.URL,
			Status: statusMiss,
		}
		results[item.ID] = result
	}
	err = s.queryCheckersLatencySummaryByProject(req.ProjectID, results)
	if err != nil {
		return nil, errors.NewServiceInvokingError(fmt.Sprintf("status_page.project_id/%d", req.ProjectID), err)
	}
	var downCount int64
	for _, item := range results {
		if item.Status == statusRED {
			downCount++
		}
	}
	return &pb.DescribeCheckersV1Response{
		Data: &pb.DescribeResultV1{
			DownCount: downCount,
			Metrics:   results,
		},
	}, nil
}

func (s *checkerV1Service) DescribeCheckerV1(ctx context.Context, req *pb.DescribeCheckerV1Request) (*pb.DescribeCheckerV1Response, error) {
	metric, err := s.metricDB.GetByID(req.Id)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	results := make(map[int64]*pb.DescribeItemV1)
	var downCount int64
	if metric != nil {
		results[req.Id] = &pb.DescribeItemV1{
			Name:   metric.Name,
			Mode:   metric.Mode,
			Url:    metric.URL,
			Status: statusMiss,
		}
		err = s.queryCheckersLatencySummary(req.Id, req.Period, results)
		if err != nil {
			return nil, errors.NewServiceInvokingError(fmt.Sprintf("status_page.metric/%d", req.Id), err)
		}
		for _, item := range results {
			if item.Status == statusRED {
				downCount++
			}
		}
	}
	return &pb.DescribeCheckerV1Response{
		Data: &pb.DescribeResultV1{
			DownCount: downCount,
			Metrics:   results,
		},
	}, nil
}

func getTimeRange(unit string, num int, align bool) (start int64, end int64, interval string) {
	now := time.Now()
	alignTime := func(interval string) time.Time {
		if align {
			ts := now.UnixNano()
			switch interval {
			case "24h":
				const day = 24 * int64(time.Hour)
				if ts%day != 0 {
					ts = ts - ts%day + day
					now = time.Unix(ts/int64(time.Second), ts%int64(time.Second))
				}
			case "1h":
				const hour = int64(time.Hour)
				if ts%hour != 0 {
					ts = ts - ts%hour + hour
					now = time.Unix(ts/int64(time.Second), ts%int64(time.Second))
				}
			case "1m":
				const minute = int64(time.Minute)
				if ts%minute != 0 {
					ts = ts - ts%minute + minute
					now = time.Unix(ts/int64(time.Second), ts%int64(time.Second))
				}
			}
		}
		return now
	}
	switch strings.ToLower(unit) {
	case "year":
		interval = "24h"
		now = alignTime(interval)
		now = now.AddDate(0, 0, 1)
	case "month":
		interval = "24h"
		now = alignTime(interval)
		start = now.AddDate(0, -1*num, 0).UnixNano() / int64(time.Millisecond)
	case "week":
		interval = "24h"
		now = alignTime(interval)
		start = now.AddDate(0, 0, -7*num).UnixNano() / int64(time.Millisecond)
	case "day":
		interval = "1h"
		now = alignTime(interval)
		start = now.AddDate(0, 0, -1*num).UnixNano() / int64(time.Millisecond)
	default:
		interval = "1m"
		now = alignTime(interval)
		start = now.Add(time.Duration(-1*int64(num)*int64(time.Hour))).UnixNano() / int64(time.Millisecond)
	}
	end = now.UnixNano() / int64(time.Millisecond)
	return
}

func (s *checkerV1Service) queryCheckersLatencySummaryByProject(projectID int64, metrics map[int64]*pb.DescribeItemV1) error {
	start, end, duration := getTimeRange("hour", 1, false)
	interval, _ := structpb.NewValue(map[string]interface{}{"duration": duration})
	return s.queryCheckerMetrics(start, end, `
	SELECT timestamp(), metric::tag, status_name::tag, round_float(avg(latency),2), max(latency), min(latency), count(latency), sum(latency)
	FROM status_page 
	WHERE project_id=$projectID 
	GROUP BY time($interval), metric::tag, status_name::tag 
	LIMIT 200`,
		map[string]*structpb.Value{
			"projectID": structpb.NewStringValue(strconv.FormatInt(projectID, 10)),
			"interval":  interval,
		}, metrics,
	)
}

func (s *checkerV1Service) queryCheckersLatencySummary(metricID int64, timeUnit string, metrics map[int64]*pb.DescribeItemV1) error {
	start, end, duration := getTimeRange(timeUnit, 1, false)
	interval, _ := structpb.NewValue(map[string]interface{}{"duration": duration})
	return s.queryCheckerMetrics(start, end, `
	SELECT timestamp(), metric::tag, status_name::tag, round_float(avg(latency),2), max(latency), min(latency), count(latency), sum(latency)
	FROM status_page 
	WHERE metric=$metric 
	GROUP BY time($interval), metric::tag, status_name::tag 
	LIMIT 200`,
		map[string]*structpb.Value{
			"metric":   structpb.NewStringValue(strconv.FormatInt(metricID, 10)),
			"interval": interval,
		}, metrics,
	)
}

func (s *checkerV1Service) queryCheckerMetrics(start, end int64, statement string, params map[string]*structpb.Value, metrics map[int64]*pb.DescribeItemV1) error {
	req := &metricpb.QueryWithInfluxFormatRequest{
		Start:     strconv.FormatInt(start, 10),
		End:       strconv.FormatInt(end, 10),
		Statement: statement,
		Params:    params,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	resp, err := s.metricq.QueryWithInfluxFormat(ctx, req)
	if err != nil {
		return err
	}
	s.parseMetricSummaryResponse(resp, metrics)
	return nil
}

const (
	statusRED   = "Major Outage"
	statusGreen = "Operational"
	statusMiss  = "Miss"
)

func (s *checkerV1Service) parseMetricSummaryResponse(resp *metricpb.QueryWithInfluxFormatResponse, metrics map[int64]*pb.DescribeItemV1) {
	type summaryItem struct {
		time  []int64
		avg   []float64
		max   []float64
		min   []float64
		sum   []float64
		count []int64
	}
	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {
		summary := make(map[string]map[string]*summaryItem)
		serie := resp.Results[0].Series[0]
		groupedRows := groupSerieRows(serie.Rows, 2, 3)
		for _, group := range groupedRows {
			if len(group.keys) < 2 {
				continue
			}
			metric := group.keys[0]
			status := summary[metric]
			if status == nil {
				status = make(map[string]*summaryItem)
				summary[metric] = status
			}
			statusName := group.keys[1]
			item := status[statusName]
			if item == nil {
				item = &summaryItem{}
				status[statusName] = item
			}
			for _, row := range group.rows {
				timestamp := int64(row.Values[1].GetNumberValue())
				avg := row.Values[4].GetNumberValue()
				max := row.Values[5].GetNumberValue()
				min := row.Values[6].GetNumberValue()
				count := int64(row.Values[7].GetNumberValue())
				sum := row.Values[8].GetNumberValue()
				item.time = append(item.time, timestamp/int64(time.Millisecond))
				item.avg = append(item.avg, avg)
				item.max = append(item.max, max)
				item.min = append(item.min, min)
				item.count = append(item.count, count)
				item.sum = append(item.sum, sum)
			}
		}
		for id, m := range metrics {
			idstr := strconv.FormatInt(id, 10)
			status, ok := summary[idstr]
			if !ok {
				continue
			}
			chart := &pb.CheckerChartV1{}
			m.Chart = chart
			for _, item := range status {
				if len(item.time) > len(chart.Time) {
					chart.Time = item.time
				}
				if len(chart.Latency) == 0 {
					chart.Latency = item.avg
				} else {
					minList, maxList := chart.Latency, item.avg
					if len(chart.Latency) > len(item.avg) {
						minList, maxList = item.avg, chart.Latency
					}
					for i, v := range minList {
						if v > maxList[i] {
							maxList[i] = v
						}
					}
					chart.Latency = maxList
				}
			}

			var downCount, upCount int64
			for stat, item := range status {
				if stat != statusRED && stat != statusGreen {
					stat = statusMiss
				}
				for i := range item.time {
					if i < len(chart.Status) {
						if item.count[i] > 0 {
							if stat == statusRED || chart.Status[i] == statusMiss {
								chart.Status[i] = stat
							}
						}
					} else {
						if item.count[i] > 0 {
							chart.Status = append(chart.Status, stat)
						} else {
							chart.Status = append(chart.Status, statusMiss)
						}
					}
					if stat == statusRED {
						downCount += item.count[i]
					} else {
						upCount += item.count[i]
					}
				}
				// TODO optimize
				if stat == statusRED {
					for i := len(item.time) - 1; i >= 0; i-- {
						if item.count[i] > 0 {
							j := i - 1
							for ; j >= 0; j-- {
								if item.count[j] <= 0 {
									break
								}
							}
							if j < 0 {
								j = 0
							}
							duration := (item.time[i]-item.time[j])/1000 + item.count[i]*30
							if duration < 60 {
								m.DownDuration = fmt.Sprintf("%d秒", duration)
							} else if duration < 60*60 {
								m.DownDuration = fmt.Sprintf("%d分钟", (duration+60-1)/60)
							} else {
								m.DownDuration = fmt.Sprintf("%d小时", (duration+60*60-1)/60*60)
							}
						}
					}
				}
			}
			if len(m.DownDuration) <= 0 {
				m.DownDuration = "0秒"
			}
			if len(chart.Status) > 0 {
				for i := len(chart.Status) - 1; i >= 0; i-- {
					if chart.Status[i] != statusMiss {
						m.Status = chart.Status[i]
						break
					}
				}
			}
			if len(m.Status) <= 0 {
				m.Status = statusMiss
			}
			totalCount := downCount + upCount
			if totalCount > 0 {
				m.Downtime = fmt.Sprintf("%.2f%%", float64(downCount)*100/float64(totalCount))
				m.Uptime = fmt.Sprintf("%.2f%%", float64(upCount)*100/float64(totalCount))
			} else {
				m.Downtime = "0%"
				m.Uptime = "0%"
			}

			var max, min, sum float64
			var count int64
			for _, item := range status {
				for _, v := range item.max {
					if v > max {
						max = v
					}
				}
				for i, v := range item.min {
					if v != 0 && item.count[i] > 0 {
						if min == 0 || v < min {
							min = v
						}
					}
				}
				for _, v := range item.sum {
					sum += v
				}
				for _, v := range item.count {
					count += v
				}
			}
			if count != 0 {
				m.Avg = roundFloat(sum / float64(count))
				m.Latency = m.Avg
			}
			m.Max = max
			m.Min = min
			m.Apdex = computeApdex(chart.Latency)
		}
	}
}

func computeApdex(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sc, tc float64
	for _, v := range values {
		v = v / 1000
		if v < 1 {
			sc++
		} else if v < 3 {
			tc++
		}
	}
	return (sc + tc/2) / float64(len(values))
}

func (s *checkerV1Service) GetCheckerStatusV1(ctx context.Context, req *pb.GetCheckerStatusV1Request) (*pb.GetCheckerStatusV1Response, error) {
	start, end, duration := getTimeRange("month", 3, true)
	interval, _ := structpb.NewValue(map[string]interface{}{"duration": duration})
	mreq := &metricpb.QueryWithInfluxFormatRequest{
		Start: strconv.FormatInt(start, 10),
		End:   strconv.FormatInt(end, 10),
		Statement: `
		SELECT timestamp(), status_name::tag, count(latency)
		FROM status_page 
		WHERE metric=$metric 
		GROUP BY time($interval), status_name::tag 
		LIMIT 200`,
		Params: map[string]*structpb.Value{
			"metric":   structpb.NewStringValue(strconv.FormatInt(req.Id, 10)),
			"interval": interval,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	resp, err := s.metricq.QueryWithInfluxFormat(ctx, mreq)
	if err != nil {
		return nil, errors.NewServiceInvokingError(fmt.Sprintf("status_page.metric/%d", req.Id), err)
	}
	var times []int64
	var status []string
	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {
		serie := resp.Results[0].Series[0]
		groupedRows := groupSerieRows(serie.Rows, 2)
		for _, group := range groupedRows {
			if len(group.keys) < 1 {
				continue
			}
			stat := group.keys[0]
			if stat != statusRED && stat != statusGreen {
				stat = statusMiss
			}
			var ts []int64
			for i, row := range group.rows {
				timestamp := int64(row.Values[1].GetNumberValue())
				count := int64(row.Values[3].GetNumberValue())

				ts = append(ts, timestamp/int64(time.Millisecond))
				if i < len(status) {
					if count > 0 {
						if stat == statusRED || status[i] == statusMiss {
							status[i] = stat
						}
					}
				} else {
					if count > 0 {
						status = append(status, stat)
					} else {
						status = append(status, statusMiss)
					}
				}
			}
			if len(ts) > len(times) {
				times = ts
			}
		}
	}
	return &pb.GetCheckerStatusV1Response{
		Data: &pb.CheckerStatusV1{
			Time:   times,
			Status: status,
		},
	}, nil
}

func (s *checkerV1Service) GetCheckerIssuesV1(ctx context.Context, req *pb.GetCheckerIssuesV1Request) (*pb.GetCheckerIssuesV1Response, error) {
	return &pb.GetCheckerIssuesV1Response{
		Data: make([]*structpb.Value, 0), // depracated, so return empty list
	}, nil
}

func roundFloat(v float64) float64 {
	v, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", v), 64)
	return v
}

type groupedRows struct {
	keys []string
	rows []*metricpb.Row
}

func groupSerieRows(rows []*metricpb.Row, keys ...int) []*groupedRows {
	var groups []*groupedRows
	lastTime := int64(math.MinInt64)
loop:
	for _, row := range rows {
		for _, val := range row.Values {
			if val == nil {
				continue loop
			}
		}
		timestamp := int64(row.Values[1].GetNumberValue())
		if len(groups) > 0 && timestamp > lastTime {
			groups[len(groups)-1].rows = append(groups[len(groups)-1].rows, row)
		} else {
			groups = append(groups, &groupedRows{rows: []*metricpb.Row{row}})
		}
		lastTime = timestamp
	}
	for _, group := range groups {
	rowsloop:
		for _, row := range group.rows {
			var kvals []string
			for _, i := range keys {
				val := row.Values[i].AsInterface()
				if val == nil {
					continue rowsloop
				}
				kvals = append(kvals, fmt.Sprint(val))
			}
			group.keys = kvals
			break
		}
	}
	return groups
}
