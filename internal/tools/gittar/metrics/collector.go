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

package metrics

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/gittar/conf"
	"github.com/erda-project/erda/internal/tools/gittar/models"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	defaultBranch = "master"
	targetBranch  = []string{"master", "Master", "main", "Main", "dev", "Dev", "develop", "Develop", "release", "Release"}
)

var (
	personalCommitsTotalMetricName      = "personal_commits_total"
	personalCommitsTotalMetricHelp      = "personal commits total for repo"
	personalFilesChangedTotalMetricName = "personal_files_changed_total"
	personalFilesChangedTotalMetricHelp = "personal files changed total for repo"
	personalAdditionTotalMetricName     = "personal_addition_total"
	personalAdditionTotalMetricHelp     = "personal addition total for repo"
	PersonalDeletionTotalMetricName     = "personal_deletion_total"
	PersonalDeletionTotalMetricHelp     = "personal deletion total for repo"

	personalDailyCommitsMetricName      = "personal_daily_commits_total"
	personalDailyCommitsMetricHelp      = "personal daily commits total for repo"
	personalDailyFilesChangedMetricName = "personal_daily_files_changed_total"
	personalDailyFilesChangedMetricHelp = "personal daily files changed total for repo"
	personalDailyAdditionMetricName     = "personal_daily_addition_total"
	personalDailyAdditionMetricHelp     = "personal daily addition total for repo"
	personalDailyDeletionMetricName     = "personal_daily_deletion_total"
	personalDailyDeletionMetricHelp     = "personal daily deletion total for repo"
)

type PersonalMetric struct {
	UserName    string
	UserEmail   string
	RepoID      uint64
	Repo        string
	ProjectID   uint64
	AppID       uint64
	ProjectName string
	AppName     string
	OrgID       uint64
	OrgName     string
	Field       *MetricField
}

type MetricField struct {
	CalculatedAt time.Time
	// Historical cumulative total
	CommitTotal     uint64
	FileChangeTotal uint64
	AdditionTotal   uint64
	DeletionTotal   uint64

	// The cumulative total within a certain duration
	DurationCommitTotal     uint64
	DurationFileChangeTotal uint64
	DurationAdditionTotal   uint64
	DurationDeletionTotal   uint64
}

func NewPersonalMetric(author *gitmodule.Signature, repo *models.Repo) *PersonalMetric {
	return &PersonalMetric{
		UserName:    author.Name,
		UserEmail:   author.Email,
		RepoID:      uint64(repo.ID),
		Repo:        conf.GittarUrl() + "/" + repo.Path,
		ProjectID:   uint64(repo.ProjectID),
		AppID:       uint64(repo.AppID),
		AppName:     repo.AppName,
		ProjectName: repo.ProjectName,
		OrgID:       uint64(repo.OrgID),
		OrgName:     repo.OrgName,
		Field:       &MetricField{},
	}
}

func IsValidBranch(s string, prefixes ...string) bool {
	if !strutil.HasPrefixes(s, prefixes...) {
		return false
	}
	return true
}

type Collector struct {
	sync.RWMutex
	errors                prometheus.Gauge
	personalContributions []*PersonalMetric
	svc                   *models.Service
}

func NewCollector(svc *models.Service) *Collector {
	return &Collector{
		personalContributions: make([]*PersonalMetric, 0),
		errors: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "gittar",
			Name:      "scrape_error",
			Help:      "1 if there was an error while getting personal contribution metrics, 0 otherwise",
		}),
		svc: svc,
	}
}

func (c *Collector) RefreshPersonalContributions() error {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local)
	contributions, err := c.IterateRepos(apistructs.GittarListRepoRequest{
		Start: &start,
		End:   &end,
	})
	if err != nil {
		return err
	}
	logrus.Infof("contributions: %v", contributions)
	c.Lock()
	c.personalContributions = contributions
	c.Unlock()
	return nil
}

