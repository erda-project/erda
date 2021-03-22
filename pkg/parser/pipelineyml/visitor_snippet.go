package pipelineyml

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/flosch/pongo2.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/expression"
	"github.com/erda-project/erda/pkg/pipeline_snippet_client"
)

const MaxDeep = 10

const SnippetActionNameLinkAddr = "_"

const (
	Snippet                = "snippet"
	SnippetLogo            = "http://terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/10/22/410935c6-e399-463a-b87b-0b774240d12e.png"
	SnippetDisplayName     = "嵌套流水线"
	SnippetDesc            = "嵌套流水线可以声明嵌套的其他 pipeline.yml"
	DicePipelinesGitFolder = ".dice/pipelines"
)

type SnippetVisitor struct {
	globalSnippetConfigLabels map[string]string
	caches                    []SnippetPipelineYmlCache
}

func NewSnippetVisitor(globalSnippetConfigLabels map[string]string, caches []SnippetPipelineYmlCache) *SnippetVisitor {
	v := SnippetVisitor{}
	v.globalSnippetConfigLabels = globalSnippetConfigLabels
	v.caches = caches
	return &v
}

func (v *SnippetVisitor) Visit(s *Spec) {
	if _, err := v.loadPipelineSnippetToAction(0, s, nil, nil, v.caches); err != nil {
		s.appendError(err)
	}
}

// 有历史原因 原本叫 template 然后变成了 snippet, 其中有些名称还没改过来

// 整体就是将模板当做action看待，模板的中的action都递归设置到一个纬度上，都设置在深度0的stage的action数组中，然后设置不同深度的action他的needs不同，后续就交给dag调度了
// 这是一个递归函数 bdl 会层层传递，其他输入参数和输出参数每层都不同

// myDeep 代表当前的递归深度，0深度会进行一些设值的特殊处理，深度目前大于10会报错，因为里面有请求，深度太深那么调用就太慢了
// needActions 陷入模板中的时候，需要将之前的allTemplateActions塞入，这样模板中的action need才会设置正确
// outputs pipeline中所有模板的output集合，后续会替换成action的output
// pipelineYamlCaches 缓存

