package slsimport

import (
	"context"
	"fmt"

	"github.com/olivere/elastic"
)

const indexTemplate = `
{
    "order":0,
    "index_patterns":[
        "%s*"
    ],
    "settings":{
        "number_of_shards": 1,
        "number_of_replicas": 0,
        "index":{
            "refresh_interval":"15s",
            "translog.durability": "async",
            "translog.sync_interval": "20s",
            "translog.flush_threshold_size": "1024mb"
        }
    },
    "mappings":{
        "logs":{
            "dynamic_templates":[
                {
                    "message_field":{
                        "path_match":"message",
                        "match_mapping_type":"string",
                        "mapping":{
                            "type":"text",
                            "norms":false
                        }
                    }
                },
                {
                    "content_field":{
                        "path_match":"content",
                        "match_mapping_type":"string",
                        "mapping":{
                            "type":"text",
                            "norms":false
                        }
                    }
                },
                {
                    "tags": {
                        "match": "*",
                        "match_mapping_type": "string",
                        "mapping": {
                            "type": "keyword",
                            "ignore_above": 10240
                        }
                    }
                }
            ],
            "properties":{
                "@timestamp":{
                    "type":"date"
                },
                "@version":{
                    "type":"keyword"
                },
                "timestamp":{
                    "type":"long"
                }
            }
        }
    },
    "aliases":{}
}
`

func (p *provider) initIndexTemplate(client *elastic.Client) error {
	if len(p.C.Output.Elasticsearch.IndexTemplateName) <= 0 {
		return fmt.Errorf("index template name is empty")
	}
	template := fmt.Sprintf(indexTemplate, "sls-")
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		resp, err := client.IndexPutTemplate(p.C.Output.Elasticsearch.IndexTemplateName).
			BodyString(template).Do(ctx)
		if err != nil {
			return fmt.Errorf("fail to set index template: %s", err)
		}
		if resp.Acknowledged {
			break
		}
	}
	p.L.Infof("Put index template (%s) success", p.C.Output.Elasticsearch.IndexTemplateName)
	return nil
}
