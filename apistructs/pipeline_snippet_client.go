package apistructs

type DicePipelineSnippetClient struct {
	ID    uint64                     `json:"id"`
	Name  string                     `json:"name"`
	Host  string                     `json:"host"`
	Extra PipelineSnippetClientExtra `json:"extra"`
}

type PipelineSnippetClientExtra struct {
	UrlPathPrefix string `json:"urlPathPrefix"`
}
