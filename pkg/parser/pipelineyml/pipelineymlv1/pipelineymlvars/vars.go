package pipelineymlvars

// envs
const (
	CLUSTER_NAME = "CLUSTER_NAME"
)

type YmlField string

func (f YmlField) String() string {
	return string(f)
}

const (
	FieldAggregate           YmlField = "aggregate"
	FieldGet                 YmlField = "get"
	FieldPut                 YmlField = "put"
	FieldTask                YmlField = "task"
	FieldDisable             YmlField = "disable"
	FieldPause               YmlField = "pause"
	FieldEnvs                YmlField = "envs"
	FieldParams              YmlField = "params"
	FieldParamForceBuildpack YmlField = "force_buildpack"
	FieldTaskConfig          YmlField = "config"
	FieldTaskConfigEnvs      YmlField = "envs"
)
