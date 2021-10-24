package exception

type Erda_error struct {
	TerminusKey   string            `json:"terminus_key"`
	ApplicationId string            `json:"application_id"`
	ServiceName   string            `json:"service_name"`
	ErrorId       string            `json:"error_id"`
	Timestamp     int64             `json:"timestamp"`
	Tags          map[string]string `json:"tags"`
}
