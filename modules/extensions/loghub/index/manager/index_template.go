package manager

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/olivere/elastic"
	cfgpkg "github.com/recallsong/go-utils/config"
)

func (p *provider) setupIndexTemplate(client *elastic.Client) error {
	if len(p.C.IndexTemplateName) <= 0 || len(p.C.IndexTemplateFile) <= 0 {
		return nil
	}
	template, err := ioutil.ReadFile(p.C.IndexTemplateFile)
	if err != nil {
		return fmt.Errorf("fail to load index template: %s", err)
	}
	template = cfgpkg.EscapeEnv(template)
	body := string(template)
	p.L.Info("load index template: \n", body)
	for i := 0; i < 2; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		resp, err := client.IndexPutTemplate(p.C.IndexTemplateName).
			BodyString(body).Do(ctx)
		if err != nil {
			cancel()
			return fmt.Errorf("fail to set index template: %s", err)
		}
		cancel()
		if resp.Acknowledged {
			break
		}
	}
	p.L.Infof("Put index template (%s) success", p.C.IndexTemplateName)
	return nil
}
