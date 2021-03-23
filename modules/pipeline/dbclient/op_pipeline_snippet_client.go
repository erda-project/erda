package dbclient

type DicePipelineSnippetClient struct {
	ID    uint64                     `json:"id" xorm:"pk autoincr"`
	Name  string                     `json:"name"`
	Host  string                     `json:"host"`
	Extra PipelineSnippetClientExtra `json:"extra" xorm:"json"`
}

func (ps *DicePipelineSnippetClient) TableName() string {
	return "dice_pipeline_snippet_clients"
}

type PipelineSnippetClientExtra struct {
	UrlPathPrefix string `json:"urlPathPrefix"`
}

func (client *Client) FindSnippetClientList() (clients []*DicePipelineSnippetClient, err error) {

	err = client.Find(&clients)
	if err != nil {
		return nil, err
	}

	return clients, err
}
