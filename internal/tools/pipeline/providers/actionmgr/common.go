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

package actionmgr

import (
	"context"
	"strings"

	"github.com/erda-project/erda/pkg/common/apis"

	"github.com/erda-project/erda-proto-go/core/dicehub/extension/pb"
)

var (
	defaultVersion = "default"
)

func (s *provider) updateExtensionCache(extension *pb.Extension) {
	// query
	extensionVersions, err := s.ExtensionSvc.QueryExtensionVersions(
		apis.WithInternalClientContext(context.Background(), "pipeline"),
		&pb.ExtensionVersionQueryRequest{
			Name:               extension.Name,
			All:                true,
			YamlFormat:         true,
			OrderByVersionDesc: true,
		})
	if err != nil {
		s.Log.Errorf("failed to query extension version, name: %s, err: %v", extension.Name, err)
		return
	}

	s.Lock()
	defer s.Unlock()

	// delete from defaultActionsCache by action name firstly, because maybe not have default versions in queried result
	delete(s.defaultActionsCache, extension.Name)
	// update
	for _, extensionVersion := range extensionVersions.Data {
		s.actionsCache[makeActionNameVersion(extensionVersion.Name, extensionVersion.Version)] = extensionVersion
		if extensionVersion.IsDefault {
			s.defaultActionsCache[extension.Name] = extensionVersion
		}
	}
	// if not found the default version, set the first public version as default
	if _, ok := s.defaultActionsCache[extension.Name]; !ok && len(extensionVersions.Data) > 0 {
		for _, extensionVersion := range extensionVersions.Data {
			if extensionVersion.Public {
				s.defaultActionsCache[extension.Name] = extensionVersion
				break
			}
		}
	}
}

// getOrUpdateExtensionFromCache get the fitted extension from the cache
// if not exist, try to update the cache by the given extension name
func (s *provider) getOrUpdateExtensionFromCache(nameVersion string) (action *pb.ExtensionVersion, found bool) {
	splits := strings.SplitN(nameVersion, "@", 2)
	name := splits[0]
	version := ""
	if len(splits) > 1 {
		version = splits[1]
	}
	if version == "" {
		s.Lock()
		action, found = s.defaultActionsCache[name]
		s.Unlock()
		if !found {
			newAction, err := s.ExtensionSvc.GetExtension(name, defaultVersion, true)
			if err != nil {
				found = false
				return
			}
			s.Lock()
			s.defaultActionsCache[name] = newAction
			s.Unlock()
			return newAction, true
		}
		return
	}
	s.Lock()
	action, found = s.actionsCache[nameVersion]
	s.Unlock()
	if !found {
		newAction, err := s.ExtensionSvc.GetExtension(name, version, true)
		if err != nil {
			found = false
			return
		}
		s.Lock()
		s.actionsCache[nameVersion] = newAction
		s.Unlock()
		return newAction, true
	}
	return
}

func getActionNameVersion(nameVersion string) (string, string) {
	splits := strings.SplitN(nameVersion, "@", 2)
	name := splits[0]
	version := ""
	if len(splits) > 1 {
		version = splits[1]
	}
	if version == defaultVersion {
		version = ""
	}
	return name, version
}

func makeActionNameVersion(name, version string) string {
	if len(version) == 0 {
		return name
	}
	return name + "@" + version
}
