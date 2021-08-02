package bundle

import (
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/i18n"
	"testing"
)

func TestBundle_CreateMSPTenant(t *testing.T) {
	type fields struct {
		hc         *httpclient.HTTPClient
		i18nLoader *i18n.LocaleResourceLoader
		urls       urls
	}
	type args struct {
		orgID      string
		userID     string
		projectID  string
		workspace  string
		tenantType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "case1",
			fields: fields{
				hc: httpclient.New(),
			},
			args: args{
				userID:     "1100",
				orgID:      "1",
				projectID:  "123",
				workspace:  "DEV",
				tenantType: "DOP",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bundle{
				hc:         tt.fields.hc,
				i18nLoader: tt.fields.i18nLoader,
				urls:       tt.fields.urls,
			}
			got, err := b.CreateMSPTenant(tt.args.orgID, tt.args.userID, tt.args.projectID, tt.args.workspace, tt.args.tenantType)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateMSPTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateMSPTenant() got = %v, want %v", got, tt.want)
			}
		})
	}
}