// allTemplateActions 返回的action数组，整个数组就是pipeline的所有action, 如果pipeline中有模板，那么模板里面的action也会被包括
// error 方法调用产生的error
func (v *SnippetVisitor) loadPipelineSnippetToAction(myDeep int, s *Spec, needActions []*Action, outputs *[]apistructs.SnippetFormatOutputs, pipelineYamlCaches []SnippetPipelineYmlCache) (allTemplateActions []*Action, returnError error) {

	if s == nil {
		return nil, errors.New(" loadPipelineSnippetToAction error: pipelineYaml spec empty ")
	}

	if s.Stages == nil {
		return nil, errors.New(" loadPipelineSnippetToAction error: pipelineYaml spec.Stages empty ")
	}

	// 判定深度是否大于9，大于9下面+1深度就大于10了
	if myDeep > MaxDeep {
		return nil, fmt.Errorf("getSnippetPipelineYaml error: snippet deep > %v，Please verify whether there is a circular dependency", MaxDeep)
	}

	// 传递给下一个递归的深度
	deep := myDeep + 1

	if outputs == nil {
		outputs = &[]apistructs.SnippetFormatOutputs{}
	}

	// 遍历步骤节点
	for si := range s.Stages {
		stage := s.Stages[si]

		// 每个节点的所有action
		var stageActions []*Action

		// 遍历action
		for typekey := range stage.Actions {
			typedAction := stage.Actions[typekey]

			for actionType := range typedAction {
				action := typedAction[actionType]

				// 不是模板就将当前action设置上
				if !IsSnippetType(actionType) {
					stageActions = append(stageActions, action)

					// 这里判定allTemplateActions假如不为空，就代表之前的stage的action需要设置到当前action的need中
					var allAlias []ActionAlias
					if allTemplateActions != nil {
						for _, v := range allTemplateActions {
							allAlias = append(allAlias, v.Alias)
						}
					}
					// needActions不为空代表，当前应该处于模板的stage当中，陷入模板递归之前的allTemplateActions就是needActions, 当前action就需要加入陷入模板之`前的allTemplateActions
					if needActions != nil {
						for _, v := range needActions {
							allAlias = append(allAlias, v.Alias)
						}
					}
					typedAction[actionType].Needs = allAlias
					continue
				}

				// ---- 模板逻辑开始

				// 这里判定别名不能为空。模板需要限制，其必须要有一个别名，不能使用snippet/dice/templateName
				if action.Alias == "" {
					return nil, errors.New("loadPipelineSnippetToAction error: snippet action fields alias is Required")
				}
				if checkErr := CheckActionSnippetConfig(action.SnippetConfig); checkErr != nil {
					return nil, checkErr
				}
				// 根据模板的action模式，获取到其具体的pipelineYaml
				pipelineYaml, outs, err := GetSnippetPipelineYaml(action, v.globalSnippetConfigLabels, pipelineYamlCaches)
				if err != nil {
					return nil, errors.New(fmt.Sprintf(" getSnippetAllAction error: %v ", err))
				}
				if outs != nil {
					*outputs = append(*outputs, outs...)
				}

				// 递归调用，获取里面全部的action
				templateActionAllAction, err := v.loadPipelineSnippetToAction(deep, pipelineYaml.Spec(), allTemplateActions, outputs, pipelineYamlCaches)
				if err != nil {
					return nil, err
				}

				// 将全部的action设置到当前stage的action数组中
				stageActions = append(stageActions, templateActionAllAction...)
			}
		}

		// 假如深度为0，这时候代表递归已经完成，应该将全部的action设置到当前stage的actions中
		if myDeep == 0 {
			maps, err := actionToMap(stageActions)
			if err != nil {
				return nil, errors.New(fmt.Sprintf(" actionToMap error:  %v ", err))
			}
			stage.Actions = maps
		} else {
			// 深度大于0，代表还是在模板的递归中，这时候需要将所有的步骤的action都设置到返回值action中
			allTemplateActions = append(allTemplateActions, stageActions...)
		}
	}

	// 全部stage遍历完成
	if myDeep == 0 {
		//遍历没层，给下一层的need设置上一层的所有action
		if err := setNeedAction(s); err != nil {
			return nil, err
		}

		setNameSpace(s)

		// 根据所有的模板的 output，替换整个p ipeline 的 output 占位符, 将 snippet 的 outputs 替换成对应的 action 的 output
		// ${snippetName:OUTPUT:xxx} -> ${action:OUTPUT:xxx}
		if outputs != nil {
			if err := replaceSpecOutputs(s, *outputs); err != nil {
				return nil, err
			}
		}

		setAllAction(s)

		return nil, nil
	} else {
		// 深度大于0，代表在模板的递归中，应该依次返回其所有的action
		return allTemplateActions, nil
	}
}

// 给 spec 重新设置好 allActions
func setAllAction(s *Spec) {
	s.allActions = make(map[ActionAlias]*indexedAction)
	for si := range s.Stages {
		stage := s.Stages[si]
		for typekey := range stage.Actions {
			typedAction := stage.Actions[typekey]
			for _, action := range typedAction {
				s.allActions[action.Alias] = &indexedAction{action, si}
			}
		}
	}
}

// 给 action 重新设置好 Namespaces 和 NeedNamespaces
func setNameSpace(s *Spec) {
	for si := range s.Stages {
		stage := s.Stages[si]
		for typekey := range stage.Actions {
			typedAction := stage.Actions[typekey]
			for actionType := range typedAction {
				stage.Actions[typekey][actionType].NeedNamespaces = needsToNeedNameSpace(typedAction[actionType].Needs)
				stage.Actions[typekey][actionType].Namespaces = []string{typedAction[actionType].Alias.String()}
			}
		}
	}
}

