package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/text/message"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) ListLabels(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	i18n := ctx.Value("i18nPrinter").(*message.Printer)
	return mkResponse(apistructs.ListLabelsResponse{
		Header: apistructs.Header{Success: true},
		Data: []apistructs.ListLabelsData{
			{
				Name:       i18n.Sprintf("locked"),
				Label:      "locked",
				Desc:       "",
				Group:      "others",
				GroupName:  i18n.Sprintf("others"),
				GroupLevel: 5,
			},
			{
				Name:       i18n.Sprintf("platform"),
				Label:      "platform",
				Desc:       "",
				Group:      "platform",
				GroupName:  i18n.Sprintf("platform"),
				GroupLevel: 1,
			},
			{
				Name:       i18n.Sprintf("pack-job"),
				Label:      "pack-job",
				Desc:       "",
				Group:      "job",
				GroupName:  i18n.Sprintf("job"),
				GroupLevel: 4,
			},
			{
				Name:       i18n.Sprintf("bigdata-job"),
				Label:      "bigdata-job",
				Desc:       "",
				Group:      "job",
				GroupName:  i18n.Sprintf("job"),
				GroupLevel: 4,
			},
			{
				Name:       i18n.Sprintf("job"),
				Label:      "job",
				Desc:       "",
				Group:      "job",
				GroupName:  i18n.Sprintf("job"),
				GroupLevel: 4,
			},
			{
				Name:       i18n.Sprintf("stateful-service"),
				Label:      "stateful-service",
				Desc:       "",
				Group:      "service",
				GroupName:  i18n.Sprintf("service"),
				GroupLevel: 3,
			},
			{
				Name:       i18n.Sprintf("stateless-service"),
				Label:      "stateless-service",
				Desc:       "",
				Group:      "service",
				GroupName:  i18n.Sprintf("service"),
				GroupLevel: 3,
			},
			{
				Name:       i18n.Sprintf("workspace-dev"),
				Label:      "workspace-dev",
				Desc:       "",
				Group:      "env",
				GroupName:  i18n.Sprintf("env"),
				GroupLevel: 2,
			},
			{
				Name:       i18n.Sprintf("workspace-test"),
				Label:      "workspace-test",
				Desc:       "",
				Group:      "env",
				GroupName:  i18n.Sprintf("env"),
				GroupLevel: 2,
			},
			{
				Name:       i18n.Sprintf("workspace-staging"),
				Label:      "workspace-staging",
				Desc:       "",
				Group:      "env",
				GroupName:  i18n.Sprintf("env"),
				GroupLevel: 2,
			},
			{
				Name:       i18n.Sprintf("workspace-prod"),
				Label:      "workspace-prod",
				Desc:       "",
				Group:      "env",
				GroupName:  i18n.Sprintf("env"),
				GroupLevel: 2,
			},
			{
				Name:       i18n.Sprintf("org-"),
				Label:      "org-",
				Desc:       "",
				IsPrefix:   true,
				Group:      "others",
				GroupName:  i18n.Sprintf("others"),
				GroupLevel: 5,
			},
			{
				Name:       i18n.Sprintf("location-"),
				Label:      "location-",
				Desc:       "",
				IsPrefix:   true,
				Group:      "others",
				GroupName:  i18n.Sprintf("others"),
				GroupLevel: 5,
			},
			{
				Name:       i18n.Sprintf("topology-zone"),
				Label:      "topology-zone",
				Desc:       "",
				Group:      "others",
				GroupName:  i18n.Sprintf("others"),
				GroupLevel: 5,
				WithValue:  true,
			},
		},
	})
}

func (e *Endpoints) UpdateLabels(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.UpdateLabelsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("failed to unmarshal to apistructs.UpdateLabelsRequest: %v", err)
		return mkResponse(apistructs.UpdateLabelsResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	userid := r.Header.Get("User-ID")
	if userid == "" {
		errstr := fmt.Sprintf("failed to get user-id in http header")
		return mkResponse(apistructs.UpdateLabelsResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	recordID, err := e.labels.UpdateLabels(req, userid)
	if err != nil {
		errstr := fmt.Sprintf("failed to updatelabels: %v", err)
		return mkResponse(apistructs.UpdateLabelsResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	return mkResponse(apistructs.UpdateLabelsResponse{
		Header: apistructs.Header{
			Success: true,
		},
		Data: apistructs.UpdateLabelsData{
			RecordID: recordID,
		},
	})
}
