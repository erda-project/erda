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

package apidocsvc

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/swagger/ddlconv"
	"github.com/erda-project/erda/pkg/swagger/oas3"
	"github.com/erda-project/erda/pkg/swagger/oasconv"

	"github.com/erda-project/erda/modules/dop/bdl"
	"github.com/erda-project/erda/modules/dop/dbclient"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/dop/services/uc"
)

func (svc *Service) createDoc(orgID uint64, userID string, dstPinode, serviceName, content string) (*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	orgIDStr := strconv.FormatUint(orgID, 10)
	filenameFromRepoRoot := filepath.Join(apiDocsPathFromRepoRoot, mustSuffix(serviceName, ".yaml"))

	// pft: parentFileTree object
	pft, err := bundle.NewGittarFileTree(dstPinode)
	if err != nil || pft.ApplicationID() == "" {
		return nil, apierrors.CreateNode.InvalidParameter(errors.Wrap(err, "不合法的父节点编号"))
	}

	// 无需前置检查分支是否存在, 如不存在, commit 时会返回错误
	// 无需检查和创建 .dice/apidocs 目录, commit 时会自动创建

	// 检查父目录中是否已存在同名文档
	ft := pft.Clone().SetPathFromRepoRoot(filenameFromRepoRoot)
	if _, err := bdl.Bdl.GetGittarTreeNodeInfo(ft.TreePath(), orgIDStr); err == nil {
		return nil, apierrors.CreateNode.InvalidParameter("已存在同名文档")
	}

	repo := strings.TrimPrefix(pft.RepoPath(), "/")
	message := "create api doc from API Design Center"
	if err = CommitAPIDocCreation(orgID, userID, repo, message, serviceName, content, pft.BranchName()); err != nil {
		return nil, apierrors.CreateNode.InternalError(err)
	}

	rspData := apistructs.FileTreeNodeRspData{
		Type:      "f",
		Inode:     ft.Inode(),
		Pinode:    dstPinode,
		Scope:     "application",
		ScopeID:   ft.ApplicationID(),
		Name:      mustSuffix(serviceName, ""),
		CreatorID: userID,
		UpdaterID: userID,
		Meta:      nil,
	}

	return &rspData, nil
}

func (svc *Service) deleteAPIDoc(orgID uint64, userID, dstInode string) *errorresp.APIError {
	// todo: 鉴权 谁能删

	// 查询这个 gittar 路径对应的节点
	ft, err := bundle.NewGittarFileTree(dstInode)
	if err != nil {
		return apierrors.DeleteNode.InvalidParameter(err)
	}
	nodeInfo, err := bdl.Bdl.GetGittarTreeNodeInfo(ft.TreePath(), strconv.FormatUint(orgID, 10))
	if err != nil {
		logrus.Errorf("查询文档节点失败, ft: %+v, err: %v", ft, err)
		return apierrors.DeleteNode.InternalError(errors.Wrap(err, "查询文档节点失败"))
	}
	if nodeInfo.Type != "blob" {
		return apierrors.DeleteNode.InvalidParameter("只可以删除文档文件, 不可以删除目录")
	}

	// 正在编辑的文档不可删除
	var (
		lock  apistructs.APIDocLockModel
		where = map[string]interface{}{
			"application_id": ft.ApplicationID(),
			"branch_name":    ft.BranchName(),
			"doc_name":       filepath.Base(ft.PathFromRepoRoot()),
		}
		timeNow = time.Now()
	)
	if record := dbclient.Sq().Where(where).
		Where("is_locked = true AND expired_at > ?", timeNow).
		First(&lock); record.RowsAffected > 0 {
		return apierrors.DeleteNode.InternalError(errors.New("文档正在编辑中, 不能删除"))
	}

	// 删除这个文件
	var commitRequest = apistructs.GittarCreateCommitRequest{
		Message: "从 API 设计中心删除了文件 " + nodeInfo.Path,
		Actions: []apistructs.EditActionItem{{
			Action:   "delete",
			Content:  "",
			Path:     ft.PathFromRepoRoot(),
			PathType: nodeInfo.Type,
		}},
		Branch: ft.BranchName(),
	}

	resp, err := bdl.Bdl.CreateGittarCommitV3(orgID, userID, ft.RepoPath(), &commitRequest)
	if err != nil {
		return apierrors.DeleteNode.InternalError(err)
	}
	if !resp.Success {
		return apierrors.DeleteNode.InternalError(errors.Errorf("%+v", resp.Error))
	}

	return nil
}

