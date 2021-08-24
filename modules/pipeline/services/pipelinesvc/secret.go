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

package pipelinesvc

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/providers/cms"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclientutil"
	"github.com/erda-project/erda/pkg/nexus"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	secretKeyDockerArtifactRegistry         = "bp.docker.artifact.registry"
	secretKeyDockerArtifactRegistryUsername = "bp.docker.artifact.registry.username"
	secretKeyDockerArtifactRegistryPassword = "bp.docker.artifact.registry.password"

	secretKeyOrgDockerUrl          = "org.docker.url"
	secretKeyOrgDockerPushUsername = "org.docker.push.username"
	secretKeyOrgDockerPushPassword = "org.docker.push.password"
)

// FetchPlatformSecrets 获取平台级别 secrets
// ignoreKeys: 平台只生成不在 ignoreKeys 列表中的 secrets
func (s *PipelineSvc) FetchPlatformSecrets(p *spec.Pipeline, ignoreKeys []string) (map[string]string, error) {
	r := make(map[string]string)

	var operatorID string
	var operatorName string

	if p.Extra.RunUser != nil {
		operatorID = p.GetRunUserID()
		operatorName = p.Extra.RunUser.Name
	}

	var cronTriggerTime string
	if p.Extra.CronTriggerTime != nil && !p.Extra.CronTriggerTime.IsZero() {
		cronTriggerTime = p.Extra.CronTriggerTime.Format(time.RFC3339)
	}

	gittarRepo := getCenterOrSaaSURL(conf.DiceCluster(), p.ClusterName, httpclientutil.WrapHttp(discover.Gittar()), httpclientutil.WrapHttp(conf.GittarPublicURL())) + "/" + p.CommitDetail.RepoAbbr

	abbrevCommit := p.GetCommitID()
	if len(p.GetCommitID()) >= 8 {
		abbrevCommit = p.GetCommitID()[:8]
	}

	clusterInfo, err := s.retryQueryClusterInfo(p.ClusterName, p.ID)
	if err != nil {
		return nil, apierrors.ErrGetCluster.InternalError(err)
	}
	mountPoint := clusterInfo.Get(apistructs.DICE_STORAGE_MOUNTPOINT)
	if mountPoint == "" {
		return nil, errors.Errorf("failed to get necessary cluster info parameter: %s", apistructs.DICE_STORAGE_MOUNTPOINT)
	}

	// if url scheme is file, insert mount point
	storageURL := conf.StorageURL()
	convertURL, err := url.Parse(storageURL)
	if err != nil {
		return nil, errors.Errorf("failed to parse url, url: %s, (%+v)", storageURL, err)
	}

	if convertURL.Scheme == "file" {
		storageURL = strings.Replace(storageURL, convertURL.Path, filepath.Join(mountPoint, convertURL.Path), -1)
	}

	r = map[string]string{
		// dice version
		"dice.version": version.Version,
		// dice
		"dice.org.id":              p.Labels[apistructs.LabelOrgID],
		"dice.org.name":            p.GetOrgName(),
		"dice.project.id":          p.Labels[apistructs.LabelProjectID],
		"dice.project.name":        p.NormalLabels[apistructs.LabelProjectName],
		"dice.application.id":      p.Labels[apistructs.LabelAppID],
		"dice.application.name":    p.NormalLabels[apistructs.LabelAppName],
		"dice.project.application": p.CommitDetail.RepoAbbr,
		"dice.env":                 string(p.Extra.DiceWorkspace),
		"dice.workspace":           string(p.Extra.DiceWorkspace),
		"dice.cluster.name":        p.ClusterName,
		"dice.operator.id":         operatorID,
		"dice.operator.name":       operatorName,
		"dice.user.id":             operatorID,
		"dice.internal.client":     p.Extra.InternalClient,

		// pipeline
		"pipeline.id":                fmt.Sprintf("%d", p.ID),
		"pipeline.type":              string(p.Type),
		"pipeline.trigger.mode":      string(p.TriggerMode),
		"pipeline.cron.expr":         p.Extra.CronExpr,
		"pipeline.cron.trigger.time": cronTriggerTime,

		// gittar
		"gittar.username":      conf.GitInnerUserName(),
		"gittar.password":      conf.GitInnerUserPassword(),
		"gittar.repo":          gittarRepo,
		"gittar.branch":        p.Labels[apistructs.LabelBranch],
		"gittar.commit":        p.GetCommitID(),
		"gittar.commit.abbrev": abbrevCommit,
		"gittar.message":       p.CommitDetail.Comment,
		"gittar.author":        p.CommitDetail.Author,

		// openApi
		"dice.openapi.public.url": conf.OpenAPIPublicURL(),
		"dice.openapi.addr":       discover.Openapi(),

		// buildpack
		"bp.repo.prefix":                "", // 兼容用户 pipeline.yml 里写的 ((bp.repo.prefix))
		"bp.repo.default.version":       "", // 兼容用户 pipeline.yml 里写的 ((bp.repo.default.version))
		"bp.nexus.url":                  httpclientutil.WrapProto(clusterInfo.Get(apistructs.NEXUS_ADDR)),
		"bp.nexus.username":             clusterInfo.Get(apistructs.NEXUS_USERNAME),
		"bp.nexus.password":             clusterInfo.Get(apistructs.NEXUS_PASSWORD),
		secretKeyDockerArtifactRegistry: httpclientutil.RmProto(clusterInfo.Get(apistructs.REGISTRY_ADDR)),
		"bp.docker.cache.registry":      httpclientutil.RmProto(clusterInfo.Get(apistructs.REGISTRY_ADDR)),

		// storage
		"pipeline.storage.url": storageURL,

		// collector 用于主动日志上报(action-agent)
		"collector.addr":       discover.Collector(),
		"collector.public.url": conf.CollectorPublicURL(),

		// others
		"date.YYYYMMDD": time.Now().Format("20060102"),
	}

	// 额外加载 labels，项目级别的流水线，对应的项目名称和企业名称是传递过来的
	if r["dice.org.name"] == "" {
		r["dice.org.name"] = p.Labels[apistructs.LabelOrgName]
	}
	if r["dice.project.name"] == "" {
		r["dice.project.name"] = p.Labels[apistructs.LabelProjectName]
	}

	for _, key := range ignoreKeys {
		delete(r, key)
	}

	return r, nil
}

