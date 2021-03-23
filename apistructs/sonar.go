package apistructs

type SonarIssueGetRequest struct {
	Type  string `schema:"type"`
	Key   string `schema:"key"`
	AppID uint64 `schema:"applicationId"`
}