func (svc *Service) renameAPIDoc(orgID uint64, userID, dstInode, docName string) (*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	ft, err := bundle.NewGittarFileTree(dstInode)
	if err != nil {
		return nil, apierrors.UpdateNode.InvalidParameter("不合法的 inode")
	}

	// 编辑中的文档不能重命名
	var (
		lock  apistructs.APIDocLockModel
		where = map[string]interface{}{
			"application_id": ft.ApplicationID(),
			"branch_name":    ft.BranchName(),
			"doc_name":       filepath.Base(ft.PathFromRepoRoot()),
		}
		timeNow = time.Now()
	)
	if record := dbclient.Sq().Where(where).
		Where("is_locked = true AND expired_at > ?", timeNow).
		First(&lock); record.RowsAffected > 0 {
		return nil, apierrors.DeleteNode.InternalError(errors.New("文档正在编辑中, 不能重命名"))
	}

	// 检查 inode 是否是用户自己构造出来的 (不允许指向别的目录)
	if filepath.Dir(ft.PathFromRepoRoot()) != apiDocsPathFromRepoRoot {
		return nil, apierrors.UpdateNode.InternalError(errors.New("不合法的 inode"))
	}

	docName = mustSuffix(docName, suffixYaml)
	if docName == filepath.Base(ft.PathFromRepoRoot()) {
		return nil, apierrors.UpdateNode.InvalidParameter(errors.New("重命名的文档名与原文档名一致"))
	}

	// 目标名称不能与目录下文件名同名
	newPath := filepath.Join(apiDocsPathFromRepoRoot, docName)
	nft := ft.Clone().SetPathFromRepoRoot(newPath) // nft: newFileTree
	if _, err := bdl.Bdl.GetGittarTreeNodeInfo(nft.TreePath(), strconv.FormatUint(orgID, 10)); err == nil {
		return nil, apierrors.CreateNode.InvalidParameter("目标名称与其他节点重名")
	}

	// 查询文档内容
	contentRsp, err2 := svc.getAPIDocContent(orgID, userID, dstInode)
	if err2 != nil {
		return nil, err2
	}
	if contentRsp.Meta == nil {
		return nil, apierrors.UpdateNode.InternalError(errors.New("未查询到原文档内容"))
	}

	var apiDocMeta apistructs.APIDocMeta
	if err = json.Unmarshal(contentRsp.Meta, &apiDocMeta); err != nil {
		return nil, apierrors.UpdateNode.InternalError(errors.Wrap(err, "未查询到原文档内容"))
	}
	if apiDocMeta.Blob == nil {
		return nil, apierrors.UpdateNode.InternalError(errors.New("未查询到原文档内容, blob is nil"))
	}

	// 提交修改
	var commit = apistructs.GittarCreateCommitRequest{
		Message: fmt.Sprintf("文档 %s 重命名为 %s", ft.PathFromRepoRoot(), docName),
		Actions: []apistructs.EditActionItem{{
			Action:   "add",
			Content:  apiDocMeta.Blob.Content,
			Path:     newPath,
			PathType: "blob",
		}, {
			Action:   "delete",
			Content:  "",
			Path:     ft.PathFromRepoRoot(),
			PathType: "blob",
		}},
		Branch: ft.BranchName(),
	}

	resp, err := bdl.Bdl.CreateGittarCommitV3(orgID, userID, ft.RepoPath(), &commit)
	if err != nil {
		return nil, apierrors.UpdateNode.InternalError(err)
	}
	if !resp.Success {
		return nil, apierrors.UpdateNode.InternalError(errors.Errorf("%+v", resp.Error))
	}

	data := &apistructs.FileTreeNodeRspData{
		Type:      "f",
		Inode:     nft.Inode(),
		Pinode:    nft.Clone().DeletePathFromRepoRoot().Inode(),
		Scope:     "application",
		ScopeID:   nft.ApplicationID(),
		Name:      mustSuffix(docName, ""),
		CreatorID: "",
		UpdaterID: "",
		Meta:      nil,
	}
	return data, nil
}