// 根据 pipeline 对应的 cluster name 信息，返回某一个组件的 vip 或者 sass url
func getCenterOrSaaSURL(diceCluster, requestCluster, center, sass string) string {
	if sass == "" {
		return center
	}
	// 如果和 daemon 在一个集群，则用内网地址
	if diceCluster == requestCluster {
		return center
	}
	return sass
}

// FetchSecrets return secrets, cmsDiceFiles and error.
// holdOnKeys: 声明 key 需要持有，不能被平台 secrets 覆盖，与 ignoreKeys 配合使用
func (s *PipelineSvc) FetchSecrets(p *spec.Pipeline) (secrets, cmsDiceFiles map[string]string, holdOnKeys, encryptSecretKeys []string, err error) {
	secrets = make(map[string]string)
	cmsDiceFiles = make(map[string]string)

	namespaces := p.GetConfigManageNamespaces()

	// 制品是否需要跨集群
	needCrossCluster := false
	if p.Snapshot.AnalyzedCrossCluster != nil && *p.Snapshot.AnalyzedCrossCluster {
		// 企业级 nexus 配置，包含 platform docker registry
		if orgIDStr := p.Labels[apistructs.LabelOrgID]; orgIDStr != "" {
			orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
			if err != nil {
				return nil, nil, nil, nil, errors.Errorf("invalid org id from label %q, err: %v", apistructs.LabelOrgID, err)
			}
			org, err := s.bdl.GetOrg(orgID)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			if org.EnableReleaseCrossCluster {
				namespaces = append(namespaces, nexus.MakeOrgPipelineCmsNs(orgID))
			}
			namespaces = strutil.DedupSlice(namespaces, true)
			needCrossCluster = true
		}
	}

	for _, ns := range namespaces {
		configs, err := s.cmsService.GetCmsNsConfigs(apis.WithInternalClientContext(context.Background(), "pipeline"),
			&pb.CmsNsConfigsGetRequest{
				Ns:             ns,
				PipelineSource: p.PipelineSource.String(),
				GlobalDecrypt:  true,
			})
		if err != nil {
			return nil, nil, nil, nil, err
		}
		for _, c := range configs.Data {
			if c.EncryptInDB && c.Type == cms.ConfigTypeKV {
				encryptSecretKeys = append(encryptSecretKeys, c.Key)
			}
			secrets[c.Key] = c.Value

			// DiceFile 类型，value 为 diceFileUUID
			if c.Type == cms.ConfigTypeDiceFile {
				cmsDiceFiles[c.Key] = c.Value
			}
		}
	}

	// docker artifact registry
	if needCrossCluster {
		secrets[secretKeyDockerArtifactRegistry] = httpclientutil.RmProto(secrets[secretKeyOrgDockerUrl])
		secrets[secretKeyDockerArtifactRegistryUsername] = secrets[secretKeyOrgDockerPushUsername]
		secrets[secretKeyDockerArtifactRegistryPassword] = secrets[secretKeyOrgDockerPushPassword]
		holdOnKeys = append(holdOnKeys,
			secretKeyDockerArtifactRegistry,
			secretKeyDockerArtifactRegistryUsername,
			secretKeyDockerArtifactRegistryPassword,
		)
	}

	return secrets, cmsDiceFiles, holdOnKeys, encryptSecretKeys, nil
}
