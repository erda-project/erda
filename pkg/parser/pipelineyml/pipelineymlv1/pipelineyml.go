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

package pipelineymlv1

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/parser/pipelineyml/pipelineymlv1/pipelineymlvars"
	"github.com/erda-project/erda/pkg/parser/pipelineyml/pipelineymlv1/steptasktype"
)

type PipelineYml struct {
	// byteData represents byte format pipeline.yml
	byteData []byte
	// obj represents the struct parsed according to byteData field
	obj *Pipeline

	option   *Option
	metadata PipelineMetadata
}

// PipelineMetadata contains extra info needs by parse process.
type PipelineMetadata struct {
	pipelineID string // pipelineID
	instanceID string // pipeline instance ID
	ContextDir string // context dir related with instanceID, under nfs mount point or oss path dir

	taskNameMap map[string]struct{}

	contextMap map[string][]string

	PlaceHolderEnvs map[string]string
}

func New(b []byte) *PipelineYml {
	return &PipelineYml{
		byteData: b,
	}
}

func (y *PipelineYml) Version() string {
	return y.obj.Version
}

func (y *PipelineYml) Object() *Pipeline {
	return y.obj
}

func (y *PipelineYml) GetOptions() *Option {
	return y.option
}

func (y *PipelineYml) GetMetadata() PipelineMetadata {
	return y.metadata
}

func (y *PipelineYml) StringData() string {
	return string(y.byteData)
}

func (y *PipelineYml) GetInstanceID() string {
	return y.metadata.instanceID
}

func (y *PipelineYml) GetPlaceholders() []apistructs.MetadataField {
	return y.option.placeholders
}

func (y *PipelineYml) GetPlaceHolderTransformedEnvs() map[string]string {
	return y.metadata.PlaceHolderEnvs
}

func (y *PipelineYml) transformPlaceholdersToEnvs(placeholder []apistructs.MetadataField) {
	if y.metadata.PlaceHolderEnvs == nil {
		y.metadata.PlaceHolderEnvs = make(map[string]string)
	}
	// placeholder ????????? ????????????
	for _, ph := range placeholder {
		trimK := strings.TrimSuffix(strings.TrimPrefix(ph.Name, "(("), "))")
		trimV := strings.TrimSuffix(strings.TrimPrefix(ph.Value, "'"), "'")
		newK := strings.Replace(strings.Replace(strings.ToUpper(trimK), ".", "_", -1), "-", "_", -1)
		y.metadata.PlaceHolderEnvs[newK] = trimV
	}
}

func (y *PipelineYml) SetPipelineID(pipelineID uint64) {
	y.metadata.pipelineID = fmt.Sprintf("%d", pipelineID)
}

