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

package monitor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/conf"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/pkg/cron"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/dlock"
	"github.com/erda-project/erda/pkg/ucauth"
)

const (
	CronBugStatisticsLock        = "/devops/cmdb/cron/bug-add-and-repair/lock"
	CronStatisticsLock           = "/devops/cmdb/cron/issue/lock"
	waitTimeIfLostDLock          = time.Minute
	CronBugStatistics            = "cron_bug_statistics"
	IssueMetricsName             = "issue_metrics_statistics"
	IssueAddAndRepairMetricsName = "issue_add_or_repair_metrics_statistics"
)

var minute = float64(apistructs.Day)

type IssueMonitor struct {
	Fields    map[string]interface{} `json:"fields"`
	Tags      map[string]string      `json:"tags"`
	timestamp time.Time
}

type StatisticsAddAndRepairBugRequest struct {
	IterationId     int
	ProjectId       int
	Timestamp       time.Time
	CreateStartTime time.Time
	CreateEndTime   time.Time
	db              *dao.DBClient
}

func (r *StatisticsAddAndRepairBugRequest) preFormat() {
	if !r.Timestamp.IsZero() {
		requestTime := r.Timestamp
		requestTimeTodayStartTime := time.Date(requestTime.Year(), requestTime.Month(), requestTime.Day(), 0, 0, 0, 0, requestTime.Location())
		requestTimeTodayEndTime := time.Date(requestTime.Year(), requestTime.Month(), requestTime.Day(), 23, 59, 59, 0, requestTime.Location())
		r.CreateStartTime = requestTimeTodayStartTime
		r.CreateEndTime = requestTimeTodayEndTime
	}
}

func (r *StatisticsAddAndRepairBugRequest) check() error {

	if r.ProjectId <= 0 {
		return errors.New(" error ProjectId is empty ")
	}
	if r.Timestamp.IsZero() {
		return errors.New(" error Timestamp is empty ")
	}

	if r.CreateStartTime.IsZero() {
		return errors.New(" error CreateStartTime is empty ")
	}
	if r.CreateEndTime.IsZero() {
		return errors.New(" error CreateEndTime is empty ")
	}

	return nil
}

func doMetrics(issueMonitor *IssueMonitor, bdl *bundle.Bundle) {
	metricsObject := apistructs.Metrics{}
	metricsObject.Metric = []apistructs.Metric{
		{
			Timestamp: time.Now().UnixNano(),
			Fields:    issueMonitor.Fields,
			Name:      IssueMetricsName,
			Tags:      issueMonitor.Tags,
		},
	}
	var count = 1
	for count < 3 {
		err := bdl.CollectMetrics(&metricsObject)
		if err != nil {
			logrus.Errorf(" doMetrics CollectMetrics error %v", err)
			count++
			time.Sleep(time.Minute)
			if count >= 3 {
				logrus.Errorf(" doMetrics CollectMetrics lost data %+v", metricsObject)
			}
			continue
		}
		break
	}
}

func batchMetrics(issueMonitor []IssueMonitor, bdl *bundle.Bundle) {
	metricsObject := apistructs.Metrics{}
	var list []apistructs.Metric
	for _, v := range issueMonitor {
		list = append(list, apistructs.Metric{
			Timestamp: v.timestamp.UnixNano(),
			Fields:    v.Fields,
			Name:      IssueAddAndRepairMetricsName,
			Tags:      v.Tags,
		})
	}
	metricsObject.Metric = list
	var count = 1
	for count < 3 {
		err := bdl.CollectMetrics(&metricsObject)
		if err != nil {
			logrus.Errorf(" batchMetrics CollectMetrics error %v", err)
			count++
			time.Sleep(time.Minute)
			if count >= 3 {
				logrus.Errorf(" batchMetrics CollectMetrics lost data %+v", metricsObject)
			}
			continue
		}
		break
	}
}

