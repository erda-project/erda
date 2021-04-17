package slsimport

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/mitchellh/mapstructure"
)

// Filters .
type Filters []*regexp.Regexp

func buildFilters(filters []string) (Filters, error) {
	var fs Filters
	for _, f := range filters {
		r, err := regexp.Compile(f)
		if err != nil {
			return nil, fmt.Errorf("invalid filter %s : %s", f, err)
		}
		fs = append(fs, r)
	}
	return fs, nil
}

// Match .
func (fs Filters) Match(text string) bool {
	for _, r := range fs {
		if r.MatchString(text) {
			return true
		}
	}
	return false
}

// sls log filters
var logFilterMap = map[string][]LogFilter{
	"rds": {&RDSLogFilter{}},
}

func initLogFilter(product string, options map[string]interface{}) {
	if fls, ok := logFilterMap[product]; ok {
		for _, fl := range fls {
			fl.InitWithOptions(options)
		}
	}
	return
}

type LogFilter interface {
	FilterSLSLog(logs []*sls.Log) []*sls.Log
	InitWithOptions(options map[string]interface{})
}

type RDSLogFilter struct {
	SlowSQLThreshold time.Duration `mapstructure:"slow_sql_threshold"`
	ExcludeSQL       []string      `mapstructure:"exclude_sql"`
}

func (rf *RDSLogFilter) InitWithOptions(options map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("catch exp: ", err)
		}
	}()

	rf.SlowSQLThreshold = time.Second * 10
	rf.ExcludeSQL = []string{}
	if v, ok := options["slow_sql_threshold"]; ok {
		rf.SlowSQLThreshold, _ = time.ParseDuration(v.(string))
		delete(options, "slow_sql_threshold")
	}

	err := mapstructure.Decode(options, &rf)
	if err != nil {
		fmt.Println("error: ", err)
	}
}

func (rf *RDSLogFilter) FilterSLSLog(logs []*sls.Log) []*sls.Log {
	// filter latency
	res := make([]*sls.Log, 0)
	for _, log := range logs {
		ignore := false
		for _, content := range log.Contents {
			// 过滤特定SQL
			if content.GetKey() == "sql" {
				if InString(content.GetValue(), rf.ExcludeSQL) {
					ignore = true
					break
				}
			}

			// 过滤快SQL
			if content.GetKey() == "latency" && len(content.GetValue()) > 0 {
				val, err := strconv.ParseInt(content.GetValue(), 10, 64) // us
				if err != nil {
					continue
				}
				if val < rf.SlowSQLThreshold.Microseconds() {
					ignore = true
					break
				}
			}
		}

		if !ignore {
			res = append(res, log)
		}
	}
	return res
}

func InString(s string, ss []string) bool {
	for _, item := range ss {
		if s == item {
			return true
		}
	}
	return false
}