func (y *PipelineYml) JSON() (string, error) {
	b, err := json.Marshal(y.obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (y *PipelineYml) JSONPretty() (string, error) {
	b, err := json.MarshalIndent(y.obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (y *PipelineYml) YAML() (string, error) {
	b, err := yaml.Marshal(y.obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ReRender ????????????????????????????????????????????????????????????????????????
func (y *PipelineYml) ReRender() error {
	newYmlByte, err := y.YAML()
	if err != nil {
		return err
	}
	y.byteData = []byte(newYmlByte)
	return y.Parse()
}

func exitMultiError(errFuncs ...func() error) error {
	for _, f := range errFuncs {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

// Unmarshal unmarshal byteData to obj with evaluate, need metadata templateVars.
func (y *PipelineYml) unmarshal() error {

	err := yaml.Unmarshal(y.byteData, &y.obj, func(decoder *json.Decoder) *json.Decoder {
		return decoder
	})
	if err != nil {
		return err
	}

	if y.obj == nil {
		return errNilPipelineYmlObj
	}

	if y.option.renderPlaceholder {
		err = y.evaluate(y.option.placeholders)
		if err != nil {
			return err
		}
	}

	// re unmarshal to obj, because byteData updated by evaluate
	err = yaml.Unmarshal(y.byteData, &y.obj)
	return err
}

// Parse parse yaml file, and validate fields content.
func (y *PipelineYml) Parse(ops ...OpOption) error {
	if y.option == nil {
		y.option = &Option{}
	}
	for _, op := range ops {
		op(y.option)
	}

	// resource type
	if y.option.builtinResourceTypeDockerImagePrefix == "" {
		y.option.builtinResourceTypeDockerImagePrefix = BuiltinResourceTypeDefaultDockerImagePrefix
	}
	y.option.builtinResourceTypeDockerImagePrefix = strings.TrimSuffix(y.option.builtinResourceTypeDockerImagePrefix, "/")
	if y.option.builtinResourceTypeDockerImageTag == "" {
		y.option.builtinResourceTypeDockerImageTag = BuiltinResourceTypeDefaultDockerImageTag
	}
	// container resource limit
	if y.option.containerResLimit.cpu == 0 {
		y.option.containerResLimit.cpu = float64(0.5)
	}
	if y.option.containerResLimit.mem == 0 {
		y.option.containerResLimit.mem = float64(2048)
	}

	// if y.metadata.pipelineID == "" {
	// 	y.metadata.pipelineID = ciuuid.NewUUID()
	// }
	if y.metadata.instanceID == "" {
		// y.metadata.instanceID = fmt.Sprintf("%s-%d", y.metadata.pipelineID, y.option.rerunVer)
		y.metadata.instanceID = y.metadata.pipelineID
	}
	y.metadata.ContextDir = filepath.Join(y.option.nfsRealPath, "pipelines", y.metadata.instanceID)

	return exitMultiError(
		y.unmarshal,
		y.checkVersion,
		func() error {
			y.transformPlaceholdersToEnvs(y.option.placeholders)
			return nil
		},
		y.validateTriggers,
		y.tempTaskConfigsSize,
		y.generateTasks,
		func() error {
			if !y.option.alreadyTransformed {
				return y.InsertDiceHub()
			}
			return nil
		},
		y.generateTasks,
		y.checkTimeout,
		y.validateResources,
		y.validateContext,
	)
}

func (y *PipelineYml) ParseWithAllFields(ops ...OpOption) error {
	if err := y.Parse(ops...); err != nil {
		return err
	}
	// other required task check TODO
	if len(y.option.clusterName) == 0 {
		return errNoClusterNameSpecify
	}
	if len(y.option.nfsRealPath) == 0 {
		return errMissingNFSRealPath
	}
	return nil
}

func (y *PipelineYml) checkVersion() error {
	if y.obj.Version != "1.0" {
		return errors.Wrap(errInvalidVersion, y.obj.Version)
	}
	return nil
}

func (y *PipelineYml) tempTaskConfigsSize() error {
	for i, stage := range y.obj.Stages {
		if len(stage.TaskConfigs) > 1 {
			return errors.Wrapf(errTempTaskConfigsSize, "stage[%d]", i)
		}
	}
	return nil
}

func (y *PipelineYml) ValidateResTypes() error {
	var duplicate, invalidType, invalidSource []string
	// ?????????????????? resource type name
	nameMap := make(map[string]int, 0)
	for name := range y.builtinResTypes() {
		nameMap[name]++
	}
	for _, resType := range y.obj.ResourceTypes {
		nameMap[resType.Name]++
	}
	for name, count := range nameMap {
		if count > 1 {
			duplicate = append(duplicate, name)
		}
	}

	for _, resType := range y.obj.ResourceTypes {
		// ?????? resource_type type
		if resType.Type != DockerImageResType {
			invalidType = append(invalidType, resType.Name)
		}
		// ?????? resource_type source
		if _, ok := resType.Source["repository"]; !ok {
			invalidSource = append(invalidSource, resType.Name)
		}
	}

	var me *multierror.Error
	if len(duplicate) > 0 {
		me = multierror.Append(me, errors.Wrap(errDuplicateResTypes, fmt.Sprintf("%v", duplicate)))
	}
	if len(invalidType) > 0 {
		me = multierror.Append(me, errors.Wrap(errInvalidTypeOfResType, fmt.Sprintf("%v", invalidType)))
	}
	if len(invalidSource) > 0 {
		me = multierror.Append(me, errors.Wrap(errInvalidSourceOfResType, fmt.Sprintf("%v", invalidSource)))
	}

	// unknownResTypes := y.findUnknownResTypes()
	//
	// for _, typeName := range unknownResTypes {
	// 	_, err := dbClient.GetAction(typeName)
	// 	if err != nil {
	// 		me = multierror.Append(me, errors.Wrap(errUnknownResTypes, fmt.Sprintf("%v", unknownResTypes)))
	// 	}
	// }

	unusedResTypes := y.findUnusedUserDefinedResTypes()
	if len(unusedResTypes) > 0 {
		me = multierror.Append(me, errors.Wrap(errUnusedResTypes, fmt.Sprintf("%v", unusedResTypes)))
	}

	return me.ErrorOrNil()
}

func (y *PipelineYml) validateResources() error {

	var me *multierror.Error

	// resource ??????
	for _, res := range y.obj.Resources {
		if res.Name == "" {
			me = multierror.Append(me, errLackResName)
		}
		if res.Type == "" {
			me = multierror.Append(me, errors.Wrap(errLackResType, res.Name))
		}
	}

	// ?????????????????? resource
	var duplicate []string
	nameMap := make(map[string]int, 0)
	for _, res := range y.obj.Resources {
		nameMap[res.Name]++
	}
	for name, count := range nameMap {
		if count > 1 {
			duplicate = append(duplicate, name)
		}
	}
	if len(duplicate) > 0 {
		me = multierror.Append(me, errors.Wrap(errDuplicateRes, fmt.Sprintf("%v", duplicate)))
	}

	// ????????????????????? resource
	unusedResources := y.findUnusedResources()
	if len(unusedResources) > 0 {
		me = multierror.Append(me, errors.Wrap(errUnusedResources, fmt.Sprintf("%v", unusedResources)))
	}

	return me.ErrorOrNil()
}

func (y *PipelineYml) generateTasks() error {
	taskNameMap := make(map[string]struct{})
	for _, stage := range y.obj.Stages {
		for _, taskConfig := range stage.TaskConfigs {
			tasks, _, err := y.taskConfig2StepTasks(taskConfig)
			if err != nil {
				return err
			}
			for _, task := range tasks {
				// task name must be unique
				if _, ok := taskNameMap[task.Name()]; ok {
					return errors.Wrap(errDuplicateTaskName, task.Name())
				}
				taskNameMap[task.Name()] = struct{}{}
			}
			stage.Tasks = tasks
		}
	}
	y.metadata.taskNameMap = taskNameMap
	return nil
}

func (y *PipelineYml) taskConfig2StepTasks(taskConfig map[string]interface{}) ([]StepTask, bool, error) {
	tasks := make([]StepTask, 0)
	// aggregate
	if _, ok := taskConfig["aggregate"]; ok {
		var agg aggregateTask
		aggDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:      &agg,
			ErrorUnused: true,
		})
		if err != nil {
			return nil, true, err
		}
		err = aggDecoder.Decode(taskConfig)
		if err != nil {
			return nil, true, errors.Wrap(err, "decode aggregate")
		}
		for _, cfg := range agg.Aggregate {
			task, err := y.validateSingleTaskConfig(cfg)
			if err != nil {
				return nil, true, err
			}
			if task != nil {
				tasks = append(tasks, task)
			}
		}
		return tasks, true, nil
	} else {
		// get / put / task
		task, err := y.validateSingleTaskConfig(taskConfig)
		if err != nil {
			return nil, false, err
		}
		tasks = append(tasks, task)
		return tasks, false, nil
	}
	return tasks, false, nil
}

func (y *PipelineYml) validateSingleTaskConfig(tc TaskConfig) (StepTask, error) {

	// ?????? steptask ??????
	_, getOK := tc[pipelineymlvars.FieldGet.String()]
	_, putOK := tc[pipelineymlvars.FieldPut.String()]
	_, taskOK := tc[pipelineymlvars.FieldTask.String()]

	// steptask ????????????
	if !getOK && !putOK && !taskOK {
		return nil, errors.Wrapf(errInvalidStepTaskConfig, "%+v", tc)
	}
	// ??????????????????????????????????????? decoder.ErrorUnused ????????????
	// put ???????????????????????? get??????????????????????????????????????????????????? steptask ????????????

	switch true {
	case getOK:
		return tc.decodeStepTaskWithValidate(steptasktype.GET, y)
	case putOK:
		return tc.decodeStepTaskWithValidate(steptasktype.PUT, y)
	case taskOK:
		return tc.decodeStepTaskWithValidate(steptasktype.TASK, y)
	}

	return nil, errors.Errorf("invalid task config, content: %v, type: %v\n", tc, reflect.TypeOf(tc))
}

// validResTypesMap return all valid resource type:
// 1. built-in
// 2. found in resource_types
func (y *PipelineYml) validResTypesMap() map[string]ResourceType {
	result := make(map[string]ResourceType)
	for name, resType := range y.builtinResTypes() {
		result[name] = resType
	}
	for _, resType := range y.obj.ResourceTypes {
		if _, ok := result[resType.Name]; !ok {
			result[resType.Name] = resType
		}
	}

	return result
}

// findUnknownResTypes find unknown resource types:
// 1. not built-in
// 2. not found in resource_types
func (y *PipelineYml) findUnknownResTypes() []string {

	m := y.validResTypesMap()

	unknownTypes := make([]string, 0)

	for _, res := range y.obj.Resources {
		if _, ok := m[res.Type]; !ok {
			unknownTypes = append(unknownTypes, res.Type)
		}
	}

	return unknownTypes
}

func (y *PipelineYml) validResourceMap() map[string]Resource {
	result := make(map[string]Resource)

	for _, res := range y.obj.Resources {
		if _, ok := result[res.Name]; !ok {
			result[res.Name] = res
		}
	}
	return result
}

func (y *PipelineYml) findUnusedResources() []string {
	unused := make([]string, 0)
	m := y.validResourceMap()

	for _, stage := range y.obj.Stages {
		for _, task := range stage.Tasks {
			for _, resName := range task.RequiredContext(*y) {
				if _, ok := m[resName]; ok {
					delete(m, resName)
				}
			}
			for _, resName := range task.OutputToContext() {
				if _, ok := m[resName]; ok {
					delete(m, resName)
				}
			}
		}
	}
	for resName := range m {
		unused = append(unused, resName)
	}

	return unused
}

func (y *PipelineYml) findUnusedUserDefinedResTypes() []string {
	unused := make([]string, 0)
	m := y.validResTypesMap()
	for _, res := range y.obj.Resources {
		delete(m, res.Type)
	}
	builtinResTypes := y.builtinResTypes()
	for resTypeName := range m {
		if _, ok := builtinResTypes[resTypeName]; !ok {
			unused = append(unused, resTypeName)
		}
	}
	return unused
}

func (y *PipelineYml) validateContext() error {
	y.metadata.contextMap = make(map[string][]string)
	currentContext := make(map[string]struct{})
	for _, stage := range y.obj.Stages { // TODO ?????? ??????????????? stage????????? task ??????????????????????????????????????????????????? out resource
		// ??????????????????

		var addToNextStage = make([]string, 0)
		for _, task := range stage.Tasks {
			if task.IsDisable() {
				continue
			}
			taskContext := shadowClone(currentContext)
			// ??????????????? context ??? ??????????????? ???????????????????????????
			for _, require := range task.RequiredContext(*y) {
				if _, ok := taskContext[require]; !ok {
					return errors.Wrapf(errNotAvailableInContext, "task: %s, require: %s", task.Name(), require)
				}
			}
			// ??? output ????????? context??????????????? task ??????
			for _, out := range task.OutputToContext() {
				_, ok := taskContext[out]
				if ok {
					return errors.Wrap(errDuplicateOutput, out)
				}
				taskContext[out] = struct{}{}
				addToNextStage = append(addToNextStage, out)
			}
			// ????????? task ???????????????
			for resName := range taskContext {
				y.metadata.contextMap[task.Name()] = append(y.metadata.contextMap[task.Name()], resName)
			}
		}

		for _, add := range addToNextStage {
			currentContext[add] = struct{}{}
		}
	}
	return nil
}

func shadowClone(src map[string]struct{}) map[string]struct{} {
	dst := make(map[string]struct{})
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
