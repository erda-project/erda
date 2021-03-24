package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/erda-project/erda/pkg/telemetry/promxp"
)

var (
	failCounter = promxp.RegisterCounter(
		"fail_counter",
		"number of failures",
		map[string]string{},
	)
	currentProcessingGauge = promxp.RegisterGauge(
		"current_processing",
		"requests currently being processed",
		map[string]string{},
	)
	apiDuration = promxp.RegisterSummary(
		"api_duration",
		"API request time-consuming, counted by percentile",
		map[string]string{},
		map[float64]float64{0.3: 0.05, 0.5: 0.01, 0.7: 0.01, 0.90: 0.01},
	)
	processDuration = promxp.RegisterHistogram(
		"process_duration",
		"processing time, statistics by value range",
		map[string]string{},
		[]float64{200, 500, 800, 900},
	)
	testMeter = promxp.RegisterMeter(
		"test_meter",
		"test rate index",
		map[string]string{},
	)
	testAutoResetCounter = promxp.RegisterAutoResetCounter(
		"test_auto_reset_counter",
		"test index",
		map[string]string{},
		"dice", "local",
	)
)

var (
	testCounter = promxp.RegisterCounterVec(
		"test_counter",
		"test index",
		map[string]string{},
		[]string{"action"},
	)
	testGauge = promxp.RegisterGaugeVec(
		"test_gauge",
		"test index",
		map[string]string{},
		[]string{"action"},
	)
	testSummaryVec = promxp.RegisterSummaryVec(
		"test_summary2",
		"test index",
		map[string]string{},
		[]string{"action"},
		map[float64]float64{0.3: 0.05, 0.5: 0.01, 0.7: 0.01, 0.90: 0.01},
	)
	testMeterVec = promxp.RegisterMeterVec(
		"test_meter_vec",
		"test index",
		map[string]string{},
		[]string{"org_id", "project_id", "application_id", "workspace"},
	)
	testAutoResetCounterVec = promxp.RegisterAutoResetCounterVec(
		"test_auto_reset_counter_vec",
		"test index",
		map[string]string{},
		[]string{"component", "org_id", "project_id", "application_id", "workspace"},
		"dice", "local",
	)
)

func main() {
	go func() {
		var i int
		for {
			if i%2 == 0 {
				testCounter.WithLabelValues("case1").Add(1)
			} else {
				testCounter.WithLabelValues("case2").Add(1)
			}
			testMeterVec.WithLabelValues("1", "2", "3", "4").Mark(3)
			testMeter.Mark(1)
			testAutoResetCounter.Inc()
			i++
			randSleep()
		}
	}()

	go func() {
		var i int
		for {
			if i%2 == 0 {
				testGauge.WithLabelValues("case1").Set(float64(i % 10))
			} else {
				testGauge.WithLabelValues("case2").Set(float64(i % 8))
			}
			testSummaryVec.WithLabelValues("case3").Observe(1)
			for x := 0; x < 15; x++ {
				testAutoResetCounterVec.WithLabelValues("con-1", "org-2", "pro-3", "app-4", strconv.Itoa(x)).Inc()
			}
			testMeter.Mark(2)
			testAutoResetCounter.Inc()
			i++
			randSleep()
		}
	}()
	apiDuration.Observe(1)
	apiDuration.Observe(2)
	apiDuration.Observe(3)
	apiDuration.Observe(4)
	apiDuration.Observe(5)
	apiDuration.Observe(6)
	apiDuration.Observe(7)
	apiDuration.Observe(8)
	apiDuration.Observe(99)
	apiDuration.Observe(100)

	processDuration.Observe(200)
	processDuration.Observe(800)

	metricName := "example"
	err := promxp.Start(":8080", metricName)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

var start = time.Now()

func randSleep() {
	oscillationFactor := func() float64 {
		return 2 + math.Sin(math.Sin(2*math.Pi*float64(time.Since(start))/float64(10*time.Minute)))
	}
	time.Sleep(time.Duration(50*oscillationFactor()) * time.Millisecond)
}
