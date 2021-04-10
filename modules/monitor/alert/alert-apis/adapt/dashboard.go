package adapt

import (
	"fmt"
	"strings"
	"time"

	"github.com/erda-project/erda/modules/monitor/utils"
)

func NewDashboard(a *Adapt) *dashgen {
	return &dashgen{a: a}
}

type dashgen struct {
	a                    *Adapt
	preview              bool
	lang, scope, scopeID string
}

func (d *dashgen) CreateChartDashboard(alertDetail *CustomizeAlertDetail) (string, error) {

	d.init(alertDetail)
	return d.createChartDashboard(alertDetail)
}

func (d *dashgen) GenerateDashboardPreView(alertDetail *CustomizeAlertDetail) (res *block.View, err error) {
	if len(alertDetail.Rules) != 1 {
		return nil, fmt.Errorf("must be only one view")
	}
	d.preview = true
	d.init(alertDetail)
	vitems, err := d.generateViewConfigItems(alertDetail.Rules[0])
	if err != nil {
		return nil, err
	}
	return vitems[0].View, nil
}

func (d *dashgen) createChartDashboard(alertDetail *CustomizeAlertDetail) (string, error) {
	block, err := d.generateDashboard(alertDetail)
	if err != nil {
		return "", fmt.Errorf("generate dashboard fialed. err=%s", err)
	}
	dash, err := d.a.dashboardAPI.CreateDashboard(block)
	if err != nil {
		return "", err
	}
	return dash.ID, nil
}

func (d *dashgen) init(alertDetail *CustomizeAlertDetail) {
	d.lang = alertDetail.Lang
	d.scopeID = alertDetail.AlertScopeID
	d.scope = alertDetail.AlertScope

	if d.scope == "org" {
		org, err := d.a.bdl.GetOrg(d.scopeID)
		if err != nil {
			d.a.l.Infof("failed to get org by orgId=%s", d.scopeID)
		}
		d.scopeID = org.Name
	}
}

func (d *dashgen) generateDashboard(alertDetail *CustomizeAlertDetail) (res *block.UserBlock, err error) {
	block := &block.UserBlock{
		Name:    "",
		Scope:   fmt.Sprintf("%s:alert", d.scope),
		ScopeID: d.scopeID,
	}
	viewConfig, err := d.createViewConfig(alertDetail.Rules)
	if err != nil {
		return nil, err
	}
	block.ViewConfig = &viewConfig

	return block, nil
}

func (d *dashgen) createViewConfig(rules []*CustomizeAlertRule) (vc block.ViewConfigDTO, err error) {
	res := []*block.ViewConfigItem{}
	for _, rule := range rules {
		vci, err := d.generateViewConfigItems(rule)
		if err != nil {
			return nil, err
		}
		res = append(res, vci...)
	}
	return res, nil
}

func (d *dashgen) generateViewConfigItems(r *CustomizeAlertRule) (res []*block.ViewConfigItem, err error) {
	res = []*block.ViewConfigItem{}
	heigth := 10
	for idx, f := range r.Functions {
		ct, err := d.getChartType(f)
		if err != nil {
			return nil, err
		}
		_url, err := d.getURL(ct, r.Metric)
		query, err := d.generateQuery(r)
		if err != nil {
			return nil, err
		}
		view := &block.View{
			Title:     "",
			ChartType: ct,
			API: &block.API{
				URL:       _url,
				Query:     query,
				Body:      nil,
				Header:    nil,
				ExtraData: d.generateExtraData(query, f, r.ActivedMetricGroups, r.Metric),
			},
		}
		vci := &block.ViewConfigItem{
			W:    24,
			H:    int64(heigth * (idx + 1)),
			X:    0,
			Y:    int64(heigth * idx),
			I:    fmt.Sprintf("view-%s", utils.RandomString(8)),
			View: view,
		}
		res = append(res, vci)
	}
	return res, nil
}

