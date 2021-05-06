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

package metronome

type MetronomeJob struct {
	Id          string            `json:"id"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Run         Run               `json:"run,omitempty"`
}

type JobHistory struct {
	FailedFinishedRuns     []RunResult `json:"failedFinishedRuns,omitempty"`
	SuccessfulFinishedRuns []RunResult `json:"successfulFinishedRuns,omitempty"`
}

// MetronomeJobResult used when GET Metronome job with "embed=history&embed=activeRuns"
type MetronomeJobResult struct {
	MetronomeJob
	ActiveRuns []ActiveRun `json:"activeRuns,omitempty"`
	History    JobHistory  `json:"history,omitempty"`
}

// ActiveRun metronome job activeRun struct, GET with "embed=activeRuns"
type ActiveRun struct {
	ID     string `json:"id,omitempty"`
	JobID  string `json:"jobId,omitempty"`
	Status string `json:"status,omitempty"`
}

type Run struct {
	Id             string            `json:"id,omitempty"`
	Artifacts      []Artifact        `json:"artifacts,omitempty"`
	Cmd            string            `json:"cmd,omitempty"`
	Cpus           float64           `json:"cpus"`
	Mem            float64           `json:"mem"` // "minimum": 32
	Disk           float64           `json:"disk"`
	Docker         Docker            `json:"docker,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
	MaxLaunchDelay int               `json:"maxLaunchDelay,omitempty"`
	Placement      *Placement        `json:"placement,omitempty"`
	Restart        Restart           `json:"restart,omitempty"`
	//User           string            `json:"user,omitempty"`
	Volumes []Volume `json:"volumes,omitempty"`
}

type Artifact struct {
	Uri        string `json:"uri,omitempty"`
	Extract    bool   `json:"extract,omitempty"`
	Executable bool   `json:"executable,omitempty"`
	Cache      bool   `json:"cache,omitempty"`
}

// TODO: accomplish the fields
type Docker struct {
	Image          string `json:"image"`
	ForcePullImage bool   `json:"forcePullImage,omitempty"`
	//Parameters     []DockerParameter `json:"parameters,omitempty"`
}

type DockerParameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TODO: accomplish the fields
type Placement struct {
	Constraints []Constraints `json:"constraints,omitempty"`
}

type Constraints struct {
	Attribute string `json:"attribute,omitempty"`
	Operator  string `json:"operator,omitempty"`
	Value     string `json:"value,omitempty"`
}

// TODO: accomplish the fields
type Restart struct {
	Policy                string `json:"policy,omitempty"`
	ActiveDeadlineSeconds int    `json:"activeDeadlineSeconds,omitempty"`
}

type Volume struct {
	ContainerPath string `json:"containerPath,omitempty"`
	HostPath      string `json:"hostPath,omitempty"`
	Mode          string `json:"mode,omitempty"`
}

// TODO: accomplish the fields
type RunResult struct {
	//completedAt string `json:"completedAt,omitempty"`,
	//"createdAt": "2016-07-15T13:02:59.735+0000",
	Id     string `json:"id"`
	JobId  string `json:"jobId"`
	Status string `json:"status"`
	//"tasks": []
}
