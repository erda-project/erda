// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package recordTable

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/conf"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-space-list/i18n"
)

type Column struct {
	DataIndex string `json:"dataIndex"`
	Title     string `json:"title"`
	Width     uint64 `json:"width,omitempty"`
}

type Props struct {
	Columns []Column `json:"columns"`
}

type Status struct {
	RenderType string `json:"renderType"`
	Value      string `json:"value"`
	Status     string `json:"status"`
}

type Result struct {
	RenderType string `json:"renderType"`
	URL        string `json:"url"`
	Value      string `json:"value"`
}

type DataItem struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Operator string `json:"operator"`
	Time     string `json:"time"`
	Desc     string `json:"desc"`
	Status   Status `json:"status"`
	Result   Result `json:"result"`
}

type Data struct {
	List []DataItem `json:"list"`
}

type State struct {
	AutoRefresh bool `json:"autoRefresh"`
	Visible     bool `json:"visible"`
}

type RecordTable struct {
	ctxBdl protocol.ContextBundle

	Type  string `json:"type"`
	Props Props  `json:"props"`
	Data  Data   `json:"data"`
	State State  `json:"state"`
}

func (r *RecordTable) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	r.ctxBdl = bdl
	return nil
}

func (r *RecordTable) GenComponentState(c *apistructs.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	r.State = state
	return nil
}

func (r *RecordTable) setProps() {
	i18nLocale := r.ctxBdl.Bdl.GetLocale(r.ctxBdl.Locale)
	r.Props.Columns = make([]Column, 0)
	r.Props.Columns = append(r.Props.Columns, Column{
		DataIndex: "id",
		Title:     "ID",
		Width:     80,
	}, Column{
		DataIndex: "type",
		Title:     i18nLocale.Get(i18n.I18nKeyTableType),
		Width:     80,
	}, Column{
		DataIndex: "operator",
		Title:     i18nLocale.Get(i18n.I18nKeyTableOperator),
		Width:     150,
	}, Column{
		DataIndex: "time",
		Title:     i18nLocale.Get(i18n.I18nKeyTableTime),
		Width:     150,
	}, Column{
		DataIndex: "desc",
		Title:     i18nLocale.Get(i18n.I18nKeyTableDesc),
	}, Column{
		DataIndex: "status",
		Title:     i18nLocale.Get(i18n.I18nKeyTableStatus),
		Width:     80,
	}, Column{
		DataIndex: "result",
		Title:     i18nLocale.Get(i18n.I18nKeyTableResult),
	})
}

func (r *RecordTable) setData() error {
	projectID, ok := r.ctxBdl.InParams["projectId"].(float64)
	if !ok {
		return errors.Errorf("invalid projectID: %v", r.ctxBdl.InParams["projectId"])
	}
	rsp, err := r.ctxBdl.Bdl.ListFileRecords(r.ctxBdl.Identity.UserID, apistructs.ListTestFileRecordsRequest{
		ProjectID: uint64(projectID),
		Types:     []apistructs.FileActionType{apistructs.FileSpaceActionTypeImport, apistructs.FileSpaceActionTypeExport},
		Locale:    r.ctxBdl.Locale,
	})
	if err != nil {
		return err
	}

	r.Data.List = make([]DataItem, 0)
	i18nLocale := r.ctxBdl.Bdl.GetLocale(r.ctxBdl.Locale)
	for _, fileRecord := range rsp.Data.List {
		var recordTypeKey string
		switch fileRecord.Type {
		case apistructs.FileSpaceActionTypeImport:
			recordTypeKey = i18n.I18nKeyImport
		case apistructs.FileSpaceActionTypeExport:
			recordTypeKey = i18n.I18nKeyExport
		}
		var statusKey string
		var recordState string
		switch fileRecord.State {
		case apistructs.FileRecordStateFail:
			statusKey = i18n.I18nKeyStatusFailed
			recordState = "error"
		case apistructs.FileRecordStateSuccess:
			statusKey = i18n.I18nKeyStatusSuccess
			recordState = "success"
		case apistructs.FileRecordStatePending:
			statusKey = i18n.I18nKeyStatusPending
			recordState = "pending"
		case apistructs.FileRecordStateProcessing:
			statusKey = i18n.I18nKeyStatusProcessing
			recordState = "processing"
		}
		var operatorName string
		operator, err := r.ctxBdl.Bdl.GetCurrentUser(fileRecord.OperatorID)
		if err == nil {
			operatorName = operator.Nick
		}
		r.Data.List = append(r.Data.List, DataItem{
			ID:       strconv.FormatInt(int64(fileRecord.ID), 10),
			Type:     i18nLocale.Get(recordTypeKey),
			Operator: operatorName,
			Time:     fileRecord.CreatedAt.Format("2006-01-02 15:04:05"),
			Desc:     fileRecord.Description,
			Status: Status{
				RenderType: "textWithBadge",
				Value:      i18nLocale.Get(statusKey),
				Status:     recordState,
			},
			Result: Result{
				RenderType: "downloadUrl",
				URL:        fmt.Sprintf("%s/api/files/%s", conf.RootDomain(), fileRecord.ApiFileUUID),
				Value:      fileRecord.FileName,
			},
		})
	}
	return nil
}

func (r *RecordTable) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := r.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := r.GenComponentState(c); err != nil {
		return err
	}
	r.setProps()
	if r.State.Visible || r.State.AutoRefresh {
		if err := r.setData(); err != nil {
			return err
		}
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &RecordTable{}
}