func needsToNeedNameSpace(needs []ActionAlias) []string {
	var needNameSpace []string
	for _, v := range needs {
		needNameSpace = append(needNameSpace, v.String())
	}
	return needNameSpace
}

// 将全部的 OUTPUT 都替换成带有 snippet 的形式
func replaceSpecOutputs(s *Spec, outputs []apistructs.SnippetFormatOutputs) error {
	if outputs == nil || len(outputs) <= 0 {
		return nil
	}
	// 获取打平了结构的 pipeline, 里面没有 snippet 的说法了
	spec, err := json.Marshal(s)
	if err != nil {
		return errors.New(fmt.Sprintf(" replaceSpecOutputs fail, json Marshal spec %v error %v ", s, err))
	}
	specStr := string(spec)
	// 塞入 map 中，方便拿
	afterMap := make(map[string]*apistructs.SnippetFormatOutputs)
	for _, v := range outputs {
		output := v
		afterMap[v.PreOutputName] = &output
	}

	// 用作校验 snippet 中的 output 是否都存在
	for _, output := range outputs {
		// 循环层级，去拿到对应的 afterOutput
		afterOutput := output
		for i := 0; i < MaxDeep; i++ {
			if afterMap[afterOutput.PreOutputName] != nil {
				if afterMap[afterOutput.AfterOutputName] == nil {
					break
				}
				afterOutput = *afterMap[afterOutput.AfterOutputName]
			} else {
				break
			}
		}
		// 替换对应的 PreOutputName 成 afterOutput
		specStr = strings.ReplaceAll(specStr, output.PreOutputName, afterOutput.AfterOutputName)
	}

	err = json.Unmarshal([]byte(specStr), s)
	if err != nil {
		return errors.New(fmt.Sprintf(" replaceSpecOutputs fail, json Unmarshal specStr %s error %v ", specStr, err))
	}

	return nil
}

//遍历每层stage，给下一个stage设置上一个stage的所有action
func setNeedAction(s *Spec) error {
	//给每个stage的action都设置上上一个stage中的全部action
	availableActions := make(map[ActionAlias]struct{})

	for si := range s.Stages {
		stage := s.Stages[si]

		stageActions := make(map[ActionAlias]struct{})

		for typekey := range stage.Actions {
			typedAction := stage.Actions[typekey]

			for actionType := range typedAction {
				// update stageActions
				stageActions[typedAction[actionType].Alias] = struct{}{}

				typedAction[actionType].Needs = append(typedAction[actionType].Needs, toList(availableActions)...)

				//设置allAction
				s.allActions[typedAction[actionType].Alias] = &indexedAction{typedAction[actionType], si}
			}
		}

		for action := range stageActions {
			availableActions[action] = struct{}{}
		}
	}
	return nil
}

func actionToMap(allTemplateActions []*Action) ([]typedActionMap, error) {
	if allTemplateActions == nil {
		return nil, errors.New(" actionToMap error: allTemplateActions empty ")
	}

	var allTemplateActionMaps []typedActionMap
	for _, v := range allTemplateActions {
		typedActionMap := typedActionMap{}
		typedActionMap[ActionType(v.Type.String())] = v
		allTemplateActionMaps = append(allTemplateActionMaps, typedActionMap)
	}

	return allTemplateActionMaps, nil
}

func CheckActionSnippetConfig(snippetConfig *SnippetConfig) error {
	if snippetConfig.Name == "" || snippetConfig.Source == "" {
		return errors.New("snippetConfig name or source is empty")
	}
	return nil
}

