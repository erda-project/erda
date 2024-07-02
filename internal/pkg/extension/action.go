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

package extension

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda-proto-go/core/extension/pb"
	"github.com/erda-project/erda/apistructs"
)

var (
	// baseTime is used to compare extension file mod time
	baseTime = time.Date(2000, 1, 1, 1, 1, 1, 1, time.Local)
)

type Repo struct {
	// addr workPath
	addr string
	// versions ExtensionVersion dir path
	versions []string
}

// Version is a version of an Extension
type Version struct {
	Name    string
	Dirname string

	Spec          *apistructs.Spec // structure of spec.yml
	SpecContent   []byte           // content of spec.yml
	DiceContent   []byte           // content of dice.yml
	ReadmeContent []byte           // content of readme.md

	SwaggerContent []byte // content of swagger.yml

	UpdateAt *timestamppb.Timestamp
}

func (s *provider) InitExtension(addr string, forceUpdate bool) error {
	logrus.Infoln("Start init extension")

	// get all extensionVersion in repo
	repo := LoadExtensions(addr)

	// get all extensions existed
	allActionVersions, err := s.db.QueryAllExtensions()
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return err
		}
	}

	// extensionVersionMap key:ExtensionName value:ExtensionVersion
	// one extension can have version more then 1
	extensionVersionMap := make(map[string][]string)

	// extensionTypeMap key:ExtensionName value:ExtensionType
	extensionTypeMap := make(map[string][]string)

	// add all actionVersions to map
	for _, v := range allActionVersions {
		version := extensionVersionMap[v.Name]
		version = append(version, v.Version)
		extensionVersionMap[v.Name] = version

		specData := apistructs.Spec{}
		err = yaml.Unmarshal([]byte(v.Spec), &specData)
		if err != nil {
			return err
		}

		extensionType := extensionTypeMap[v.Name]
		extensionType = append(extensionType, specData.Type)
		extensionTypeMap[v.Name] = extensionType
	}

	// push all actionVersions
	for _, v := range repo.versions {
		name, version, err := s.RunExtensionsPush(v, extensionVersionMap, extensionTypeMap, forceUpdate)
		if err == nil {
			logrus.Infoln("extension create success, name: ", name, ", version: ", version)
		} else {
			logrus.Infoln("extension create false, name: ", name, ", version: ", version, " err: ", err)
		}
	}
	return nil
}

// RunExtensionsPush push extensions
func (s *provider) RunExtensionsPush(dir string, extensionVersionMap, extensionTypeMap map[string][]string, forceUpdate bool) (string, string, error) {
	version, err := NewVersion(dir)
	if err != nil {
		return "", "", err
	}

	specData := version.Spec
	if specData.Type != s.Cfg.ReloadExtensionType {
		return "", "", errors.Errorf("invalid extension type: %s, want: %s", specData.Type, s.Cfg.ReloadExtensionType)
	}

	// if extension is existed, return
	versionNow := extensionVersionMap[specData.Name]
	typeNow := extensionTypeMap[specData.Name]
	needCreate := true
	for i, version := range versionNow {
		if version == specData.Version && typeNow[i] == specData.Type {
			needCreate = false
			break
		}
	}
	if !needCreate && !forceUpdate {
		return specData.Name, specData.Version, errors.New("extension is existed")
	}

	var request = &pb.ExtensionVersionCreateRequest{
		Name:        specData.Name,
		Version:     specData.Version,
		SpecYml:     string(version.SpecContent),
		DiceYml:     string(version.DiceContent),
		SwaggerYml:  string(version.SwaggerContent),
		Readme:      string(version.ReadmeContent),
		Public:      specData.Public,
		ForceUpdate: forceUpdate,
		All:         true,
		IsDefault:   specData.IsDefault,
		UpdatedAt:   version.UpdateAt,
	}

	_, err = s.CreateExtensionVersionByRequest(request)
	if err != nil {
		return request.Name, request.Version, err
	}

	return request.Name, request.Version, err
}

func NewVersion(dirname string) (*Version, error) {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ReadDir")
	}

	var version = Version{
		Name:           filepath.Base(dirname),
		Dirname:        dirname,
		Spec:           new(apistructs.Spec),
		SpecContent:    nil,
		DiceContent:    nil,
		ReadmeContent:  nil,
		SwaggerContent: nil,
	}
	updateTime := baseTime
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get file info")
		}

		if entry.IsDir() {
			continue
		}
		updateTime = latestTime(updateTime, info.ModTime())
		switch {
		case strings.EqualFold(entry.Name(), "spec.yml") || strings.EqualFold(entry.Name(), "spec.yaml"):
			version.SpecContent, err = os.ReadFile(filepath.Join(dirname, entry.Name()))
			if err != nil {
				return nil, errors.Wrap(err, "failed to ReadFile")
			}
			if err = yaml.Unmarshal(version.SpecContent, version.Spec); err != nil {
				return nil, errors.Wrap(err, "failed to parse "+entry.Name())
			}

		case strings.EqualFold(entry.Name(), "dice.yml") || strings.EqualFold(entry.Name(), "dice.yaml"):
			version.DiceContent, _ = os.ReadFile(filepath.Join(dirname, entry.Name()))

		case strings.EqualFold(entry.Name(), "readme.md") || strings.EqualFold(entry.Name(), "readme.markdown"):
			version.ReadmeContent, _ = os.ReadFile(filepath.Join(dirname, entry.Name()))

		case strings.EqualFold(entry.Name(), "swagger.json") || strings.EqualFold(entry.Name(), "swagger.yml") ||
			strings.EqualFold(entry.Name(), "swagger.yaml"):
			version.SwaggerContent, _ = os.ReadFile(filepath.Join(dirname, entry.Name()))
		}
	}

	if version.Spec == nil || len(version.SpecContent) == 0 {
		return nil, errors.Errorf("spec file not found in %s", dirname)
	}
	version.UpdateAt = timestamppb.New(updateTime)

	return &version, nil
}

// LoadExtensions loads all extensions from the repo (contains all versions below)
func LoadExtensions(addr string) *Repo {
	repo := &Repo{
		addr: addr,
	}
	repo.locate(repo.addr, 0)
	return repo
}

// locate Recursively traverse folders
func (repo *Repo) locate(dirname string, deep int) {
	infos, ok := isThereSpecFile(dirname)
	if ok {
		repo.versions = append(repo.versions, dirname)
		return
	}

	for _, cur := range infos {
		// only find path /repoName/actions||addons/extensionsName
		if deep == 1 && cur.Name() != "actions" && cur.Name() != "addons" {
			continue
		}
		repo.locate(filepath.Join(dirname, cur.Name()), deep+1)
	}
}

// isThereSpecFile  check is there have spec.yml
func isThereSpecFile(dirname string) ([]os.FileInfo, bool) {
	var dirs []os.FileInfo
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return nil, false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				return nil, false
			}
			dirs = append(dirs, info)
			continue
		}
		if strings.EqualFold(entry.Name(), "spec.yml") || strings.EqualFold(entry.Name(), "spec.yaml") {
			return nil, true
		}
	}
	return dirs, false
}

// latestTime compare timeA and timeB return the latest time
func latestTime(timeA time.Time, timeB time.Time) time.Time {
	if timeA.After(timeB) {
		return timeA
	}
	return timeB
}
