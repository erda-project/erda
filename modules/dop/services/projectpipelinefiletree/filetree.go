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

package projectpipelinefiletree

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/strutil"
)

// Pipeline pipeline 结构体
type FileTree struct {
	bdl *bundle.Bundle
}

// Option Pipeline 配置选项
type Option func(*FileTree)

// New Pipeline service
func New(options ...Option) *FileTree {
	r := &FileTree{}
	for _, op := range options {
		op(r)
	}
	return r
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(f *FileTree) {
		f.bdl = bdl
	}
}

func (svc *FileTree) ListFileTreeNodes(req apistructs.UnifiedFileTreeNodeListRequest, orgID uint64) ([]apistructs.UnifiedFileTreeNode, error) {

	var treeNodes []apistructs.UnifiedFileTreeNode

	if err := strutil.Validate(req.ScopeID, strutil.MinLenValidator(1)); err != nil {
		return nil, fmt.Errorf("invalid scopeID: %v", err)
	}

	// scope 为空代表查询所有的目录名称
	if req.Scope == "" {
		for _, scope := range apistructs.AllScope {
			treeNodes = append(treeNodes, scopeConvertToUnifiedFileTreeNode(scope, req.ScopeID))
		}
		return treeNodes, nil
	}

	if err := req.BasicValidate(); err != nil {
		return nil, apierrors.ErrGetFileTreeNode.InvalidParameter(err)
	}

	return svc.bdl.ListFileTreeNodes(req, orgID)
}

func (svc *FileTree) GetFileTreeNode(req apistructs.UnifiedFileTreeNodeGetRequest, orgID uint64) (*apistructs.UnifiedFileTreeNode, error) {
	// 参数校验
	if err := req.BasicValidate(); err != nil {
		return nil, apierrors.ErrGetFileTreeNode.InvalidParameter(err)
	}

	result, err := svc.bdl.GetFileTreeNode(req, orgID)
	if err != nil {
		return nil, err
	}

	// 设置上snippet_source标识
	if result != nil {

		snippetConfigMapJson, err := json.Marshal(result.Meta[apistructs.AutoTestFileTreeNodeMetaKeySnippetAction])
		if err != nil {
			return result, nil
		}
		snippetAction := apistructs.PipelineYmlAction{}
		err = json.Unmarshal(snippetConfigMapJson, &snippetAction)
		if err != nil {
			return result, nil
		}

		snippetAction.SnippetConfig.Labels[apistructs.LabelSnippetScope] = req.Scope
		result.Meta[apistructs.AutoTestFileTreeNodeMetaKeySnippetAction] = snippetAction
		return result, nil
	}

	return result, nil
}

func (svc *FileTree) FuzzySearchFileTreeNodes(req apistructs.UnifiedFileTreeNodeFuzzySearchRequest, orgID uint64) ([]apistructs.UnifiedFileTreeNode, error) {
	// fuzzy
	if req.PrefixFuzzy == "" && req.SuffixFuzzy == "" && req.Fuzzy == "" {
		if req.CreatorID == "" {
			return nil, fmt.Errorf("missing fuzzy condition")
		}
	}

	return svc.bdl.FuzzySearchFileTreeNodes(req, orgID)
}

func (svc *FileTree) DeleteFileTreeNode(req apistructs.UnifiedFileTreeNodeDeleteRequest, orgID uint64) (*apistructs.UnifiedFileTreeNode, error) {
	if err := req.BasicValidate(); err != nil {
		return nil, apierrors.ErrGetFileTreeNode.InvalidParameter(err)
	}
	if err := strutil.Validate(req.ScopeID, strutil.MinLenValidator(1)); err != nil {
		return nil, fmt.Errorf("invalid scopeID: %v", err)
	}
	if err := strutil.Validate(req.Scope, strutil.MinLenValidator(1)); err != nil {
		return nil, fmt.Errorf("invalid scope: %v", err)
	}

	return svc.bdl.DeleteFileTreeNodes(req, orgID)
}

func (svc *FileTree) CreateFileTreeNode(req apistructs.UnifiedFileTreeNodeCreateRequest, orgID uint64) (*apistructs.UnifiedFileTreeNode, error) {
	// 创建前校验
	if err := req.BasicValidate(); err != nil {
		return nil, apierrors.ErrGetFileTreeNode.InvalidParameter(err)
	}
	if err := strutil.Validate(req.ScopeID, strutil.MinLenValidator(1)); err != nil {
		return nil, fmt.Errorf("invalid scopeID: %v", err)
	}
	if err := strutil.Validate(req.Scope, strutil.MinLenValidator(1)); err != nil {
		return nil, fmt.Errorf("invalid scope: %v", err)
	}

	return svc.bdl.CreateFileTreeNodes(req, orgID)
}

func (svc *FileTree) FindFileTreeNodeAncestors(req apistructs.UnifiedFileTreeNodeFindAncestorsRequest, orgID uint64) ([]apistructs.UnifiedFileTreeNode, error) {
	// 参数校验
	if err := req.BasicValidate(); err != nil {
		return nil, apierrors.ErrFindFileTreeNodeAncestors.InvalidParameter(err)
	}

	return svc.bdl.FindFileTreeNodeAncestors(req, orgID)
}

func scopeConvertToUnifiedFileTreeNode(scope string, scopeID string) apistructs.UnifiedFileTreeNode {
	return apistructs.UnifiedFileTreeNode{
		Type:    apistructs.UnifiedFileTreeNodeTypeDir,
		Name:    getScopeName(scope),
		ScopeID: scopeID,
		Scope:   scope,
		Pinode:  "-1",
		Inode:   "0",
	}
}

func getScopeName(scope string) string {
	switch scope {
	case apistructs.FileTreeScopeAutoTest:
		return "自动化测试用例"
	case apistructs.FileTreeScopeAutoTestConfigSheet:
		return "配置单用例"
	case apistructs.FileTreeScopeProjectApp:
		return "应用流水线"
	case apistructs.FileTreeScopeProject:
		return "项目流水线"
	case apistructs.FileTreeScopeAutoTestPlan:
		return "自动化测试计划"
	}

	return "未知"
}