func (svc *Service) moveAPIDco(orgID uint64, userID, srcInode, dstPinode string) (*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	ft, err := bundle.NewGittarFileTree(srcInode)
	if err != nil {
		return nil, apierrors.MoveNode.InvalidParameter(errors.New("不合法的 inode"))
	}

	// 正在编辑中的文档不能移动
	var lock apistructs.APIDocLockModel
	if err = dbclient.Sq().First(&lock, map[string]interface{}{
		"project_id":     ft.ProjectID(),
		"application_id": ft.ApplicationID(),
		"branch_name":    ft.BranchName(),
		"doc_name":       filepath.Base(ft.PathFromRepoRoot()),
	}).Error; err == nil {
		return nil, apierrors.DeleteNode.InternalError(errors.New("文档正在编辑中, 不能删除"))
	}

	// 拷贝文档到目标分支
	data, err2 := svc.copyAPIDoc(orgID, userID, srcInode, dstPinode)
	if err2 != nil {
		return nil, err2
	}

	// 在本分支中删除文档
	if err2 := svc.deleteAPIDoc(orgID, userID, srcInode); err2 != nil {
		return nil, err2
	}

	return data, nil
}

func (svc *Service) copyAPIDoc(orgID uint64, userID, srcInode, dstPinode string) (*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	// 查找文档内容
	ft, err := bundle.NewGittarFileTree(srcInode)
	if err != nil {
		return nil, apierrors.CopyNode.InvalidParameter("不合法的 inode")
	}
	if dstPinode == ft.Clone().DeletePathFromRepoRoot().Inode() {
		return nil, apierrors.CopyNode.InvalidParameter("目标分支不得与源分支相同")
	}
	blob, err := bdl.Bdl.GetGittarBlobNodeInfo(ft.BlobPath(), strconv.FormatUint(orgID, 10))
	if err != nil {
		return nil, apierrors.CopyNode.InternalError(err)
	}

	// 将文档内容写入目标路径
	data, err2 := svc.createDoc(orgID, userID, dstPinode, filepath.Base(blob.Path), blob.Content)
	if err2 != nil {
		return nil, err2
	}

	return data, nil
}

// 查询应用下所有的分支, 构成节点列表
func (svc *Service) listBranches(orgID, appID uint64, userID string) ([]*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	appIDStr := strconv.FormatUint(appID, 10)

	// 查询 application (由于不存在项目级目录树, 所以不讨论 pinode != 0 的情况)
	branches, err := svc.branchRuleSvc.GetAllValidBranchWorkspaces(int64(appID))
	if err != nil {
		return nil, apierrors.ListChildrenNodes.InternalError(errors.Wrap(err, "failed to GetAllValidBranchWorkspace"))
	}
	// 查询 application 数据构造目录树 node
	app, err := bdl.Bdl.GetApp(appID)
	if err != nil {
		return nil, apierrors.ListChildrenNodes.InternalError(errors.Wrap(err, "failed to GetApp"))
	}
	proIDStr := strconv.FormatUint(app.ProjectID, 10)
	ft, _ := bundle.NewGittarFileTree("")
	ft.SetProjectIDName(proIDStr, app.ProjectName)
	ft.SetApplicationIDName(appIDStr, app.Name)

	var (
		results      []*apistructs.FileTreeNodeRspData
		isManager, _ = bdl.IsManager(userID, apistructs.AppScope, appID)
		m            = sync.Map{}
		w            = sync.WaitGroup{}
	)

	for _, branch := range branches {
		branch := branch
		w.Add(1)
		go func() {
			branchInode := ft.Clone().SetBranchName(branch.Name).Inode()
			hasAPIDoc := branchHasAPIDoc(orgID, branchInode)
			m.Store(branch, hasAPIDoc)
			w.Done()
		}()
	}
	w.Wait()

	m.Range(func(key, value interface{}) bool {
		branch := key.(*apistructs.ValidBranch)
		readOnly := !isManager && branch.IsProtect
		meta, _ := json.Marshal(map[string]bool{"readOnly": readOnly, "hasDoc": value.(bool)})
		results = append(results, &apistructs.FileTreeNodeRspData{
			Type:      "d",
			Inode:     ft.Clone().SetBranchName(branch.Name).Inode(),
			Pinode:    ft.Inode(),
			Scope:     "application",
			ScopeID:   appIDStr,
			Name:      branch.Name,
			CreatorID: "",
			UpdaterID: "",
			Meta:      meta,
		})

		return true
	})

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results, nil
}

