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
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda-proto-go/msp/apm/checker/pb"
	"github.com/erda-project/erda/modules/msp/apm/checker/storage/cache"
	"github.com/erda-project/erda/modules/msp/apm/checker/storage/db"
	"github.com/erda-project/erda/pkg/common/errors"
)

type checkerV1Service struct {
	projectDB *db.ProjectDB
	metricDB  *db.MetricDB
	cache     *cache.Cache
	metricq   metricpb.MetricServiceServer
}

func (s *checkerV1Service) CreateCheckerV1(ctx context.Context, req *pb.CreateCheckerV1Request) (*pb.CreateCheckerV1Response, error) {
	if req.Data == nil {
		return nil, errors.NewMissingParameterError("data")
	}
	proj, err := s.projectDB.GetByProjectID(req.ProjectID)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	if proj == nil {
		return nil, errors.NewNotFoundError(fmt.Sprintf("project/%d", req.ProjectID))
	}
	now := time.Now()
	m := &db.Metric{
		ProjectID:  proj.ID,
		Env:        req.Data.Env,
		Name:       req.Data.Name,
		Mode:       req.Data.Mode,
		URL:        req.Data.Url,
		CreateTime: now,
		UpdateTime: now,
	}
	if err := s.metricDB.Create(m); err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	checker := s.metricDB.ConvertToChecker(m, req.ProjectID)
	if checker != nil {
		s.cache.Put(checker)
	}
	return &pb.CreateCheckerV1Response{Data: m.ID}, nil
}

func (s *checkerV1Service) UpdateCheckerV1(ctx context.Context, req *pb.UpdateCheckerV1Request) (*pb.UpdateCheckerV1Response, error) {
	if req.Data == nil {
		return nil, errors.NewMissingParameterError("data")
	}
	metric, err := s.metricDB.GetByID(req.Id)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	if metric == nil {
		return nil, errors.NewNotFoundError(fmt.Sprintf("metric/%d", req.Id))
	}
	metric.URL = req.Data.Url
	metric.Name = req.Data.Name
	if err := s.metricDB.Update(metric); err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	checker := s.metricDB.ConvertToChecker(metric, -1)
	if checker != nil {
		s.cache.Put(checker)
	}
	return &pb.UpdateCheckerV1Response{Data: req.Id}, nil
}

func (s *checkerV1Service) DeleteCheckerV1(ctx context.Context, req *pb.DeleteCheckerV1Request) (*pb.DeleteCheckerV1Response, error) {
	metric, err := s.metricDB.GetByID(req.Id)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	err = s.metricDB.Delete(req.Id)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}

	s.cache.Remove(req.Id)

	c := &pb.CheckerV1{}
	if metric != nil {
		c.Name = metric.Name
		c.Mode = metric.Mode
		c.Url = metric.URL
	}
	return &pb.DeleteCheckerV1Response{Data: c}, nil
}