func RunIssueHistoryData(db *dao.DBClient, uc *ucauth.UCClient, bdl *bundle.Bundle) {
	logrus.Infof("start RunIssueHistoryData time %v", time.Now().Unix())

	allProject, err := bdl.GetAllProjects()
	if err != nil {
		logrus.Errorf("issue cron GetAllProjects error %v", err)
		return
	}

	for _, project := range allProject {
		iterations, err := db.FindIterations(project.ID)
		if err != nil {
			logrus.Errorf("issue cron FindIterations error %v, projectId %v ", err, project.ID)
			return
		}
		iterations = append(iterations, dao.Iteration{BaseModel: dbengine.BaseModel{CreatedAt: project.CreatedAt}})
		for i, iteration := range iterations {
			// the previous baseMode's ID is int64

			iterationID := int64(iteration.ID)
			if i == len(iterations)-1 {
				iterationID = -1
			}
			request := apistructs.IssueListRequest{ProjectID: project.ID, IterationID: int64(iterationID), OnlyIDResult: true}
			issueList, err := db.ListIssue(request)
			if err == nil && issueList != nil {
				for _, issue := range issueList {
					MetricsIssueById(int(issue.ID), db, uc, bdl)
				}
			}
		}

		logrus.Infof("RunIssueHistoryData issue end project %v", project.ID)
	}
	logrus.Infof("end RunIssueHistoryData time %v", time.Now().Unix())
}

func RunHistoryData(db *dao.DBClient, bdl *bundle.Bundle) {
	logrus.Infof("start RunHistoryData time %v", time.Now().Unix())

	allProject, err := bdl.GetAllProjects()
	if err != nil {
		logrus.Errorf("issue cron GetAllProjects error %v", err)
		return
	}
	for _, project := range allProject {
		projectObj := project
		func(project apistructs.ProjectDTO) {

			iterations, err := db.FindIterations(project.ID)
			if err != nil {
				logrus.Errorf("issue cron FindIterations error %v, projectId %v ", err, project.ID)
				return
			}
			iterations = append(iterations, dao.Iteration{BaseModel: dbengine.BaseModel{CreatedAt: project.CreatedAt}})
			for i, iteration := range iterations {
				var projectBatchList []IssueMonitor

				createTime := project.CreatedAt
				for {
					if createTime.Unix() > time.Now().Unix() {
						break
					}
					// the previous baseMode's ID is int64
					iterationID := int(iteration.ID)
					if i == len(iterations)-1 {
						iterationID = -1
					}

					statisticsBugSeverityRequest := StatisticsAddAndRepairBugRequest{}
					statisticsBugSeverityRequest.ProjectId = int(project.ID)
					statisticsBugSeverityRequest.IterationId = int(iterationID)
					statisticsBugSeverityRequest.Timestamp = createTime
					statisticsBugSeverityRequest.db = db

					result, err := statisticsAddAndRepairBug(statisticsBugSeverityRequest)
					if err != nil {
						logrus.Errorf("statisticsAddAndRepairBug error %v project_id %v iteration_id %v", err, project.ID, iteration.ID)
						continue
					}
					projectBatchList = append(projectBatchList, *result)
					createTime = createTime.Add(24 * time.Hour)
				}

				if projectBatchList != nil {
					batchMetrics(projectBatchList, bdl)
					logrus.Infof("end RunHistoryData project %v iteration %v time %v", project.ID, iteration.ID, time.Now().Unix())
				}
			}

		}(projectObj)
	}

	logrus.Infof("end RunHistoryData time %v", time.Now().Unix())
}

