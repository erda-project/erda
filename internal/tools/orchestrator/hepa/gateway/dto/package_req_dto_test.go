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

package dto_test

import (
	"testing"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
)

func TestPackageDto_CheckValid(t *testing.T) {
	var pkgDto dto.PackageDto
	pkgDto.Scene = ""
	if err := pkgDto.CheckValid(); err == nil {
		t.Fatal("pkgDto should be invalid")
	} else {
		t.Log(err)
	}

	pkgDto.Scene = orm.OpenapiScene
	if err := pkgDto.CheckValid(); err == nil {
		t.Fatal("pkgDto should be invalid")
	} else {
		t.Log(err)
	}

	pkgDto.Name = "some-package"
	if err := pkgDto.CheckValid(); err == nil {
		t.Fatal("pkg should be invalid")
	} else {
		t.Log(err)
	}

	pkgDto.BindDomain = []string{
		"baidu.com",
		"google.com",
	}
	pkgDto.NeedBindCloudapi = false
	pkgDto.AuthType = dto.AT_ALIYUN_APP
	if err := pkgDto.CheckValid(); err == nil {
		t.Fatal("pkgDto should be invalid")
	} else {
		t.Log(err)
	}

	pkgDto.NeedBindCloudapi = true
	if err := pkgDto.CheckValid(); err != nil {
		t.Fatalf("err should be nil, got: %v", err)
	}
}
