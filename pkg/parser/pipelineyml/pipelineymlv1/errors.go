package pipelineymlv1

import "github.com/pkg/errors"

var (
	errDuplicateResTypes      = errors.New("duplicate resource_type found")
	errDuplicateRes           = errors.New("duplicate resource found")
	errInvalidTypeOfResType   = errors.New("type of resource_type invalid, only support [" + DockerImageResType + "]")
	errInvalidSourceOfResType = errors.New("source of resource_type invalid")

	errLackResName = errors.New("lack of resource name")
	errLackResType = errors.New("lack of resource type")

	errUnknownResTypes = errors.New("unknown resource_types found")
	errUnusedResTypes  = errors.New("unused resource_types found")

	errUnusedResources = errors.New("unused resources found")

	errInvalidVersion = errors.New("invalid version")

	errTempTaskConfigsSize = errors.New("temporary the size of tasks list limit to 1")

	errNoClusterNameSpecify = errors.New("no clusterName specified")

	errInvalidResource = errors.New("invalid resource")

	errNotAvailableInContext = errors.New("not available in context currently")

	errDuplicateOutput = errors.New("this output already exist")

	errDuplicateTaskName = errors.New("task name already used")

	errNilPipelineYmlObj = errors.New("PipelineYml.obj is nil pointer")

	errInvalidStepTaskConfig = errors.New("invalid step task config found, type should be one of: get, put, task")

	errDecodeGetStepTask  = errors.New("error decode get step task")
	errDecodePutStepTask  = errors.New("error decode put step task")
	errDecodeTaskStepTask = errors.New("error decode task step task")

	errMissingNFSRealPath = errors.New("missing nfs real path for context store")

	errTriggerScheduleCron      = errors.New("invalid trigger schedule cron syntax")
	errTriggerScheduleFilters   = errors.New("invalid trigger schedule filter syntax")
	errTooManyLegalTriggerFound = errors.New("more than one legal triggers found!")
)
