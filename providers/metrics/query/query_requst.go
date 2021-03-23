package query

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

type TimeAlignCondition string

type FuncCondition struct {
	Name, Field string
}

type GroupCondition string

type FilterCondition struct {
	Key, Op, Val string
}

type FieldCondition struct {
	key, op, val string
	req          *MetricQueryRequest
}

type RangeCondition struct {
	Name   string // 区间聚合名
	Ranges []struct {
		From int
		To   int
	}
	// 下述参数需配合使用
	Split     int     // 区间个数
	RangeSize float64 // 区间范围
}

type MetricQueryRequest struct {
	diagram    string
	scope      string
	format     string
	start, end time.Time
	point      int
	align      TimeAlignCondition
	filters    []FilterCondition
	fieldMaps  []FieldCondition
	groups     []GroupCondition
	limits     []int
	functions  []FuncCondition
	sorts      []string
	range_     *RangeCondition
}

func CreateQueryRequest(metricName string) *MetricQueryRequest {
	now := time.Now()
	start := now.Add(-time.Hour * 1)
	req := &MetricQueryRequest{
		scope:     metricName,
		start:     start,
		end:       now,
		point:     10,
		filters:   make([]FilterCondition, 0),
		groups:    make([]GroupCondition, 0),
		functions: make([]FuncCondition, 0),
		sorts:     make([]string, 0),
		range_:    nil,
	}
	return req
}


func (req *MetricQueryRequest) StartFrom(start time.Time) *MetricQueryRequest {
	req.start = start
	return req
}

func (req *MetricQueryRequest) EndWith(end time.Time) *MetricQueryRequest {
	req.end = end
	return req
}

// 设置数据返回的格式
// raw: 原始数据格式
// chart: 图表格式(默认)
// chartv2: 图表格式v2
func (req *MetricQueryRequest) FormatAs(format string) *MetricQueryRequest {
	req.format = format
	return req
}

// 设置查询数据类型
// histogram: 查询折线图类型的数据
// 若未设置(默认)：查询单个点的数据，如统计集群中所有的机器数量
func (req *MetricQueryRequest) SetDiagram(diagram string) *MetricQueryRequest {
	req.diagram = diagram
	return req
}

// 条件过滤
// 等同于SQL中的 WHERE id=1
func (req *MetricQueryRequest) Filter(key, val string) *MetricQueryRequest {
	req.filters = append(req.filters, FilterCondition{key, "filter", val})
	return req
}

// 前缀模糊匹配
// 等同于SQL中的 name LIKE an%
func (req *MetricQueryRequest) Match(key, val string) *MetricQueryRequest {
	req.filters = append(req.filters, FilterCondition{key, "match", val})
	return req
}

// 设置映射条件条件
func (req *MetricQueryRequest) Field(key string) *FieldCondition {
	return &FieldCondition{
		req: req,
		key: key,
	}
}

// in条件
// 等同于SQL中的 id IN (1,2,5)
func (req *MetricQueryRequest) In(key string, values []string) *MetricQueryRequest {
	for _, val := range values {
		req.filters = append(req.filters, FilterCondition{key, "in", val})
	}
	return req
}

// 限制返回的数据点数量
// 注意：当diagram为histogram时，必须设置该属性
func (req *MetricQueryRequest) LimitPoint(point int) *MetricQueryRequest {
	req.point = point
	return req
}

// 设置聚合函数
// 具体函数请参考：https://yuque.antfin.com/spot/develop-docs/hr2c1y#f14b2b31
func (req *MetricQueryRequest) Apply(funcName, field string) *MetricQueryRequest {
	req.functions = append(req.functions, FuncCondition{funcName, field})
	return req
}

// 设置聚合条件
func (req *MetricQueryRequest) GroupBy(groups []string) *MetricQueryRequest {
	for _, g := range groups {
		req.groups = append(req.groups, GroupCondition(g))
	}
	return req
}

// 设置group聚合后的结果数量
func (req *MetricQueryRequest) LimitGroup(limit int) *MetricQueryRequest {
	req.limits = append(req.limits, limit)
	return req
}

// 设置排序条件
func (req *MetricQueryRequest) Sort(field string) *MetricQueryRequest {
	req.sorts = append(req.sorts, field)
	return req
}

// 设置时间对齐
// 类似于整数取整
func (req *MetricQueryRequest) Align(align TimeAlignCondition) *MetricQueryRequest {
	req.align = align
	return req
}

func (req *MetricQueryRequest) ConstructParam() *url.Values {
	param := url.Values{}
	if req.point > 0 {
		param.Add("points", strconv.Itoa(req.point))
	}
	if req.format != "" {
		param.Add("format", req.format)
	}
	// mile second
	param.Add("start", strconv.Itoa(int(req.start.Unix())*1000))
	param.Add("end", strconv.Itoa(int(req.end.Unix())*1000))
	for _, f := range req.filters {
		param.Add(strings.Join([]string{f.Op, f.Key}, "_"), f.Val)
	}
	for _, f := range req.functions {
		param.Add(f.Name, f.Field)
	}

	for _, g := range req.groups {
		param.Add("group", string(g))
	}
	for _, v := range req.limits {
		param.Add("limit", strconv.Itoa(v))
	}

	for _, fie := range req.fieldMaps {
		param.Add(strings.Join([]string{"field", fie.op, fie.key}, "_"), fie.val)
	}

	if req.range_ != nil {
		// todo
	}

	return &param
}