func GetSnippetPipelineYaml(action *Action, globalSnippetConfigLabels map[string]string, pipelineYamlCaches []SnippetPipelineYmlCache) (*PipelineYml, []apistructs.SnippetFormatOutputs, error) {
	if action == nil {
		return nil, nil, errors.New(" getSnippetPipelineYaml error: action empty ")
	}
	if action.SnippetConfig == nil {
		return nil, nil, errors.New(" getSnippetPipelineYaml error: action snippet_config empty ")
	}

	snippetConfig := HandleSnippetConfigLabel(action.SnippetConfig, globalSnippetConfigLabels)

	// 从缓存获取 snippet 的 pipeline yml 的 graph
	var graph = GetCacheYaml(pipelineYamlCaches, action.SnippetConfig)
	if graph == nil {
		var err error
		graph, err = SnippetToPipelineGraph(snippetConfig)
		if err != nil {
			return nil, nil, err
		}
	}

	// 渲染 snippet yml 并获取 pipeline yml 描述
	spec := convertToPipelineSnippetSpec(graph, snippetConfig.Name)
	yml, renderOutputs, err := DoRenderTemplateNotReplaceParamsValue(action.Params, spec, action.Alias.String(), apistructs.TemplateVersionV2)
	if err != nil {
		return nil, nil, apierrors.ErrCreatePipelineTask.InternalError(err)
	}
	pipelineYml, err := New([]byte(yml))
	if err != nil {
		return nil, nil, apierrors.ErrParsePipelineYml.InternalError(err)
	}

	// runParams 这里 json 化后再 replace, 要不然 yml 类型会不匹配
	specStr, err := json.Marshal(pipelineYml.s)
	if err != nil {
		return nil, nil, apierrors.ErrParsePipelineYml.InternalError(err)
	}
	result := ReplacePipelineParams(string(specStr), action.Params)
	var specObject Spec
	err = json.Unmarshal([]byte(result), &specObject)
	if err != nil {
		return nil, nil, apierrors.ErrParsePipelineYml.InternalError(err)
	}
	pipelineYml.s = &specObject

	// 给 snippet pipeline 设置上 action 上的 if
	pipelineYml.s = mergeSnippetActionIf(pipelineYml.s, action)
	return pipelineYml, renderOutputs, nil
}

// 给 snippet 的 action 拼接上 snippet 的 if 条件
func mergeSnippetActionIf(s *Spec, snippet *Action) *Spec {
	if s == nil {
		return nil
	}

	if snippet.If == "" {
		return s
	}

	snippetCondition := expression.ReplacePlaceholder(snippet.If)
	if snippetCondition == "" {
		return s
	}

	s.LoopStagesActions(func(stage int, action *Action) {

		if action.If == "" {
			action.If = snippet.If
			return
		}

		actionCondition := expression.ReplacePlaceholder(action.If)
		actionCondition = fmt.Sprintf("%s && %s", snippetCondition, actionCondition)
		action.If = expression.AppendPlaceholder(actionCondition)
	})

	return s
}

//func CheckAllParamsAreReplace(yml string, alias string) (err error) {
//	// 校验是否存在没有替换的 params
//	r, _ := regexp.Compile("\\$\\{([^}]*)\\}")
//	params := r.FindStringSubmatch(yml)
//	if params != nil {
//		for index, param := range params {
//			if index == 0 {
//				continue
//			}
//			if strings.HasPrefix(param, SnippetParamsPrefix) {
//				return fmt.Errorf("snippet %s params %s not input value or this snippet yml not definition this params", alias, param)
//			}
//		}
//	}
//
//	return nil
//}

func GetCacheYaml(pipelineYamlCaches []SnippetPipelineYmlCache, snippetConfig *SnippetConfig) *apistructs.PipelineYml {

	if pipelineYamlCaches == nil {
		return nil
	}
	if snippetConfig == nil {
		return nil
	}

	for _, cache := range pipelineYamlCaches {
		if ComparedSnippet(&cache.SnippetConfig, snippetConfig) {
			return cache.PipelineYaml
		}
	}

	return nil
}

