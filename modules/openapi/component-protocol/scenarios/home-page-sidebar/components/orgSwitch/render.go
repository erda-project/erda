package orgSwitch

import (
	"context"
	"fmt"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/sirupsen/logrus"
	"strconv"
)

const (
	DefaultType = "DropdownSelect"
	DefaultPageSize = 100
	DefaultPrefixIcon = "ISSUE_ICON.severity.XX"
	DefaultPriorityIcon = "ISSUE_ICON.severity.HIGH"
)

func RenderCreator() protocol.CompRender {
	return &OrgSwitch{}
}

type OrgSwitch struct {
	ctxBdl protocol.ContextBundle
	Type string `json:"type"`
	Props props `json:"props"`
	//Data Data `json:"data"`
	State State `json:"state"`
}

type Data struct {
	List []OrgItem `json:"list"`
}

type state struct {
	//OrgID string `json:"orgID"`
	//PrefixImage string `json:"prefixImage"`
}

type OrgItem struct {
	ID string `json:"id"`
	Name string `json:"name"`
	DisplayName string `json:"display_name"`
	IsPublic bool `json:"is_public"`
	Status string `json:"status"`
	Logo string `json:"logo"`
}

type props struct {
	Visible bool `json:"visible"`
	Options []MenuItem `json:"options"`
	QuickSelect []interface{} `json:"quickSelect"`
}

type MenuItem struct {
	//Name string `json:"name"`
	//Key string `json:"key"`
	//ImgSrc string `json:"ImgSrc"`
	Label string `json:"label"`
	Value string `json:"value"`
	PrefixImgSrc string `json:"prefixImgSrc"`
	Operations map[string]interface{} `json:"operations"`
}

type Meta struct {
	Id string `json:"id"`
	Severity string `json:"severity"`
}

type State struct {
	Value string `json:"value"`
}

type Operation struct {
	Key string `json:"key"`
	Reload bool `json:"reload"`
	Disabled bool `json:"disabled"`
	Text string `json:"text"`
	PrefixIcon string `json:"prefixIcon"`
	Meta Meta `json:"meta"`
}

func (this *OrgSwitch) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

//func (this *OrgSwitch) setComponentValue(orgID string) error {
//	//this.Props = props{}
//	//this.Data = Data{}
//	this.Type = DefaultType
//	if len(this.Data.List) == 0 {
//		this.Props.Visible = false
//		return nil
//	}
//	var (
//		orgName string
//	)
//	if orgID != "" {
//		orgDTO, err := this.ctxBdl.Bdl.GetOrg(orgID)
//		if err != nil {
//			return err
//		}
//		if orgDTO == nil {
//			return fmt.Errorf("can not get org")
//		}
//		orgName = orgDTO.Name
//	} else {
//		orgID = this.Data.List[0].ID
//		orgName = this.Data.List[0].Name
//	}
//	this.setProps(orgName, orgID)
//	//this.State.OrgID = orgID
//	//this.State.PrefixImage = orgLogo
//	return nil
//}

//func (this *OrgSwitch) setProps(orgName, orgID string){
//	this.Props.Visible = true
//	this.Props.Value = orgName
//	this.Props.PrefixIcon =DefaultPrefixIcon
//	this.Props.Operations = make(map[string]Operation)
//	priorityOperation := Operation{
//		Key: string(apistructs.ISTChangePriority),
//		Reload: true,
//		Disabled: false,
//		Text: orgName,
//		PrefixIcon: DefaultPriorityIcon,
//		Meta: Meta{
//			Id: orgID,
//			Severity: "1",
//		},
//	}
//	this.Props.Operations[string(apistructs.ISTChangePriority)] = priorityOperation
//	return
//}

func (this *OrgSwitch) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}
	if this.ctxBdl.Identity.OrgID == "" {
		this.Props.Visible = false
		return nil
	}
	this.Type = "DropdownSelect"
	this.Props.Visible = true
	switch event.Operation {
	case apistructs.InitializeOperation:
		//if err := this.setComponentValue(""); err != nil {
		//	return err
		//}
		if err := this.RenderList(); err != nil {
			return err
		}
		orgDTO, err := this.ctxBdl.Bdl.GetOrg(this.ctxBdl.Identity.OrgID)
		if err != nil {
			return err
		}
		if orgDTO == nil {
			return fmt.Errorf("can not get org")
		}
		this.State.Value = strconv.FormatInt(int64(orgDTO.ID), 10)
		this.Props.QuickSelect = []interface{}{
			map[string]interface{}{
				"value": "orgList",
				"label": "浏览公开组织",
				"operations": map[string]interface{}{
					"click": map[string]interface{}{
						"key": "click",
						"show": false,
						"reload": false,
						"command": map[string]interface{}{
							"key": "goto",
							"target": "orgList",
							"jumpOut": false,
						},
					},
				},
			},
		}
	case apistructs.ChangePriority:
		//orgID, ok := event.OperationData["id"].(string)
		//if !ok {
		//	return fmt.Errorf("invalid operation id")
		//}
		//orgID := this.Props.Operations[string(apistructs.ISTChangePriority)].Meta.Id
		//if err := this.setComponentValue(orgID); err != nil {
		//	return err
		//}
	}
	return nil
}

func RenItem(org apistructs.OrgDTO) MenuItem {
	logo := "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQYQY0vUTJwftJ8WqXoLiLeB--2MJkpZLpYOA&usqp=CAU"
	if org.Logo != "" {
		logo = fmt.Sprintf("https:%s", org.Logo)
	}
	item := MenuItem{
		Label: org.DisplayName,
		Value: strconv.FormatInt(int64(org.ID), 10),
		PrefixImgSrc: logo,
		Operations: map[string]interface{}{
			"click": map[string]interface{}{
				"key": "click",
				"show": false,
				"reload": false,
				"command": map[string]interface{}{
					"key": "goto",
					"target": "orgRoot",
					"jumpOut": false,
					//"orgName": fmt.Sprintf("https://dice.dev.terminus.io/%s", org.Domain),
					"state": map[string]interface{}{
						"params": map[string]interface{}{
							"orgName": org.Name,
						},
					},
				},
			},
		},
	}
	return item
}

func (this *OrgSwitch) RenderList() error {
	identity := apistructs.IdentityInfo{UserID: this.ctxBdl.Identity.UserID}
	req := &apistructs.OrgSearchRequest{
		IdentityInfo: identity,
		PageSize: DefaultPageSize,
	}
	pagingOrgDTO, err := this.ctxBdl.Bdl.ListOrgs(req)
	if err != nil {
		return err
	}
	this.Props.Options = make([]MenuItem, 0)
	for _, v := range pagingOrgDTO.List {
		this.Props.Options = append(this.Props.Options, RenItem(v))
	}
	return nil
}