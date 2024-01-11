// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test_case

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/internal/pkg/ai-functions/sdk"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/strutil"
)

//go:embed group-messages.yml
var GroupMessages json.RawMessage

//go:embed group-schema.yaml
var GroupSchema json.RawMessage

type RequirementGroup struct {
	ID     int64
	Groups []string
}

type GroupList struct {
	List []string `json:"list"`
}

func GenerateGroupsForRequirements(ctx context.Context, requirements []*apistructs.Issue, openaiURL *url.URL, xAIProxyModelId string) (requirementIdToGroups map[uint64][]string, err error) {
	userInfo, err := sdk.GetUserInfo(ctx)
	if err != nil {
		return nil, err
	}
	userName := userInfo.Nick
	if userName == "" {
		userName = userInfo.Name
	}

	// get requirements
	results := make([]RequirementGroup, 0)
	var wg sync.WaitGroup

	for _, r := range requirements {
		wg.Add(1)
		go generateGroupsFromRequirement(ctx, &wg, userName, r, &userInfo, openaiURL, xAIProxyModelId, &results)
	}
	wg.Wait()

	// 异步操作可能请求参数、Header 错误或者后端 ai-proxy 服务不可用，导致实际没有生成任何结果，因此要判断是否按目标生成期望结果
	err = verifyAIGenerateResults(results)
	if err != nil {
		return nil, err
	}

	requirementIdToGroups = make(map[uint64][]string)
	for _, rmg := range results {
		requirementIdToGroups[uint64(rmg.ID)] = rmg.Groups
	}
	return requirementIdToGroups, nil
}

func generateGroupsFromRequirement(ctx context.Context, wg *sync.WaitGroup, userName string, requirement *apistructs.Issue, userInfo *apistructs.UserInfo, openaiURL *url.URL, xAIProxyModelId string, results *[]RequirementGroup) (err error) {
	defer wg.Done()
	bdl := bundle.New(bundle.WithErdaServer())

	tasks := make([]string, 0)
	issueRelations, err := bdl.GetIssueRelations(uint64(requirement.ID))
	if err != nil {
		return errors.Wrap(err, "generateGroupsFromRequirement get requirement related tasks info failed")
	}
	for _, include := range issueRelations.IssueInclude {
		if include.Type == apistructs.IssueTypeTask && (include.TaskType == "dev" || include.TaskType == "开发") {
			tasks = append(tasks, include.Title)
		}
	}

	lang := apis.GetLang(ctx)
	messByLang := adjustMessageByLanguage(lang, "")

	// 从参考模板解析基础 messages
	messages := make([]openai.ChatCompletionMessage, 0)
	err = yaml.Unmarshal(GroupMessages, &messages)
	if err != nil {
		return errors.Wrap(err, "yaml.Unmarshal group indirection messages failed")
	}

	// 实际请求
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "This is my requirement title：\nrequirement name: " + requirement.Title,
	})

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "This is my requirement description (markdown format): " + requirement.Content,
	})

	if len(tasks) > 0 {
		for idx, task := range tasks {
			tasks[idx] = "\ntask name:" + task
		}
		taskContent := messByLang.TaskContent + strings.Join(tasks, ",")

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: taskContent,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: messByLang.GenerateGroup,
	})

	schema, err := strutil.YamlOrJsonToJson(GroupSchema)
	if err != nil {
		return errors.Wrap(err, "get group schema failed")
	}
	fd := openai.FunctionDefinition{
		Name:        "Group",
		Description: "Get Group info from requirement info",
		Parameters:  schema,
	}
	logrus.Debugf("openai.FunctionDefinition fd.Parameters string: %s\n", fd.Parameters)

	options := &openai.ChatCompletionRequest{
		Messages:     messages,
		Functions:    []openai.FunctionDefinition{fd},
		FunctionCall: openai.FunctionCall{Name: fd.Name},
		Temperature:  0.5,
		Stream:       false,
		Model:        "gpt4",
	}

	optionsStr, _ := json.Marshal(*options)
	logrus.Infof("Try AI Gernerate group for requirementId=%d  with req=%s\n", requirement.ID, string(optionsStr))

	reqOpts := []sdk.RequestOption{
		sdk.RequestOptionWithResetAPIVersion("2023-07-01-preview"),
	}

	// 在 request option 中添加认证信息: 以某组织下某用户身份调用 ai-proxy,
	// ai-proxy 中的 filter erda-auth 会回调 erda.cloud 的 openai, 检查该企业和用户是否有权使用 AI 能力
	ros := append(reqOpts, func(r *http.Request) {
		r.Header.Set(vars.XAIProxySource, base64.StdEncoding.EncodeToString([]byte("testcase___"+os.Getenv(string(apistructs.DICE_CLUSTER_NAME)))))
		r.Header.Set(vars.XAIProxyOrgId, base64.StdEncoding.EncodeToString([]byte(apis.GetOrgID(ctx))))
		r.Header.Set(vars.XAIProxyUserId, base64.StdEncoding.EncodeToString([]byte(apis.GetUserID(ctx))))
		r.Header.Set(vars.XAIProxyEmail, base64.StdEncoding.EncodeToString([]byte(userInfo.Email)))
		r.Header.Set(vars.XAIProxyName, base64.StdEncoding.EncodeToString([]byte(userName)))
		r.Header.Set(vars.XAIProxyPhone, base64.StdEncoding.EncodeToString([]byte(userInfo.Phone)))
		// TODO: 获取表示使用 chatGPT4 对应的 X-AI-Proxy-Model-Id 的值
		r.Header.Set(vars.XAIProxyModelId, xAIProxyModelId)
	})
	client, err := sdk.NewClient(openaiURL, http.DefaultClient, ros...)
	if err != nil {
		return err
	}
	completion, err := client.CreateCompletion(ctx, options)
	if err != nil {
		return errors.Wrap(err, "failed to CreateCompletion for group")
	}

	if len(completion.Choices) == 0 || completion.Choices[0].Message.FunctionCall == nil {
		return errors.New("no idea for group")
	}
	arguments, err := strutil.YamlOrJsonToJson([]byte(completion.Choices[0].Message.FunctionCall.Arguments))
	if err != nil {
		arguments = json.RawMessage(completion.Choices[0].Message.FunctionCall.Arguments)
	}

	if err = sdk.VerifyArguments(schema, arguments); err != nil {
		return errors.Wrap(err, "invalid arguments from FunctionCall Group")
	}

	var groupList GroupList
	if err := json.Unmarshal(arguments, &groupList); err != nil {
		return errors.Wrap(err, "Unmarshal arguments to GroupList failed")
	}

	groupMap := make(map[string]bool)
	groups := make([]string, 0)
	for _, groupName := range groupList.List {
		groupMap[groupName] = true
	}
	// 分组名称去重
	for name := range groupMap {
		groups = append(groups, name)
	}
	logrus.Infof("Get AI Groups for requirementId=%d  groups=%+v\n", requirement.ID, groups)
	if len(groups) == 0 {
		return errors.Errorf("Get 0 groups for requirementId=%d\n", requirement.ID)
	}
	*results = append(*results, RequirementGroup{
		ID:     requirement.ID,
		Groups: groups,
	})

	return nil
}
