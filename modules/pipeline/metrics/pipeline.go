// Package metrics work with monitor metrics.
//
// There are some monitor reserved labels(tags):
// - cluster_name
// - cluster_type
// - component
// - field
// - host
// - host_ip
// - is_edge
// - metric_name
// - org_name
// - version
package metrics

// import (
// 	"sort"
// 	"strconv"

// 	"github.com/prometheus/client_golang/prometheus"

// 	"github.com/erda-project/erda/apistructs"
// 	"github.com/erda-project/erda/modules/pipeline/spec"
// 	"terminus.io/dice/telemetry/metrics"
// 	"terminus.io/dice/telemetry/promxp"
// )

// const (
// 	fieldPipelineTotal      = "pipeline_total"
// 	fieldPipelineProcessing = "pipeline_processing"

// 	labelPipelineID      = "pipeline_id"
// 	labelExecuteCluster  = "execute_cluster"
// 	labelPipelineStatus  = "pipeline_status"
// 	labelPipelineSource  = "pipeline_source"
// 	labelPipelineYmlName = "pipeline_yml_name"

// 	num = "num"
// )

// // pipelineMustLabelNames 受限制于 promp 的用法，可查询的 labelNames 需要提前全量定义好，
// // 因此以后大盘要加新的过滤条件需要在这里追加
// // return: metricKeys(fixed order), metricValues(corresponding to metricKeys order), metricLabelMap
// func generatePipelineMetricLabels(p spec.Pipeline) ([]string, []string, map[string]string) {
// 	pLabels := p.MergeLabels()

// 	metricLabelMap := map[string]string{
// 		// pipeline base
// 		labelPipelineID:                     strconv.FormatUint(p.ID, 10),
// 		labelPipelineYmlName:                p.PipelineYmlName,
// 		labelExecuteCluster:                 p.ClusterName,
// 		labelPipelineSource:                 p.PipelineSource.String(),
// 		labelPipelineStatus:                 p.Status.String(),
// 		apistructs.LabelPipelineTriggerMode: p.TriggerMode.String(),
// 		// tenant
// 		apistructs.LabelOrgID:         pLabels[apistructs.LabelOrgID],
// 		apistructs.LabelOrgName:       pLabels[apistructs.LabelOrgName],
// 		apistructs.LabelProjectID:     pLabels[apistructs.LabelProjectID],
// 		apistructs.LabelProjectName:   pLabels[apistructs.LabelProjectName],
// 		apistructs.LabelAppID:         pLabels[apistructs.LabelAppID],
// 		apistructs.LabelAppName:       pLabels[apistructs.LabelAppName],
// 		apistructs.LabelDiceWorkspace: pLabels[apistructs.LabelDiceWorkspace],
// 		// repo
// 		apistructs.LabelBranch: pLabels[apistructs.LabelBranch],
// 		// fdp
// 		apistructs.LabelFdpWorkflowID:          pLabels[apistructs.LabelFdpWorkflowID],
// 		apistructs.LabelFdpWorkflowName:        pLabels[apistructs.LabelFdpWorkflowName],
// 		apistructs.LabelFdpWorkflowProcessType: pLabels[apistructs.LabelFdpWorkflowProcessType],
// 		apistructs.LabelFdpWorkflowRuntype:     pLabels[apistructs.LabelFdpWorkflowRuntype],
// 	}

// 	var metricsKeys []string
// 	for k := range metricLabelMap {
// 		metricsKeys = append(metricsKeys, k)
// 	}
// 	// metricsKeys 顺序保证一致，promp 才不会报错
// 	sort.Strings(metricsKeys)

// 	var metricsValues []string
// 	for _, k := range metricsKeys {
// 		metricsValues = append(metricsValues, metricLabelMap[k])
// 	}

// 	return metricsKeys, metricsValues, metricLabelMap
// }

// var pipelineCounterTotal *promxp.AutoResetCounterVec
// var pipelineGaugeProcessing *prometheus.GaugeVec

// func init() {
// 	labelKeys, _, _ := generatePipelineMetricLabels(spec.Pipeline{})
// 	pipelineCounterTotal = promxp.RegisterAutoResetCounterVec(
// 		fieldPipelineTotal,
// 		"pipeline total counter",
// 		map[string]string{},
// 		labelKeys,
// 	)
// 	pipelineGaugeProcessing = promxp.RegisterGaugeVec(
// 		fieldPipelineProcessing,
// 		"processing pipeline",
// 		nil,
// 		labelKeys,
// 	)
// }

// // PipelineCounterTotalAdd 某时间段内累计执行次数、成功次数、失败次数
// var PipelineCounterTotalAdd = func(p spec.Pipeline, value float64) {
// 	if disableMetrics {
// 		return
// 	}
// 	defer func() {
// 		if r := recover(); r != nil {
// 			pipelineErrorLog(p, "[alert] failed to do metric PipelineCounterTotalAdd, err: %v", r)
// 		}
// 	}()
// 	_, labelValues, _ := generatePipelineMetricLabels(p)
// 	pipelineCounterTotal.WithLabelValues(labelValues...).Add(value)
// 	pipelineDebugLog(p, "metric: PipelineCounterTotalAdd, value: %v, labelValues: %v", value, labelValues)
// }

// // PipelineGaugeProcessingAdd 正在处理中的个数
// var PipelineGaugeProcessingAdd = func(p spec.Pipeline, value float64) {
// 	if disableMetrics {
// 		return
// 	}
// 	defer func() {
// 		if r := recover(); r != nil {
// 			pipelineErrorLog(p, "[alert] failed to do metric PipelineGaugeProcessingAdd, err: %v", r)
// 		}
// 	}()
// 	_, labelValues, _ := generatePipelineMetricLabels(p)
// 	pipelineGaugeProcessing.WithLabelValues(labelValues...).Add(value)
// 	pipelineDebugLog(p, "metric: PipelineGaugeProcessingAdd, value: %v, labelValues: %v", value, labelValues)
// }

// // PipelineEndEvent 终态时推送事件
// // 带有 org_id 或 org_name 才能在多云管理里通过租户信息查询出来
// var PipelineEndEvent = func(p spec.Pipeline) {
// 	if disableMetrics {
// 		return
// 	}
// 	defer func() {
// 		if r := recover(); r != nil {
// 			pipelineErrorLog(p, "[alert] failed to do event PipelineEndEvent, err: %v", r)
// 		}
// 	}()
// 	request := metrics.CreateBulkMetricRequest()
// 	// tags 用于索引，fields 用于存其他非索引数据
// 	_, _, metricLabels := generatePipelineMetricLabels(p)
// 	request.Add("dice_pipeline", metricLabels, generatePipelineEventFields(p))
// 	if err := bulkClient.Push(request); err != nil {
// 		pipelineErrorLog(p, "[alert] failed to push pipeline bulk event, err: %v", err)
// 	}
// 	pipelineDebugLog(p, "send pipeline end event success")
// }

// func generatePipelineEventFields(p spec.Pipeline) map[string]interface{} {
// 	return map[string]interface{}{
// 		num: 1,
// 	}
// }
