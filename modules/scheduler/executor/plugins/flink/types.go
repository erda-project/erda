package flink

type FlinkCreateResponse struct {
	JobId string `json:"jobid"`
}

type FlinkGetResponse struct {
	Name        string `json:"name"`
	State       string `json:"state"`
	StartTime   int64  `json:"start-time"`
	CurrentTime int64  `json:"now"`
	EndTime     int64  `json:"end-time"`
}