func ComparedSnippet(left *SnippetConfig, right *SnippetConfig) bool {
	// 判定 snippetConfig是否相等
	if left.Name != right.Name || left.Source != right.Source {
		return false
	}
	// name 和 souce 都相等，而且 labels都是空，就代表相等
	if left.Labels == nil && right.Labels == nil {
		return true
	}

	var allSame = true
	// 判断labels是否相等
	for k, v := range left.Labels {

		// 假如找到相等的
		var find = false
		for snippetConfigLabelsKey, snippetConfigLabelsValue := range right.Labels {
			if k == snippetConfigLabelsKey && v == snippetConfigLabelsValue {
				find = true
				break
			}
		}

		// 不相等就代表labels不同应该
		if !find {
			allSame = false
			break
		}
	}

	return allSame
}

func SnippetToPipelineGraph(snippetConfig SnippetConfig) (*apistructs.PipelineYml, error) {

	pipelineYmlContext, err := pipeline_snippet_client.GetSnippetPipelineYml(apistructs.SnippetConfig{
		Name:   snippetConfig.Name,
		Source: snippetConfig.Source,
		Labels: snippetConfig.Labels,
	})

	if err != nil {
		return nil, err
	}

	graph, err := ConvertToGraphPipelineYml([]byte(pipelineYmlContext))
	if err != nil {
		return nil, err
	}

	return graph, nil
}

func convertToPipelineSnippetSpec(yml *apistructs.PipelineYml, name string) *apistructs.PipelineTemplateSpec {
	spec := apistructs.PipelineTemplateSpec{}
	spec.Params = yml.Params
	spec.Outputs = yml.Outputs
	spec.Name = name
	spec.Version = ""
	spec.DefaultVersion = ""
	spec.Template = yml.YmlContent
	return &spec
}

func IsSnippetType(actionType ActionType) bool {
	actionName := actionType.String()
	if strings.HasPrefix(actionName, Snippet) {
		return true
	}
	return false
}

func HandleSnippetConfigLabel(snippetConfig *SnippetConfig, globalSnippetConfigLabels map[string]string) SnippetConfig {

	if snippetConfig.Labels == nil {
		snippetConfig.Labels = map[string]string{}
	}

	if globalSnippetConfigLabels != nil {
		for k, v := range globalSnippetConfigLabels {
			snippetConfig.Labels[k] = v
		}
	}

	// dice
	scopeID := snippetConfig.Labels[apistructs.LabelDiceSnippetScopeID]
	if scopeID == "" {
		scopeID = "0"
	}

	version := snippetConfig.Labels[apistructs.LabelChooseSnippetVersion]
	if version == "" {
		version = "latest"
	}

	return *snippetConfig
}

func DoRenderTemplateHandler(params map[string]interface{}, templateAction *apistructs.PipelineTemplateSpec,
	alias string, templateVersion apistructs.TemplateVersion,
	handler func(snippet string, params map[string]interface{}) (string, error)) (string, []apistructs.SnippetFormatOutputs, error) {

	if err := setDefaultValue(templateAction, params); err != nil {
		return "", nil, err
	}

	if err := checkParams(templateAction, params); err != nil {
		return "", nil, err
	}

	snippet, err := handler(templateAction.Template, params)
	if err != nil {
		return "", nil, err
	}
	templateAction.Template = snippet

	if templateVersion == apistructs.TemplateVersionV2 {
		pipelineYml, err := New([]byte(templateAction.Template), WithFlatParams(true))
		if err != nil {
			return "", nil, errors.New(fmt.Sprintf("doRenderTemplate new template error: error %v", err))
		}

		pipelineYmlStr, aliasMap, err := replaceActionName(pipelineYml, alias)
		if err != nil {
			return "", nil, err
		}

		template := replaceOutput(pipelineYmlStr, aliasMap)

		outputs, err := getTemplateOutputs(templateAction, alias)
		if err != nil {
			return "", nil, err
		}

		return template, outputs, nil
	}

	return templateAction.Template, nil, nil
}