func (c *Collector) IterateRepos(req apistructs.GittarListRepoRequest) ([]*PersonalMetric, error) {
	repos, err := c.svc.ListAllRepos(req)
	if err != nil {
		return nil, err
	}
	repoMetrics := make(map[string]*PersonalMetric)
	var start, end string
	if req.Start != nil {
		start = req.Start.Format("2006-01-02 15:04:05")
	}
	if req.End != nil {
		end = req.End.Format("2006-01-02 15:04:05")
	}
	for _, repo := range repos {
		gitRepository, err := gitmodule.OpenRepositoryWithInit(conf.RepoRoot(), repo.Path)
		if err != nil {
			logrus.Errorf("failed to open repo %s, err: %v", repo.Path, err)
			continue
		}
		gitRepository.ID = repo.ID
		gitRepository.ProjectId = repo.ProjectID
		gitRepository.ProjectName = repo.ProjectName
		gitRepository.ApplicationId = repo.AppID
		gitRepository.ApplicationName = repo.AppName
		gitRepository.OrgId = repo.OrgID
		gitRepository.OrgName = repo.OrgName
		gitRepository.Size = repo.Size
		gitRepository.Url = conf.GittarUrl() + "/" + repo.Path
		gitRepository.IsExternal = repo.IsExternal

		branches, err := gitRepository.GetBranches()
		if err != nil {
			logrus.Warningf("failed to get all branch for repo: %s, err: %v, use branch: %s as default", repo.Path, err, defaultBranch)
			branches = []string{defaultBranch}
		}
		var allSourceCommit []*gitmodule.Commit
		for _, branch := range branches {
			if !IsValidBranch(branch, targetBranch...) {
				continue
			}
			SourceCommit, err := gitRepository.GetBranchCommit(branch)
			if err != nil {
				logrus.Errorf("failed to get curCommit for branch: %s", branch)
				continue
			}
			allSourceCommit = append(allSourceCommit, SourceCommit)
		}
		commitsSet := make(map[string]struct{})
		var commits []*gitmodule.Commit
		for _, sourceCommit := range allSourceCommit {
			curCommits, err := gitRepository.CommitsBetweenDuration(sourceCommit.ID, start, end)
			if err != nil {
				logrus.Errorf("failed to list commits for repo: %s, commit-id: %s", repo.Path, sourceCommit.ID)
				continue
			}
			for _, curCommit := range curCommits {
				if _, ok := commitsSet[curCommit.ID]; ok {
					continue
				} else {
					commits = append(commits, curCommit)
					commitsSet[curCommit.ID] = struct{}{}
				}
			}
		}
		for _, commit := range commits {
			author := commit.Author
			uniqueKey := makePersonalUniqueKey(author, repo)
			commitInDuration := inDuration(author.When, req.Start, req.End)
			if repoMetrics[uniqueKey] == nil {
				personalMetric := NewPersonalMetric(author, repo)
				if req.Start != nil {
					personalMetric.Field.CalculatedAt = *req.Start
				} else {
					personalMetric.Field.CalculatedAt = time.Now()
				}
				personalMetric.Field.CommitTotal += 1
				repoMetrics[uniqueKey] = personalMetric
			} else {
				repoMetrics[uniqueKey].Field.CommitTotal += 1
			}
			if commitInDuration {
				repoMetrics[uniqueKey].Field.DurationCommitTotal += 1
			}
			if len(commit.Parents) == 0 {
				continue
			}
			oldCommit, _ := gitRepository.GetCommit(commit.Parents[0])
			if oldCommit != nil {
				diff, err := gitRepository.GetDiff(commit, oldCommit)
				if err != nil {
					logrus.Errorf("failed to get diff for repo: %s, commit: %s, err: %v", repo.Path, commit.ID, err)
				} else {
					repoMetrics[uniqueKey].Field.FileChangeTotal += uint64(diff.FilesChanged)
					repoMetrics[uniqueKey].Field.AdditionTotal += uint64(diff.TotalAddition)
					repoMetrics[uniqueKey].Field.DeletionTotal += uint64(diff.TotalDeletion)
					if commitInDuration {
						repoMetrics[uniqueKey].Field.DurationAdditionTotal += uint64(diff.TotalAddition)
						repoMetrics[uniqueKey].Field.DurationDeletionTotal += uint64(diff.TotalDeletion)
						repoMetrics[uniqueKey].Field.DurationFileChangeTotal += uint64(diff.FilesChanged)
					}
				}
			}
		}
	}
	res := make([]*PersonalMetric, 0, len(repoMetrics))
	for _, metric := range repoMetrics {
		res = append(res, metric)
	}
	return res, nil
}