func MetricsIssueById(ID int, db *dao.DBClient, uc *ucauth.UCClient, bdl *bundle.Bundle) {
	if db == nil {
		logrus.Errorf("MetricsIssueById db is empty")
		return
	}
	if uc == nil {
		logrus.Errorf("MetricsIssueById uc is empty")
		return
	}

	if ID <= 0 {
		logrus.Errorf("MetricsIssueById id is empty")
		return
	}

	issue, err := db.GetIssueWithOutDelete(int64(ID))
	if err != nil {
		logrus.Warnf("not find issue in db id %v error %v", ID, err)
		return
	}

	if issue.Type != apistructs.IssueTypeRequirement && issue.Type != apistructs.IssueTypeTask && issue.Type != apistructs.IssueTypeBug {
		return
	}

	var isReOpen = false
	var reOpenTime = 0
	issueStream, err := db.FindIssueStream(ID)
	if err == nil {
		for k, v := range issueStream {
			if k == 0 && v.StreamParams.NewState == string(apistructs.IssueStateReopen) {
				isReOpen = true
			}

			if v.StreamParams.NewState == string(apistructs.IssueStateReopen) {
				reOpenTime = reOpenTime + 1
			}
		}
	}

	user, err := uc.GetUser(issue.Assignee)
	var nick string
	if err != nil {
		logrus.Warnf("get issue assignee user nick error %v", err)
	} else {
		nick = user.Nick
	}

	var manHour apistructs.IssueManHour
	if issue.ManHour != "" {
		err = json.Unmarshal([]byte(issue.ManHour), &manHour)
		if err != nil {
			logrus.Warnf("Unmarshal issue ManHour %v error %v", issue.ManHour, err)
		}
	}
	var issueMonitor *IssueMonitor
	if issue.Deleted {
		issueMonitor = NewIssueMonitor(
			WithDelete(int(issue.ID)),
		)
	} else {
		stateName, err := db.GetIssueStateByID(issue.State)
		if err != nil {
			logrus.Warnf("get issue state  error %v", err)
		}
		issueMonitor = NewIssueMonitor(
			WithIssueId(int(issue.ID)),
			WithIssueTitle(issue.Type.GetZhName()+"-"+strconv.Itoa(int(issue.ID))),
			WithAssignee(issue.Assignee),
			WithIssueIteratorId(int(issue.IterationID)),
			WithIssueProjectId(int(issue.ProjectID)),
			WithIssuePriority(issue.Priority),
			WithIssueSeverity(issue.Severity),
			WithIssueState(stateName.Belong),
			WithReOpen(isReOpen),
			WithReOpenTime(reOpenTime),
			WithManHour(manHour),
			WithIssueType(issue.Type),
			WithNick(nick),
			WithResponse(issue.CreatedAt, issue.UpdatedAt),
		)
	}

	if err != nil {
		logrus.Warnf("failed to send issue create monitor, (%v)", err)
		return
	}
	doMetrics(issueMonitor, bdl)
}

func TimedTaskMetricsIssue(db *dao.DBClient, uc *ucauth.UCClient, bdl *bundle.Bundle) {
	if db == nil {
		logrus.Errorf("db is empty")
		return
	}
	crond := cron.New()
	crond.Start()

	ctx, cancel := context.WithCancel(context.Background())

	lock, err := dlock.New(
		CronStatisticsLock,
		func() {
			logrus.Errorf("[alert] dlock lost, stop current issue cron")
			cancel()
			time.Sleep(waitTimeIfLostDLock)
			logrus.Warn("try to continue issue cron again")
			go TimedTaskMetricsIssue(db, uc, bdl)
		},
		dlock.WithTTL(30),
	)
	if err != nil {
		logrus.Errorf("[alert] failed to get dlock, err: %v", err)
		time.Sleep(waitTimeIfLostDLock)
		go TimedTaskMetricsIssue(db, uc, bdl)
		return
	}
	if err := lock.Lock(context.Background()); err != nil {
		logrus.Errorf("[alert] failed to lock dlock, err: %v", err)
		time.Sleep(waitTimeIfLostDLock)
		go TimedTaskMetricsIssue(db, uc, bdl)
		return
	}

	defer func() {
		if lock != nil {
			_ = lock.UnlockAndClose()
		}
	}()

	logrus.Info("issue cron: start")

	if err = crond.AddFunc(conf.MetricsIssueCron(), func() {
		logrus.Info("cron run start TimedTaskMetricsIssue")
		go RunIssueHistoryData(db, uc, bdl)
	}, CronBugStatistics); err != nil {
		logrus.Errorf("failed to load issue cron item: %s, err: %v", CronBugStatistics, err)
		return
	}

	select {
	case <-ctx.Done():
		_ = crond.Remove(CronBugStatistics)
		crond.Stop()
		logrus.Info("stop issue cron, received cancel signal from channel")
		return
	}
}