func (d *dashgen) getURL(ct string, field string) (string, error) {
	path := ""
	switch d.scope {
	case "org":
		path = "/api/orgCenter/metrics"
	case "micro_service":
		path = "/api/tmc/metrics"
	default:
		return "", fmt.Errorf("invalid scope %s", d.scope)
	}
	res := fmt.Sprintf("%s/%s/histogram", path, field)
	if ct == "table" {
		res = fmt.Sprintf("/api/orgCenter/metrics/%s", field)
	}
	return res, nil
}

func (d *dashgen) getChartType(f *CustomizeAlertRuleFunction) (string, error) {
	ct := "chart:line"
	agg, err := d.a.metricq.GetSingleAggregationMeta("en", "analysis", f.Aggregator)
	if err != nil {
		d.a.l.Errorf("cant find agg %s", f.Aggregator)
	}
	if agg.ResultType != "number" {
		ct = "table"
	}
	return ct, nil
}

func (d *dashgen) generateExtraData(query map[string]interface{}, f *CustomizeAlertRuleFunction, groups []string, metric string) map[string]interface{} {
	res := make(map[string]interface{})
	res["aggregation"] = f.Aggregator
	res["metric"] = strings.Join([]string{metric, f.Field}, "-")

	group := []string{}
	if val, ok := query["group"]; ok {
		val = strings.TrimLeft(val.(string), "(")
		val = strings.TrimRight(val.(string), ")")
		for _, vv := range strings.Split(val.(string), ",") {
			group = append(group, vv)
		}
	}
	res["group"] = group

	filters := []map[string]string{}
	for k, v := range query {
		idx := strings.Index(k, "_")
		if idx == -1 {
			continue
		}
		if _, ok := v.(string); !ok {
			continue
		}
		method, tag := k[:idx], k[idx+1:]
		if strings.HasPrefix(tag, "tags.") {
			filters = append(filters, map[string]string{
				"key":    utils.RandomString(1),
				"method": method,
				"tag":    tag,
				"value":  v.(string),
			})
		}
	}
	res["filters"] = filters

	res["activedMetricGroups"] = groups
	return res
}

func (d *dashgen) generateQuery(rule *CustomizeAlertRule) (map[string]interface{}, error) {
	metrics, err := d.a.metricq.MetricMeta(d.lang, d.scope, d.scopeID, rule.Metric)
	if err != nil {
		return nil, err
	} else if metrics == nil {
		return nil, fmt.Errorf("%s has no metric meta", rule.Metric)
	}
	metric := metrics[0]
	res := make(map[string]interface{})
	res["start"] = utils.ConvertTimeToMS(time.Now().Add(-1 * time.Hour))
	res["end"] = utils.ConvertTimeToMS(time.Now())

	// 转换group
	if len(rule.Group) > 0 {
		group := "("
		for i, item := range rule.Group {
			if i > 0 {
				group += ","
			}
			name := "tags." + item
			group += name
			if !d.preview {
				res["filter_"+name] = "{{" + item + "}}"
			}
		}
		group += ")"
		res["group"] = group
	}

	// 转换filter
	for _, item := range rule.Filters {
		res[item.Operator+"_tags."+item.Tag] = item.Value
	}
	// 微服务，告警预览需加强制过滤
	if d.scope == "micro_service" && d.preview {
		res["filter_terminus_key"] = d.scopeID
	}

	// 转换function
	if rule.Functions != nil {
		function := rule.Functions[0]
		name := function.Field
		if strings.HasPrefix("tags.", function.Field) {
			if t, ok := metric.Tags[function.Field[len("tags."):]]; ok {
				name = t.Name
			}
		} else if strings.HasPrefix("fields.", function.Field) {
			if f, ok := metric.Fields[function.Field[len("fields."):]]; ok {
				name = f.Name
			}
		} else {
			if f, ok := metric.Fields[function.Field]; ok {
				name = f.Name
			}
		}
		res[function.Aggregator] = function.Field
		res["alias_"+function.Aggregator+"."+function.Field] = name
	}
	return res, nil
}
