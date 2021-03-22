package healthcheck

import (
	"os"
	"sync"
)

var report Report

type Report struct {
	monitors []ReportMonitor
	wait     sync.WaitGroup
	lock     sync.Mutex
	once     sync.Once
}

func OnceDo(fun func()) {
	report.once.Do(fun)
}

type ReportMessage struct {
	Name    string                  `json:"name"`
	Status  Status                  `json:"status"`
	Tags    map[string]string       `json:"tags"`
	Modules []MonitorCollectMessage `json:"modules"`
}

type MonitorCollectMessage struct {
	Name    string `json:"name"`
	Status  Status `json:"status"`
	Message string `json:"message"`
}

type Status string

const Ok Status = "ok"
const Fail Status = "fail"

func (s Status) isOk() bool {
	return s == Ok
}

func (s Status) isFail() bool {
	return !s.isOk()
}

func RegisterMonitor(monitor ReportMonitor) *Report {
	report.monitors = append(report.monitors, monitor)
	return &report
}

func DoReport() ReportMessage {

	var message ReportMessage
	message.Name = os.Getenv("DICE_COMPONENT")
	message.Tags = map[string]string{}
	message.Tags["pod_name"] = os.Getenv("POD_NAME")
	message.Tags["version"] = os.Getenv("DICE_VERSION")

	for _, collect := range report.monitors {
		report.wait.Add(1)
		go func(collect ReportMonitor) {
			defer report.wait.Done()

			collectMessage := collect.Collect()

			if collectMessage != nil {
				report.lock.Lock()
				defer report.lock.Unlock()
				message.Modules = append(message.Modules, collectMessage...)
			}
		}(collect)
	}
	report.wait.Wait()

	message.Status = Ok
	for _, v := range message.Modules {
		if v.Status.isFail() {
			message.Status = v.Status
			break
		}
	}

	return message
}