func TimedTaskMetricsAddAndRepairBug(db *dao.DBClient, bdl *bundle.Bundle) {
	if db == nil {
		logrus.Errorf("db is empty")
		return
	}

	crond := cron.New()
	crond.Start()

	ctx, cancel := context.WithCancel(context.Background())

	lock, err := dlock.New(
		CronBugStatisticsLock,
		func() {
			logrus.Errorf("[alert] dlock lost, stop current issue cron")
			cancel()
			time.Sleep(waitTimeIfLostDLock)
			logrus.Warn("try to continue issue cron again")
			go TimedTaskMetricsAddAndRepairBug(db, bdl)
		},
		dlock.WithTTL(30),
	)
	if err != nil {
		logrus.Errorf("[alert] failed to get dlock, err: %v", err)
		time.Sleep(waitTimeIfLostDLock)
		go TimedTaskMetricsAddAndRepairBug(db, bdl)
		return
	}
	if err := lock.Lock(context.Background()); err != nil {
		logrus.Errorf("[alert] failed to lock dlock, err: %v", err)
		time.Sleep(waitTimeIfLostDLock)
		go TimedTaskMetricsAddAndRepairBug(db, bdl)
		return
	}

	defer func() {
		if lock != nil {
			_ = lock.UnlockAndClose()
		}
	}()

	logrus.Info("issue cron: start")

	//每天凌晨一点执行一次
	if err = crond.AddFunc("0 0 1 * * ?", func() {
		logrus.Info("cron run start DoCronStatisticsAddAndRepairBug")
		DoCronStatisticsAddAndRepairBug(db, bdl)
	}, CronBugStatistics); err != nil {
		logrus.Errorf("failed to load issue cron item: %s, err: %v", CronBugStatistics, err)
		return
	}

	select {
	case <-ctx.Done():
		_ = crond.Remove(CronBugStatistics)
		crond.Stop()
		logrus.Info("stop issue cron, received cancel signal from channel")
		return
	}
}

func DoCronStatisticsAddAndRepairBug(db *dao.DBClient, bdl *bundle.Bundle) {
	allProject, err := bdl.GetAllProjects()
	if err != nil {
		logrus.Errorf("issue cron GetAllProjects error %v", err)
		return
	}

	nowTime := time.Now()
	twoHourBeforeTime := nowTime.Add(-2 * time.Hour)

	for _, project := range allProject {

		var batchList []IssueMonitor

		iterations, err := db.FindIterations(uint64(project.ID))
		if err != nil {
			logrus.Errorf("issue cron FindIterations error %v, projectId %v ", err, project.ID)
			continue
		}
		iterations = append(iterations, dao.Iteration{BaseModel: dbengine.BaseModel{CreatedAt: project.CreatedAt}})
		for i, iteration := range iterations {
			// // the previous baseMode's ID is int64
			iterationID := int(iteration.ID)
			if i == len(iterations) {
				iterationID = -1
			}
			statisticsBugSeverityRequest := StatisticsAddAndRepairBugRequest{}
			statisticsBugSeverityRequest.ProjectId = int(project.ID)
			statisticsBugSeverityRequest.IterationId = int(iterationID)
			statisticsBugSeverityRequest.Timestamp = twoHourBeforeTime
			statisticsBugSeverityRequest.db = db

			result, err := statisticsAddAndRepairBug(statisticsBugSeverityRequest)
			if err != nil {
				logrus.Errorf("statisticsAddAndRepairBug error %v project_id %v iteration_id %v", err, project.ID, iteration.ID)
				continue
			}
			batchList = append(batchList, *result)
		}

		if batchList != nil {
			batchMetrics(batchList, bdl)
		}
	}
}

