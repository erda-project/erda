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

package diceymlchecker

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/commonutil/thirdparty/gittarutil"
	"github.com/erda-project/erda/modules/pipeline/precheck/prechecktype"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

type diceymlCheck struct{}

func New() *diceymlCheck {
	return &diceymlCheck{}
}

func (c *diceymlCheck) Check(ctx context.Context, data interface{}, itemsForCheck prechecktype.ItemsForCheck) (abort bool, message []string) {

	// data type: string
	diceymlContent, ok := data.(string)
	if !ok {
		abort = false
		return
	}

	// validate=false, 是否需要 validate 由 release-action precheker 实现
	d, err := diceyml.New([]byte(diceymlContent), false)
	if err != nil {
		abort = true
		message = append(message, fmt.Sprintf("failed to parse dice.yml without validate, err: %v", err))
		return
	}

	// we can add d.Compose here
	_ = d

	return
}

func checkDiceYmlAndDiceWorkspaceYml(p *spec.Pipeline) error {
	var diceymlName = "dice.yml"
	var diceworkspaceymlName string
	var diceymlworkspace string
	switch p.Extra.DiceWorkspace {
	case apistructs.ProdWorkspace:
		diceymlworkspace = "production"
	case apistructs.StagingWorkspace:
		diceymlworkspace = "staging"
	case apistructs.TestWorkspace:
		diceymlworkspace = "test"
	case apistructs.DevWorkspace:
		diceymlworkspace = "development"
	}
	diceworkspaceymlName = fmt.Sprintf("dice_%s.yml", diceymlworkspace)

	repo := gittarutil.NewRepo(discover.Gittar(), p.CommitDetail.RepoAbbr)
	diceymlcontent, err := repo.FetchFile(p.GetCommitID(), diceymlName)
	if err != nil {
		return err
	}
	if len(diceymlcontent) == 0 {
		return errors.Errorf("%s exist but content is empty", diceymlName)
	}

	// compose dice.yml and dice_workspace.yml
	diceYml, err := diceyml.New([]byte(diceymlcontent), true)
	if err != nil {
		return err
	}

	diceworkspaceymlcontent, err := repo.FetchFile(p.GetCommitID(), diceworkspaceymlName)
	// dice_<workspace>.yml 存在并且有内容
	if err == nil && len(diceworkspaceymlcontent) > 0 {
		diceworkspaceYml, err := diceyml.New([]byte(diceworkspaceymlcontent), false)
		if err != nil {
			return err
		}
		if err = diceYml.Compose(diceymlworkspace, diceworkspaceYml); err != nil {
			return err
		}
	}
	if _, err = diceYml.YAML(); err != nil {
		return err
	}
	return nil
}
