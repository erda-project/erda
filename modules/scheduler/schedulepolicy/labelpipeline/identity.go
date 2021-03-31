package labelpipeline

import (
	"sort"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/strutil"
)

// 身份标签
func IdentityFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	jobLabelFilter(r, r2, li)
	packLabelFilter(r, r2, li)
	daemonsetLabelFilter(r, r2, li)
	statefulLabelFilter(r, r2, li)
	statelessLabelFilter(r, r2, li)
	bigdataLabelFilter(r, r2, li)
	platformLabelFilter(r, r2, li)
	lockLabelFilter(r, r2, li)
	projectLabelFilter(r, r2, li)
	anyLabelFilter(r, r2, li)
}

// job
func jobLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	if li.ExecutorKind == labelconfig.EXECUTOR_METRONOME || li.ExecutorKind == labelconfig.EXECUTOR_K8SJOB {
		if kind, ok := li.Label[apistructs.LabelJobKind]; !ok || kind != apistructs.TagBigdata {
			r.Likes = append(r.Likes, apistructs.TagJob)
			r2.Job = true
		}
	}
}

// pack
// 当前几乎未使用该标签
func packLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	if li.ExecutorKind == labelconfig.EXECUTOR_METRONOME || li.ExecutorKind == labelconfig.EXECUTOR_K8SJOB {
		if li.Label[apistructs.LabelPack] == "true" {
			// marathon&metronome 不用 pack 标, k8s 会用, 所以 r 中不设置, r2 中设置
			// r.Likes = append(r.Likes, apistructs.TagPack)
			r2.Pack = true
		}
	}
}

// 在 dcos 上不支持 daemonset 部署
func daemonsetLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	if li.Label["SERVICE_TYPE"] == "DAEMONSET" {
		r2.IsDaemonset = true
	}
}

// service-stateful
func statefulLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	if li.Label["SERVICE_TYPE"] == "ADDONS" {
		r.Likes = append(r.Likes, apistructs.TagServiceStateful)
		r2.Stateful = true
	}
}

// service-stateless
func statelessLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	if li.Label["SERVICE_TYPE"] == "STATELESS" {
		r.Likes = append(r.Likes, apistructs.TagServiceStateless)
		r2.Stateless = true
	}
}

// bigdata
func bigdataLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	// 当前 bigdata 只有对应的 job，没有 runtime
	if li.ExecutorKind == labelconfig.EXECUTOR_METRONOME ||
		li.ExecutorKind == labelconfig.EXECUTOR_K8SJOB ||
		li.ExecutorKind == labelconfig.EXECUTOR_SPARK ||
		li.ExecutorKind == labelconfig.EXECUTOR_K8SSPARK ||
		li.ExecutorKind == labelconfig.EXECUTOR_FLINK {
		if kind, ok := li.Label[apistructs.LabelJobKind]; ok && kind == apistructs.TagBigdata {
			r.ExclusiveLikes = append(r.ExclusiveLikes, apistructs.TagBigdata)
			r2.BigData = true
		}
	}
	// TODO: stand for bigdata service
}

// platform
func platformLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	// TODO: 完全消除 v1 API 后，删除 li.Label
	if _, ok := li.Label[labelconfig.PLATFORM]; ok {
		r.IsPlatform = true
		r2.IsPlatform = true
		return
	}

	var anyfalse bool
	for _, selectors := range li.Selectors {
		platformSelector := selectors["platform"]
		if platformSelector.Not ||
			len(platformSelector.Values) == 0 ||
			strutil.ToLower(strutil.Trim(platformSelector.Values[0])) != "true" {
			anyfalse = true
			break
		}
	}
	if !anyfalse && len(li.Selectors) > 0 {
		r.IsPlatform = true
		r2.IsPlatform = true
		return
	}
	r.IsPlatform = false
	r2.IsPlatform = false
}

// locked
func lockLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	r.IsUnLocked = true
	r2.IsUnLocked = true
	// r.UnLikes = append(r.UnLikes, apistructs.TagLocked)
}

// any
func anyLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	// 一般而言 "platform" 和 "locked" 标签不会被打到服务或者job上
	sort.Strings(r.ExclusiveLikes)
	idx := sort.SearchStrings(r.ExclusiveLikes, apistructs.TagBigdata)
	exist := idx < len(r.ExclusiveLikes) && r.ExclusiveLikes[idx] == apistructs.TagBigdata
	if exist || elemPrefixInArray(apistructs.TagProjectPrefix, r.Likes) {
		return
	}
	// 对 any 标签特殊处理
	//r.Likes = append(r.Likes, spec.TagAny)
	r.Flag = true
	r2.PreferJob = true
	r2.PreferPack = true
	r2.PreferStateful = true
	r2.PreferStateless = true
}

// project
// this label is to be deprecated soon
func projectLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	if projectId, ok := li.Label["DICE_PROJECT"]; ok {
		if exists := li.ExecutorConfig.ProjectIDForCompatibility(projectId); exists {
			r.ExclusiveLikes = append(r.ExclusiveLikes, apistructs.TagProjectPrefix+projectId)
			r2.HasProject = true
			r2.Project = projectId
			return
		}
	}
	r.UnLikePrefixs = append(r.UnLikePrefixs, apistructs.TagProjectPrefix)
	r2.HasProject = false
}

func elemPrefixInArray(prefix string, Array []string) bool {
	for _, v := range Array {
		if strings.HasPrefix(v, prefix) {
			return true
		}
	}
	return false
}