func statisticsAddAndRepairBug(r StatisticsAddAndRepairBugRequest) (*IssueMonitor, error) {

	r.preFormat()
	if err := r.check(); err != nil {
		return nil, err
	}

	var sql string
	sql = " select count(*) as counts from dice_issues WHERE 1=1 "
	if r.IterationId != 0 {
		sql += " and iteration_id = ? "
	}
	if r.ProjectId > 0 {
		sql += " and project_id = ? "
	}
	sql += " and type = ? and deleted = 0 "

	addSql := sql
	if !r.CreateStartTime.IsZero() {
		addSql += " and created_at > ? "
	}
	if !r.CreateEndTime.IsZero() {
		addSql += "and created_at < ? "
	}

	var params []interface{}
	if r.IterationId != 0 {
		params = append(params, r.IterationId)
	}
	if r.ProjectId > 0 {
		params = append(params, r.ProjectId)
	}
	params = append(params, apistructs.IssueTypeBug)

	if !r.CreateStartTime.IsZero() {
		params = append(params, r.CreateStartTime)
	}
	if !r.CreateEndTime.IsZero() {
		params = append(params, r.CreateEndTime)
	}

	var repairCounts int
	var addCounts int

	//一个时间区间的总数
	row := r.db.Raw(addSql, params...).Row()
	if row != nil {
		var counts int
		_ = row.Scan(&counts)
		addCounts = counts
	}

	//一个update时间区间的state为CLOSED的
	repairSql := sql
	if !r.CreateStartTime.IsZero() {
		repairSql += " and updated_at > ? "
	}
	if !r.CreateEndTime.IsZero() {
		repairSql += "and updated_at < ? "
	}
	repairSql += fmt.Sprintf(" and state = (select id from dice_issue_state where project_id = %v and issue_type = '%s' and belong = '%s')", r.ProjectId, apistructs.IssueTypeBug, apistructs.IssueStateClosed)
	row = r.db.Raw(repairSql, params...).Row()
	if row != nil {
		var counts int
		_ = row.Scan(&counts)
		repairCounts = counts
	}

	issueMonitorData := NewIssueMonitor()
	issueMonitorData.Tags["project_id"] = strconv.Itoa(r.ProjectId)
	issueMonitorData.Tags["_id"] = strconv.Itoa(r.IterationId) + "_" + strconv.Itoa(r.ProjectId) +
		"_" + strconv.Itoa(r.CreateStartTime.Year()) + "-" + strconv.Itoa(int(r.CreateStartTime.Month())) + "-" + strconv.Itoa(r.CreateStartTime.Day())
	issueMonitorData.Tags["issue_iterator_id"] = strconv.Itoa(r.IterationId)
	issueMonitorData.Fields["bug_repair_counts"] = repairCounts
	issueMonitorData.Fields["bug_add_counts"] = addCounts
	//issueMonitorData.Tags["bug_add_or_repair_type"] = "repair"
	issueMonitorData.timestamp = r.Timestamp
	return issueMonitorData, nil
}

type Option func(monitor *IssueMonitor)

func NewIssueMonitor(opts ...Option) *IssueMonitor {
	issueMonitor := IssueMonitor{}
	issueMonitor.Tags = map[string]string{}
	issueMonitor.Fields = map[string]interface{}{}
	for _, v := range opts {
		v(&issueMonitor)
	}

	if issueMonitor.Fields["re_open_time"] != nil && issueMonitor.Fields["re_open_time"].(int) > 0 {
		for _, v := range apistructs.IssueSeveritys {
			if string(v) == issueMonitor.Tags["issue_severity"] {
				issueMonitor.Fields["is_re_open_"+string(v)] = 1
			}
			//} else {
			//	issueMonitor.Fields["is_re_open_"+string(v)] = 0
			//}
		}
	}

	return &issueMonitor
}

func WithDelete(id int) Option {
	return func(monitor *IssueMonitor) {
		monitor.Fields = map[string]interface{}{}
		monitor.Tags = map[string]string{}
		monitor.Tags["_id"] = strconv.Itoa(id)
		monitor.Tags["not_delete"] = "0"
	}
}

func WithAssignee(name string) Option {
	return func(monitor *IssueMonitor) {
		monitor.Tags["issue_user"] = name
	}
}