// 查询分支下所有 API 文档
func (svc *Service) listAPIDocs(orgID uint64, userID, pinode string) ([]*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	return svc.listServices(orgID, userID, pinode, apiDocsPathFromRepoRoot, func(node apistructs.TreeEntry) bool {
		return node.Type == "blob" && matchSuffix(node.Name, suffixYaml, suffixYml)
	})
}

// 查询分支下所有的 migration 的 service 名称, 即 migration 的目录名
func (svc *Service) listSchemas(orgID uint64, userID, pinode string) ([]*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	return svc.listServices(orgID, userID, pinode, migrationsPathFromRepoRoot, func(node apistructs.TreeEntry) bool {
		return node.Type == "tree"
	})
}

// 列出目录树的 service 层的节点列表
func (svc *Service) listServices(orgID uint64, userID, pinode, pathFromRepoRoot string, filter func(node apistructs.TreeEntry) bool) ([]*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	ft, err := bundle.NewGittarFileTree(pinode)
	if err != nil {
		return nil, apierrors.ListChildrenNodes.InvalidParameter(err)
	}
	ft.SetPathFromRepoRoot(pathFromRepoRoot)

	appID, err := strconv.ParseUint(ft.ApplicationID(), 10, 64)
	if err != nil {
		return nil, apierrors.ListChildrenNodes.InvalidParameter(err)
	}

	orgIDStr := strconv.FormatUint(orgID, 10)

	// 查找目录下的文档. 允许错误, 错误则认为目录下没有任何文档
	nodes, err := bdl.Bdl.GetGittarTreeNode(ft.TreePath(), orgIDStr, true)
	if err != nil {
		logrus.Errorf("failed to GetGittarTreeNode, err: %v", err)
	}
	if len(nodes) == 0 {
		return nil, nil
	}

	// 文档是否只读: 如果用户是应用管理员, 则
	isManager, _ := bdl.IsManager(userID, apistructs.AppScope, appID)
	readOnly := !isManager && isBranchProtected(ft.ApplicationID(), ft.BranchName(), svc.branchRuleSvc)
	meta, _ := json.Marshal(map[string]bool{"readOnly": readOnly})

	var (
		results   []*apistructs.FileTreeNodeRspData
		nodeTypeM = map[string]apistructs.NodeType{"blob": "f", "tree": "d"}
	)
	for _, node := range nodes {
		if !filter(node) {
			continue
		}
		// cft: childFileTree
		cft := ft.Clone().SetPathFromRepoRoot(filepath.Join(ft.PathFromRepoRoot(), node.Name))
		results = append(results, &apistructs.FileTreeNodeRspData{
			Type:      nodeTypeM[node.Type],
			Inode:     cft.Inode(),
			Pinode:    pinode,
			Scope:     "application",
			ScopeID:   cft.ApplicationID(),
			Name:      mustSuffix(node.Name, ""),
			CreatorID: "",
			UpdaterID: "",
			Meta:      meta,
		})
	}

	return results, nil
}

func (svc *Service) getAPIDocContent(orgID uint64, userID, inode string) (*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	data, apiError := FetchAPIDocContent(orgID, userID, inode, oasconv.OAS3JSON, svc.branchRuleSvc)
	if apiError != nil {
		return nil, apiError
	}

	var meta apistructs.APIDocMeta
	if err := json.Unmarshal(data.Meta, &meta); err != nil {
		return nil, apierrors.GetNodeDetail.InternalError(err)
	}

	if meta.Blob == nil {
		marshal, _ := json.Marshal(meta)
		data.Meta = marshal
		return data, nil
	}

	// 查询文档锁状态
	meta.Lock = &apistructs.APIDocMetaLock{Locked: false}
	if ft, err := bundle.NewGittarFileTree(inode); err == nil {
		var lock apistructs.APIDocLockModel
		var where = map[string]interface{}{
			"application_id": ft.ApplicationID(),
			"branch_name":    ft.BranchName(),
			"doc_name":       filepath.Base(ft.TreePath()),
		}
		if record := dbclient.Sq().First(&lock, where); record.RowsAffected > 0 {
			meta.Lock.Locked = true
			meta.Lock.UserID = lock.CreatorID
			if users, _ := uc.GetUsers([]string{lock.CreatorID}); len(users) > 0 {
				if user := users[lock.CreatorID]; user != nil {
					meta.Lock.NickName = user.Nick
				}
			}
		}
	}

	// 校验 oas3 格式合法性 只做基础的反序列化校验
	meta.Valid = true
	v3, err := oas3.LoadFromData([]byte(meta.Blob.Content))
	if err != nil {
		meta.Valid = false
		meta.Error = err.Error()
		marshal, _ := json.Marshal(meta)
		data.Meta = marshal
		return data, nil
	}

	// 查询集市相关信息
	meta.Asset = new(apistructs.APIDocMetaAssetInfo)
	var version apistructs.APIAssetVersionsModel
	first := dbclient.Sq().First(&version, map[string]interface{}{
		"source":       apistructs.CreateAPIAssetSourceDesignCenter,
		"service_name": data.Name,
	})
	if first.Error != nil {
		meta.Asset.Major = 1
	} else {
		meta.Asset.AssetID = version.AssetID
		meta.Asset.AssetName = version.AssetName
		_ = dbclient.GenSemVer(orgID, version.AssetID, v3.Info.Version, &meta.Asset.Major, &meta.Asset.Minor, &meta.Asset.Patch)
	}

	data.Meta, _ = json.Marshal(meta)

	return data, nil
}