func inDuration(when time.Time, start, end *time.Time) bool {
	if start == nil || end == nil {
		return false
	}
	return when.After(*start) && when.Before(*end)
}

func makePersonalUniqueKey(auth *gitmodule.Signature, repo *models.Repo) string {
	return fmt.Sprintf("org_%d_project_%d_app_%d_repo_%d_email_%s", repo.OrgID, repo.ProjectID, repo.AppID, repo.ID, auth.Email)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.errors.Set(0)
	c.Lock()
	defer c.Unlock()
	for _, field := range c.personalContributions {
		if field.Field == nil || field.Field.CalculatedAt.Day() != time.Now().Day() {
			continue
		}
		rawLabels := map[string]struct{}{}
		for l := range c.personalLabelsFunc(field) {
			rawLabels[l] = struct{}{}
		}
		values := make([]string, 0, len(rawLabels))
		labels := make([]string, 0, len(rawLabels))
		personalLabels := c.personalLabelsFunc(field)
		for l := range rawLabels {
			duplicate := false
			for _, x := range labels {
				if l == x {
					duplicate = true
					break
				}
			}
			if !duplicate {
				labels = append(labels, l)
				values = append(values, personalLabels[l])
			}
		}
		ch <- prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(prometheus.NewDesc(personalCommitsTotalMetricName, personalCommitsTotalMetricHelp, labels, nil), prometheus.CounterValue, float64(field.Field.CommitTotal), values...),
		)
		ch <- prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(prometheus.NewDesc(personalFilesChangedTotalMetricName, personalFilesChangedTotalMetricHelp, labels, nil), prometheus.CounterValue, float64(field.Field.FileChangeTotal), values...),
		)
		ch <- prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(prometheus.NewDesc(personalAdditionTotalMetricName, personalAdditionTotalMetricHelp, labels, nil), prometheus.CounterValue, float64(field.Field.AdditionTotal), values...),
		)
		ch <- prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(prometheus.NewDesc(PersonalDeletionTotalMetricName, PersonalDeletionTotalMetricHelp, labels, nil), prometheus.CounterValue, float64(field.Field.DeletionTotal), values...),
		)
		ch <- prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(prometheus.NewDesc(personalDailyCommitsMetricName, personalDailyCommitsMetricHelp, labels, nil), prometheus.CounterValue, float64(field.Field.DurationCommitTotal), values...),
		)
		ch <- prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(prometheus.NewDesc(personalDailyFilesChangedMetricName, personalDailyFilesChangedMetricHelp, labels, nil), prometheus.CounterValue, float64(field.Field.DurationFileChangeTotal), values...),
		)
		ch <- prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(prometheus.NewDesc(personalDailyAdditionMetricName, personalDailyAdditionMetricHelp, labels, nil), prometheus.CounterValue, float64(field.Field.DurationAdditionTotal), values...),
		)
		ch <- prometheus.NewMetricWithTimestamp(
			time.Now(),
			prometheus.MustNewConstMetric(prometheus.NewDesc(personalDailyDeletionMetricName, personalDailyDeletionMetricHelp, labels, nil), prometheus.CounterValue, float64(field.Field.DurationDeletionTotal), values...),
		)
	}
	c.errors.Collect(ch)
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.errors.Describe(ch)
	ch <- prometheus.NewDesc("personal_commits_total", "commits_ total", []string{"email"}, nil)
}

func (c *Collector) personalLabelsFunc(personalContributor *PersonalMetric) map[string]string {
	labels := map[string]string{
		"user_name":    personalContributor.UserName,
		"user_email":   personalContributor.UserEmail,
		"repo_id":      strconv.FormatUint(personalContributor.RepoID, 10),
		"repo":         personalContributor.Repo,
		"project_id":   strconv.FormatUint(personalContributor.ProjectID, 10),
		"app_id":       strconv.FormatUint(personalContributor.AppID, 10),
		"project_name": personalContributor.ProjectName,
		"app_name":     personalContributor.AppName,
		"org_id":       strconv.FormatUint(personalContributor.OrgID, 10),
		"org_name":     personalContributor.OrgName,
	}
	return labels
}