func WithIssueId(id int) Option {
	return func(monitor *IssueMonitor) {
		monitor.Tags["_id"] = strconv.Itoa(id)
		monitor.Tags["not_delete"] = "1"
		monitor.Fields["counts"] = 1
	}
}

func WithIssueTitle(title string) Option {
	return func(monitor *IssueMonitor) {
		monitor.Tags["issue_title"] = title
	}
}

func WithIssuePriority(priority apistructs.IssuePriority) Option {
	return func(monitor *IssueMonitor) {
		monitor.Tags["issue_priority"] = string(priority)
	}
}

func WithIssueState(state apistructs.IssueStateBelong) Option {
	return func(monitor *IssueMonitor) {
		monitor.Tags["issue_state"] = string(state)
		if monitor.Tags["issue_state"] != string(apistructs.IssueStateBelongClosed) && monitor.Tags["issue_state"] != string(apistructs.IssueStateBelongDone) {
			monitor.Fields["not_close"] = 1
		}
		//} else {
		//	monitor.Fields["not_close"] = 0
		//}
	}
}

func WithIssueProjectId(projectId int) Option {
	return func(monitor *IssueMonitor) {
		monitor.Tags["project_id"] = strconv.Itoa(projectId)
	}
}

func WithIssueIteratorId(iteratorId int) Option {
	return func(monitor *IssueMonitor) {
		monitor.Tags["issue_iterator_id"] = strconv.Itoa(iteratorId)
	}
}

func WithIssueSeverity(severity apistructs.IssueSeverity) Option {
	return func(monitor *IssueMonitor) {
		monitor.Tags["issue_severity"] = string(severity)
	}
}

func WithReOpen(reOpen bool) Option {
	return func(monitor *IssueMonitor) {
		if reOpen {
			monitor.Fields["is_re_open"] = 1
		}
		//} else {
		//	monitor.Fields["is_re_open"] = 0
		//}
	}
}

func WithReOpenTime(time int) Option {
	return func(monitor *IssueMonitor) {
		if time > 0 {
			monitor.Fields["re_open_time"] = time
		}
	}
}

func WithIssueType(typer apistructs.IssueType) Option {
	return func(monitor *IssueMonitor) {
		monitor.Tags["issue_type"] = string(typer)
		for _, v := range apistructs.IssueTypes {
			if string(v) == string(typer) {
				monitor.Fields["issue_type_"+string(v)] = 1
			}
			//} else {
			//	monitor.Fields["issue_type_"+string(v)] = 0
			//}
		}
	}
}

func WithNick(nick string) Option {
	return func(monitor *IssueMonitor) {
		monitor.Tags["issue_user_nick"] = nick
	}
}

func WithManHour(manHour apistructs.IssueManHour) Option {
	return func(monitor *IssueMonitor) {
		estimateTime, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", float64(manHour.EstimateTime)/minute), 64)
		thisElapsedTime, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", float64(manHour.ThisElapsedTime)/minute), 64)
		elapsedTime, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", float64(manHour.ElapsedTime)/minute), 64)
		remainingTime, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", float64(manHour.RemainingTime)/minute), 64)
		if estimateTime > 0 {
			monitor.Fields["issue_estimate_time"] = estimateTime
		}
		if thisElapsedTime > 0 {
			monitor.Fields["issue_this_elapsed_time"] = thisElapsedTime
		}
		if elapsedTime > 0 {
			monitor.Fields["issue_elapsed_time"] = elapsedTime
		}
		if remainingTime > 0 {
			monitor.Fields["issue_remaining_time"] = remainingTime
		}
	}
}

func WithResponse(createTime time.Time, endUpdateTime time.Time) Option {
	return func(monitor *IssueMonitor) {
		createSecond := createTime.Unix()
		updateSecond := endUpdateTime.Unix()
		defSecond := updateSecond - createSecond
		day, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", float64(defSecond)/float64(60*24*apistructs.Hour)), 64)
		if day > 0 {
			monitor.Fields["issue_response_time"] = day
		}
	}
}