func DoRenderTemplateNotReplaceParamsValue(params map[string]interface{}, templateAction *apistructs.PipelineTemplateSpec, alias string, templateVersion apistructs.TemplateVersion) (string, []apistructs.SnippetFormatOutputs, error) {
	return DoRenderTemplateHandler(params, templateAction, alias, templateVersion, func(snippet string, params map[string]interface{}) (string, error) {
		return snippet, nil
	})
}

func DoRenderTemplateWithFormat(params map[string]interface{}, templateAction *apistructs.PipelineTemplateSpec, alias string, templateVersion apistructs.TemplateVersion) (string, []apistructs.SnippetFormatOutputs, error) {
	return DoRenderTemplateHandler(params, templateAction, alias, templateVersion, doFormatAndReplaceValue)
}

func getTemplateOutputs(action *apistructs.PipelineTemplateSpec, alias string) ([]apistructs.SnippetFormatOutputs, error) {

	outputs := action.Outputs
	var result []apistructs.SnippetFormatOutputs
	for _, v := range outputs {
		if v.Ref == "" {
			return nil, errors.New(fmt.Sprintf(" error to format snippet output, output %s ref is empty", v.Name))
		}

		// ${xxx:OUTPUT:xxx} -> ${xxx_xxx:OUTPUT:xxx}
		matchValues := expression.OldRe.FindStringSubmatch(v.Ref)
		if len(matchValues) > 1 {
			afterOutputMame := alias + SnippetActionNameLinkAddr + matchValues[1]
			formatOutput := apistructs.SnippetFormatOutputs{
				PreOutputName:   fmt.Sprintf("%s%s:%s:%s%s", expression.OldLeftPlaceholder, alias, RefOpOutput, v.Name, expression.OldRightPlaceholder),
				AfterOutputName: fmt.Sprintf("%s%s%s", expression.OldLeftPlaceholder, afterOutputMame, expression.OldRightPlaceholder),
			}
			result = append(result, formatOutput)
		}

		// ${{ outputs.xxx.key }} -> ${{ outputs.xxx_xxx.key }}
		matchValues = expression.Re.FindStringSubmatch(v.Ref)
		if len(matchValues) > 1 {
			splitNames := strings.Split(matchValues[1], ".")
			if splitNames == nil || len(splitNames) != 3 || strings.HasPrefix(splitNames[0], expression.LeftPlaceholder) {
				continue
			}

			if splitNames[0] != expression.Outputs {
				return nil, fmt.Errorf("can not parsing this output ref %s", v.Ref)
			}

			newFormatOutput := apistructs.SnippetFormatOutputs{
				PreOutputName: fmt.Sprintf("%s %s.%s.%s %s", expression.LeftPlaceholder, expression.Outputs, alias, v.Name, expression.RightPlaceholder),
				AfterOutputName: fmt.Sprintf("%s %s.%s.%s %s", expression.LeftPlaceholder, expression.Outputs,
					alias+SnippetActionNameLinkAddr+splitNames[1], splitNames[2], expression.RightPlaceholder),
			}
			result = append(result, newFormatOutput)
		}
	}
	return result, nil
}

func replaceActionName(pipelineYml *PipelineYml, alias string) (string, map[string]string, error) {
	aliasMap := make(map[string]string)
	for si := range pipelineYml.Spec().Stages {
		stage := pipelineYml.Spec().Stages[si]

		for typekey := range stage.Actions {
			typedAction := stage.Actions[typekey]

			for actionType := range typedAction {
				action := typedAction[actionType]

				if action.Alias == "" {
					action.Alias = ActionAlias(actionType.String())
				}

				newAlias := alias + SnippetActionNameLinkAddr + action.Alias.String()
				aliasMap[action.Alias.String()] = newAlias
				action.Alias = ActionAlias(newAlias)
			}
		}
	}

	pipelineYaml, err := GenerateYml(pipelineYml.Spec())
	if err != nil {
		return "", nil, err
	}

	return string(pipelineYaml), aliasMap, nil
}

