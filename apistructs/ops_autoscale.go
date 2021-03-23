package apistructs

import "time"

const (
	AutoScaleUserID     = "1110"
	AutoScaleLockPrefix = "/autoscale/lock"
)

// mns
type BasicCloudConf struct {
	Region          string `json:"region"`
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
}

type MnsReq struct {
	BasicCloudConf
	ClusterName string `json:"clusterName"`
	AccountId   string `json:"accountId"` //if empty, auto get it
}

type EssActivityMsg struct {
	Content       EssActivityContent `json:"content"`
	Event         string             `json:"event"`
	EventStatus   string             `json:"eventStatus"`
	Product       string             `json:"product"`
	RegionId      string             `json:"regionId"`
	ReceiptHandle string             `json:"receiptHandle"` // mark received msg, used when delete this msg in mns
}

type EssActivityContent struct {
	Cause       string    `json:"cause"`
	Description string    `json:"description"`
	EndTime     time.Time `json:"endTime"`
	StartTime   time.Time `json:"startTime"`
	InstanceIds []string  `json:"instanceIds"`
}

type ScaleInfo struct {
	ReceiptHandles []string
	Instances      map[string]string // instanceId/ip
}

// ecs
type EcsInfoReq struct {
	BasicCloudConf
	InstanceIds []string `json:"instanceIds"`
	PrivateIPs  []string `json:"privateIPs"`
}

type SchedulerScaleReq struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	Region          string `json:"region"`
	ClusterName     string `json:"clusterName"`
	VSwitchID       string `json:"vSwitchID"`
	EcsPassword     string `json:"ecsPassword"`
	SgID            string `json:"sgID"`
	Num             int    `json:"num"`
	LaunchTime      string `json:"launchTime"`
	ScaleDuration   int    `json:"scaleDuration"`
	RecurrenceType  string `json:"recurrenceType"`
	RecurrenceValue string `json:"recurrenceValue"`
	OrgID           int    `json:"orgID"`
	IsEdit          bool   `json:"isEdit"`
	ScheduledTaskId string `json:"scheduledTaskId"`
}
