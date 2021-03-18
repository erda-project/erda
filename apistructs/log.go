package apistructs

type GetLogsResponse struct {
	Success bool      `json:"success"`
	Err     string    `json:"err"`
	Data    LogDetail `json:"data"`
}

type LogLine struct {
	Source    string `json:"source"`
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Offset    string `json:"offset"`
	Content   string `json:"content"`
	Level     string `json:"level"`
}

type LogDetail struct {
	Lines []LogLine `json:"lines"`
}

// LogPushRequest 推日志请求
type LogPushRequest struct {
	Lines []LogPushLine
}

// LogPushLine 推日志请求行
type LogPushLine struct {
	ID        string      `json:"id"`
	Source    string      `json:"source"`
	Timestamp int64       `json:"timestamp"`
	Content   string      `json:"content"`
	Stream    *string     `json:"stream,omitempty"`
	Offset    *int        `json:"offset,omitempty"`
	Tags      interface{} `json:"tags,omitempty"`
}

var (
	CollectorLogPushStreamStdout = "stdout"
	CollectorLogPushStreamStderr = "stderr"
)