func replaceOutput(pipelineYamlStr string, aliasMap map[string]string) string {
	// template-output = template
	for aliasBefore, aliasAfter := range aliasMap {
		// ${xxx} ${xxx:output -> ${xxx_xxx} ${xxx_xxx:output
		pipelineYamlStr = strings.ReplaceAll(pipelineYamlStr, fmt.Sprintf("%s%s%s", expression.OldLeftPlaceholder, aliasBefore, expression.OldRightPlaceholder),
			fmt.Sprintf("%s%s%s", expression.OldLeftPlaceholder, aliasAfter, expression.OldRightPlaceholder))
		pipelineYamlStr = strings.ReplaceAll(pipelineYamlStr, fmt.Sprintf("%s%s:%s", expression.OldLeftPlaceholder, aliasBefore, RefOpOutput),
			fmt.Sprintf("%s%s:%s", expression.OldLeftPlaceholder, aliasAfter, RefOpOutput))

		// ${{ dirs.xxx }} -> ${{ dirs.xxx_xxx }}
		// ${{ outputs.xxx -> ${{ outputs.xxx_xxx
		pipelineYamlStr = strings.ReplaceAll(pipelineYamlStr, fmt.Sprintf("%s %s.%s %s", expression.LeftPlaceholder, expression.Dirs, aliasBefore, expression.RightPlaceholder),
			fmt.Sprintf("%s %s.%s %s", expression.LeftPlaceholder, expression.Dirs, aliasAfter, expression.RightPlaceholder))

		pipelineYamlStr = strings.ReplaceAll(pipelineYamlStr, fmt.Sprintf("%s %s.%s", expression.LeftPlaceholder, expression.Outputs, aliasBefore),
			fmt.Sprintf("%s %s.%s", expression.LeftPlaceholder, expression.Outputs, aliasAfter))

	}
	return pipelineYamlStr
}

func checkParams(specYaml *apistructs.PipelineTemplateSpec, params map[string]interface{}) error {
	paramsList := specYaml.Params
	for _, v := range paramsList {
		if v.Required == false {
			continue
		}

		if params[v.Name] == nil {
			return errors.New(fmt.Sprintf(" param %s need value ", v.Name))
		}

	}

	for name, v := range params {
		typ := reflect.TypeOf(v)
		if typ == nil {
			return errors.New(fmt.Sprintf(" param %s value type error ", name))
		}
		switch typ.Kind() {
		case reflect.String, reflect.Int, reflect.Float32, reflect.Float64, reflect.Bool:
			continue
		default:
			return errors.New(fmt.Sprintf(" param %s value type %v not support ", name, typ.Kind()))
		}
	}

	return nil
}

func setDefaultValue(specYaml *apistructs.PipelineTemplateSpec, params map[string]interface{}) error {
	paramsList := specYaml.Params
	for _, v := range paramsList {
		runValue, ok := params[v.Name]
		if runValue == nil && v.Default == nil && v.Required && ok {
			return fmt.Errorf("params %s is required", v.Name)
		}

		if runValue == nil && v.Default != nil {
			params[v.Name] = v.Default
		}

		if runValue == nil && v.Default == nil {
			params[v.Name] = GetDefaultValue(v.Type)
		}

		if params[v.Name] == nil {
			return fmt.Errorf("params %s is required", v.Name)
		}
	}
	return nil
}

func GetDefaultValue(paramType string) interface{} {
	switch paramType {
	case apistructs.PipelineParamStringType, apistructs.PipelineParamIntType:
		return ""
	case apistructs.PipelineParamBoolType:
		return "false"
	}
	return ""
}

func doReplaceValue(snippet string, params map[string]interface{}) (string, error) {
	snippet = ReplacePipelineParams(snippet, params)
	return snippet, nil
}

