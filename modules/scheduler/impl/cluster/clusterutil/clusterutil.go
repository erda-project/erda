// Package clusterutil cluster utils
package clusterutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	// ServiceKindMarathon marathon executor type
	ServiceKindMarathon = "MARATHON"
	// ServiceKindK8S k8s executor type
	ServiceKindK8S = "K8S"
	// ServiceKindEdas edas executor type
	ServiceKindEdas = "EDAS"
	// JobKindMetronome metronome executor type
	JobKindMetronome = "METRONOME"
	// JobKindK8S k8s job executor type
	JobKindK8S = "K8SJOB"
	// EdasKindInK8s k8s service executor in edas
	EdasKindInK8s = "EDASINK8S"
	// DiceCluster DICE_CLUSTER env key
	DiceCluster = "DICE_CLUSTER"
	// k8sKindFlink 标识 k8s flink 类型
	K8SKindFlink = "K8SFLINK"
	// k8sKindSpark 标识 k8s spark 类型
	K8SKindSpark = "K8SSpark"
	// k8sSparkVersion 标识 k8s spark 类型
	K8SSparkVersion = "2.4.0"
)

// SetRuntimeExecutorByCluster runtime 的 executor 为空时，根据 cluster 值设置 executor
func SetRuntimeExecutorByCluster(runtime *apistructs.ServiceGroup) error {
	if len(runtime.Executor) > 0 {
		return nil
	}
	if runtime.ClusterName == "" {
		return errors.Errorf("runtime(%s/%s) neither executor nor cluster is set", runtime.Type, runtime.ID)
	}

	runtime.Executor = GenerateExecutorByCluster(runtime.ClusterName, ServiceKindMarathon)

	logrus.Infof("generate executor(%s) for runtime(%s) in cluster(%s)", runtime.Executor, runtime.ID, runtime.ClusterName)
	return nil
}

// SetJobExecutorByCluster job 的 executor 为空时，根据 cluster 值设置 executor
func SetJobExecutorByCluster(job *apistructs.Job) error {
	if len(job.Executor) > 0 {
		return nil
	}
	if job.ClusterName == "" {
		return errors.Errorf("job(%s/%s) neither executor nor cluster is set", job.Namespace, job.Name)
	}

	if job.Kind == "" { // FIXME 兼容老的未传Kind的情况，后续可去除
		job.Kind = JobKindMetronome
	}

	job.Executor = GenerateExecutorByCluster(job.ClusterName, strutil.ToUpper(job.Kind))

	logrus.Infof("generate executor(%s) for job(%s) in cluster(%s)", job.Executor, job.Name, job.ClusterName)
	return nil
}

func SetJobVolumeExecutorByCluster(jobvolume *apistructs.JobVolume) error {
	if len(jobvolume.Executor) > 0 {
		return nil
	}
	if jobvolume.ClusterName == "" {
		return errors.Errorf("jobvolume(%s/%s) neither executor nor cluster is set",
			jobvolume.Namespace, jobvolume.Name)
	}
	if jobvolume.Kind == "" {
		jobvolume.Kind = JobKindMetronome
	}
	jobvolume.Executor = GenerateExecutorByCluster(jobvolume.ClusterName, strutil.ToUpper(jobvolume.Kind))
	return nil
}

var (
	preFetcher, _ = jsonstore.New()
)

// GenerateExecutorByCluster 根据 `cluster` 和 `executorType` 得到 executorname
func GenerateExecutorByCluster(cluster, executorType string) string {
	cname := cluster
	switch cname {
	case "xhsd-t2":
		cname = "XINHUATEST"
	case "fsg-prod":
		cname = "WAIFU"
	case "xhsd-prod":
		cname = "XINHUAPROD"
	}

	var executor string
	if err := preFetcher.Get(context.Background(),
		fmt.Sprintf("/dice/clustertoexecutor/%s/%s", cluster, executorType), &executor); err == nil {
		return executor
	}
	// 处理 edas 的 service executor
	if executorType == ServiceKindEdas {
		return "EDASFOR" + strutil.ToUpper(strings.Replace(cname, "-", "", -1))
	}
	// terminus-y -> terminusy
	c := strings.Replace(cname, "-", "", -1)
	//e.g. MARATHONFORTERIMINUSY, METRONOMEFORTERMINUSY
	return executorType + "FOR" + strutil.ToUpper(c)
}

func GenerateExecutorByClusterName(clustername string) string {
	return GenerateExecutorByCluster(clustername, ServiceKindMarathon)
}