func (svc *Service) getSchemaContent(orgID uint64, inode string) (*apistructs.FileTreeNodeRspData, *errorresp.APIError) {
	// todo: 鉴权 谁能查

	// 解析路径
	ft, err := bundle.NewGittarFileTree(inode)
	if err != nil {
		return nil, apierrors.GetNodeDetail.InternalError(err)
	}
	serviceName := mustSuffix(filepath.Base(ft.PathFromRepoRoot()), "")
	ft.SetPathFromRepoRoot(filepath.Join(migrationsPathFromRepoRoot, serviceName))

	orgIDStr := strconv.FormatUint(orgID, 10)

	// 查找服务目录下所有的子目录, 允许错误和结果数量为 0
	nodes, err := bdl.Bdl.GetGittarTreeNode(ft.TreePath(), orgIDStr, true)
	if err != nil {
		logrus.Errorf("failed to GetGittarTreeNode, err: %v", err)
	}
	if len(nodes) == 0 {
		return nil, nil
	}

	// 按版本编号排序
	sort.Slice(nodes, func(i, j int) bool {
		nodeIs := strings.Split(nodes[i].Name, "_")
		nodeJs := strings.Split(nodes[j].Name, "_")
		numI, err := strconv.ParseUint(strings.TrimLeft(nodeIs[0], "0"), 10, 64)
		if err != nil {
			return false
		}
		numJ, err := strconv.ParseUint(strings.TrimLeft(nodeJs[0], "0"), 10, 64)
		if err != nil {
			return false
		}
		return numI < numJ
	})

	var openapi, _ = ddlconv.New()
	for _, node := range nodes {
		// 服务目录名应当以数字开头
		if c := node.Name[0]; c < '0' || c > '9' {
			continue
		}

		// .dice/migrations/{svc_name} ==> .dice/migrations/{svc_name}/{ver_name}/schema.sql
		cft := ft.Clone().SetPathFromRepoRoot(filepath.Join(ft.PathFromRepoRoot(), node.Name, "schema.sql"))

		// 查询 schema.sql 的内容
		nodeInfo, err := bdl.Bdl.GetGittarBlobNodeInfo(cft.BlobPath(), orgIDStr)
		if err != nil {
			logrus.Errorf("failed to GetGittarBlobNodeInfo, blobPath: %s, err: %v", cft.BlobPath(), err)
			continue
		}

		if _, err = openapi.WriteString(nodeInfo.Content); err != nil {
			logrus.Errorf("failed to WriteString(sql) to openapi, err: %v", err)
			return nil, apierrors.GetNodeDetail.InternalError(err)
		}
	}

	meta, _ := json.Marshal(openapi.Components)

	var data = apistructs.FileTreeNodeRspData{
		Type:      "f",
		Inode:     inode,
		Pinode:    ft.Clone().DeletePathFromRepoRoot().Inode(),
		Scope:     "application",
		ScopeID:   ft.ApplicationID(),
		Name:      path.Base(ft.PathFromRepoRoot()),
		CreatorID: "",
		UpdaterID: "",
		Meta:      json.RawMessage(meta),
	}

	return &data, nil
}
