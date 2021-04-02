package chronos

import (
	"github.com/erda-project/erda/apistructs"
)

type Job struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Command   string   `json:"command"`
	Arguments []string `json:"arguments"`
	Shell     bool     `json:"shell"`
	Async     bool     `json:"async"`
	Retries   int      `json:"retries"`

	Owner     string `json:"owner,omitempty"`
	OwnerName string `json:"ownerName,omitempty"`

	Cpus float64 `json:"cpus"`
	Mem  float64 `json:"mem"`
	Disk float64 `json:"disk"`

	SuccessCount int    `json:"successCount,omitempty"`
	ErrorCount   int    `json:"errorCount,omitempty"`
	LastSuccess  string `json:"lastSuccess,omitempty"`
	LastError    string `json:"lastError,omitempty"`

	Disabled bool `json:"disabled"`

	Schedule         string   `json:"schedule,omitempty"`
	ScheduleTimeZone string   `json:"scheduleTimeZone,omitempty"`
	Epsilon          string   `json:"epsilon,omitempty"`
	Parents          []string `json:"parents,omitempty"`

	Container *JobContainer `json:"container"`

	Fetch                []FetchField `json:"fetch,omitempty"`
	EnvironmentVariables []NVField    `json:"environmentVariables,omitempty"`

	Constraints [][]string `json:"constraints,omitempty"`
}

type FetchField struct {
	Uri        string `json:"uri"`
	Extract    bool   `json:"extract"`
	Executable bool   `json:"executable"`
	Cache      bool   `json:"cache"`
}

type NVField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type KVField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type JobContainer struct {
	Type string `json:"type"`

	Image          string               `json:"image"`
	ForcePullImage bool                 `json:"forcePullImage"`
	Network        string               `json:"network"`
	NetworkName    string               `json:"networkName"`
	Volumes        []JobContainerVolume `json:"volumes,omitempty"`

	Parameters []KVField `json:"parameters"`
}

type JobContainerVolume struct {
	ContainerPath string `json:"containerPath"`
	HostPath      string `json:"hostPath"`
	Mode          string `json:"mode"`
}

type chronosAction string

const (
	chronosCreated chronosAction = "Created"
	chronosDestory chronosAction = "Destory"
)

type destoryInfo struct {
	Name       string                `json:"name"`
	Namespace  string                `json:"namespace"`
	Action     chronosAction         `json:"action"`
	LastStatus apistructs.StatusCode `json:"status"`
}

// summary
type jobSummary struct {
	Jobs []jobStatus `json:"jobs"`
}

// status
type jobStatus struct {
	Name     string   `json:"name"`
	Status   string   `json:"status"`
	State    string   `json:"state"`
	Schedule string   `json:"schedule"`
	Parents  []string `json:"parents"`
	Disabled bool     `json:"disabled"`
}