func doFormatAndReplaceValue(snippet string, params map[string]interface{}) (string, error) {
	//根据formats中的模板进行逐一渲染
	tpl, err := pongo2.FromString(snippet)
	if err != nil {
		logrus.Errorf("template %s, format error %v", snippet, err)
		return "", err
	}

	out, err := tpl.Execute(params)
	if err != nil {
		logrus.Errorf("template %s, params %v,  to render error %v", params, snippet, err)
		return "", err
	}
	out = ReplacePipelineParams(out, params)
	snippet = out
	return snippet, nil
}

type SnippetCache struct {
	lock               sync.Mutex
	PipelineYamlCaches []SnippetPipelineYmlCache
	err                error
	wg                 sync.WaitGroup
}

func (cache *SnippetCache) getCaches() []SnippetPipelineYmlCache {
	cache.lock.Lock()
	defer func() {
		cache.lock.Unlock()
	}()
	return cache.PipelineYamlCaches
}

func (cache *SnippetCache) setErr(err error) {
	cache.lock.Lock()
	cache.err = err
	cache.lock.Unlock()
}

func (cache *SnippetCache) getErr() error {
	cache.lock.Lock()
	defer func() {
		cache.lock.Unlock()
	}()
	return cache.err
}

func (cache *SnippetCache) appendCache(graph *apistructs.PipelineYml, snippetConfig SnippetConfig) {
	cache.lock.Lock()
	cache.PipelineYamlCaches = append(cache.PipelineYamlCaches, SnippetPipelineYmlCache{
		SnippetConfig: snippetConfig,
		PipelineYaml:  graph,
	})
	cache.lock.Unlock()
}

func (cache *SnippetCache) InitCache(stages [][]*apistructs.PipelineYmlAction, globalSnippetConfigLabels map[string]string) error {
	if cache.PipelineYamlCaches == nil {
		cache.PipelineYamlCaches = make([]SnippetPipelineYmlCache, 0)
	}
	cache.setPipelineYamlCachesByStages(0, stages, globalSnippetConfigLabels)
	cache.wg.Wait()

	return cache.err
}

func (cache *SnippetCache) setPipelineYamlCachesByStages(myDeep int64, stages [][]*apistructs.PipelineYmlAction, globalSnippetConfigLabels map[string]string) {

	if stages == nil {
		return
	}

	// 判定深度是否大于9，大于9下面+1深度就大于10了
	if myDeep > MaxDeep {
		cache.setErr(fmt.Errorf("getSnippetPipelineYaml error: snippet deep > %v，Please verify whether there is a circular dependency", MaxDeep))
		return
	}

	// 传递给下一个递归的深度
	deep := myDeep + 1

	for _, stage := range stages {
		for _, action := range stage {
			if IsSnippetType(ActionType(action.Type)) {
				cache.wg.Add(1)
				go func(action *apistructs.PipelineYmlAction) {
					defer func() {
						cache.wg.Done()
					}()

					if cache.getErr() != nil {
						return
					}

					if action.SnippetConfig == nil {
						cache.setErr(fmt.Errorf(" getSnippetPipelineYaml error: action %s snippet_config empty ", action.Alias))
						return
					}

					snippetConfig := SnippetConfig{
						Name:   action.SnippetConfig.Name,
						Source: action.SnippetConfig.Source,
						Labels: action.SnippetConfig.Labels,
					}
					snippetConfig = HandleSnippetConfigLabel(&snippetConfig, globalSnippetConfigLabels)

					pipelineYamlCaches := cache.getCaches()

					cache.lock.Lock()
					for _, pipelineYamlCache := range pipelineYamlCaches {
						if ComparedSnippet(&pipelineYamlCache.SnippetConfig, &snippetConfig) {
							return
						}
					}
					cache.lock.Unlock()

					graph, err := SnippetToPipelineGraph(snippetConfig)
					if err != nil {
						cache.setErr(err)
						return
					}

					cache.appendCache(graph, snippetConfig)
					cache.setPipelineYamlCachesByStages(deep, graph.Stages, globalSnippetConfigLabels)

					if cache.getErr() != nil {
						return
					}
				}(action)
			}
		}
	}
}