func (s *checkerV1Service) DescribeCheckersV1(ctx context.Context, req *pb.DescribeCheckersV1Request) (*pb.DescribeCheckersV1Response, error) {
	proj, err := s.projectDB.GetByProjectID(req.ProjectID)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	if proj == nil {
		return nil, errors.NewNotFoundError(fmt.Sprintf("project/%d", req.ProjectID))
	}
	list, err := s.metricDB.ListByProjectID(proj.ID)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	results := make(map[int64]*pb.DescribeItemV1)
	for _, item := range list {
		result := &pb.DescribeItemV1{
			Name: item.Name,
			Mode: item.Mode,
			Url:  item.URL,
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
		return nil, errors.NewDataBaseError(err)
	}
	results := make(map[int64]*pb.DescribeItemV1)
	var downCount int64
	if metric != nil {
		results[req.Id] = &pb.DescribeItemV1{
			Name: metric.Name,
			Mode: metric.Mode,
			Url:  metric.URL,
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

func (s *checkerV1Service) getTimeRange(unit string, num int) (start int64, end int64) {
	now := time.Now()
	end = now.UnixNano() / int64(time.Millisecond)
	switch strings.ToLower(unit) {
	case "year":
		start = now.AddDate(-1*num, 0, 0).UnixNano() / int64(time.Millisecond)
	case "month":
		start = now.AddDate(0, -1*num, 0).UnixNano() / int64(time.Millisecond)
	case "week":
		start = now.AddDate(0, 0, -7*num).UnixNano() / int64(time.Millisecond)
	case "day":
		start = now.AddDate(0, 0, -1*num).UnixNano() / int64(time.Millisecond)
	default:
		start = now.Add(time.Duration(-1*int64(num)*int64(time.Hour))).UnixNano() / int64(time.Millisecond)
	}
	return
}

func (s *checkerV1Service) queryCheckersLatencySummaryByProject(projectID int64, metrics map[int64]*pb.DescribeItemV1) error {
	return s.queryCheckerMetrics(`
	SELECT timestamp, metric::tag, status_name::tag, avg(latency), max(latency), min(latency), count(latency), sum(latency)
	FROM status_page 
	WHERE project_id=$projectID 
	GROUP time(1m), metric::tag, status_name::tag 
	LIMIT 200`,
		map[string]*structpb.Value{
			"projectID": structpb.NewStringValue(strconv.FormatInt(projectID, 10)),
		},
		"hour", metrics,
	)
}

func (s *checkerV1Service) queryCheckersLatencySummary(metricID int64, timeUnit string, metrics map[int64]*pb.DescribeItemV1) error {
	return s.queryCheckerMetrics(`
	SELECT timestamp, metric::tag, status_name::tag, avg(latency), max(latency), min(latency), count(latency), sum(latency)
	FROM status_page 
	WHERE metric=$metric 
	GROUP time(1m), metric::tag, status_name::tag 
	LIMIT 200`,
		map[string]*structpb.Value{
			"metric": structpb.NewStringValue(strconv.FormatInt(metricID, 10)),
		},
		timeUnit, metrics,
	)
}

func (s *checkerV1Service) queryCheckerMetrics(statement string, params map[string]*structpb.Value, timeUnit string, metrics map[int64]*pb.DescribeItemV1) error {
	start, end := s.getTimeRange(timeUnit, 1)
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
	loop:
		for _, row := range serie.Rows {
			if row == nil || len(row.Values) != 6 {
				continue
			}
			for _, val := range row.Values {
				if val == nil {
					continue loop
				}
			}
			idx := 1
			timestamp := int64(row.Values[idx].GetNumberValue())
			idx++
			metric := row.Values[idx].GetStringValue()
			idx++
			statusName := row.Values[idx].GetStringValue()
			idx++
			avg := row.Values[idx].GetNumberValue()
			idx++
			max := row.Values[idx].GetNumberValue()
			idx++
			min := row.Values[idx].GetNumberValue()
			idx++
			sum := row.Values[idx].GetNumberValue()
			idx++
			count := int64(row.Values[idx].GetNumberValue())

			status := summary[metric]
			if status == nil {
				status = make(map[string]*summaryItem)
				summary[metric] = status
			}
			item := status[statusName]
			if item == nil {
				item = &summaryItem{}
				status[statusName] = item
			}
			item.time = append(item.time, timestamp)
			item.avg = append(item.avg, avg)
			item.max = append(item.max, max)
			item.min = append(item.min, min)
			item.sum = append(item.min, sum)
			item.count = append(item.count, count)
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
				if stat != statusRED {
					stat = statusGreen
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
			} else {
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
				m.Avg = sum / float64(count)
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
	start, end := s.getTimeRange("mount", 3)
	mreq := &metricpb.QueryWithInfluxFormatRequest{
		Start: strconv.FormatInt(start, 10),
		End:   strconv.FormatInt(end, 10),
		Statement: `
		SELECT timestamp, status_name::tag, count(latency)
		FROM status_page 
		WHERE metric=$metric 
		GROUP time(1m), status_name::tag 
		LIMIT 200`,
		Params: map[string]*structpb.Value{
			"metric": structpb.NewStringValue(strconv.FormatInt(req.Id, 10)),
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
		type groupItem struct {
			times []int64
			count []int64
		}
		group := make(map[string]*groupItem)
	loop:
		for _, row := range serie.Rows {
			for _, val := range row.Values {
				if val == nil {
					continue loop
				}
			}
			idx := 1
			timestamp := int64(row.Values[idx].GetNumberValue())
			idx++
			statusName := row.Values[idx].GetStringValue()
			idx++
			count := int64(row.Values[idx].GetNumberValue())
			item := group[statusName]
			if item == nil {
				item = &groupItem{}
				group[statusName] = item
			}
			item.times = append(item.times, timestamp)
			item.count = append(item.count, count)
		}
		for stat, item := range group {
			if len(item.times) > len(times) {
				times = item.times
			}
			if stat != statusRED {
				stat = statusGreen
			}
			for i := range item.times {
				if i < len(status) {
					if item.count[i] > 0 {
						if stat == statusRED || status[i] == statusMiss {
							status[i] = stat
						}
					}
				} else {
					if item.count[i] > 0 {
						status = append(status, stat)
					} else {
						status = append(status, statusMiss)
					}
				}
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
