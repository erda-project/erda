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

package source

import (
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type ExtensionSource interface {
	match(addr string) bool
	add(addr string) error
	remove(addr string) error
	start() error
}

type GitExtensionSource struct {
	GitCloneAddr map[string]string
}

func (g GitExtensionSource) match(addr string) bool {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") || strings.HasPrefix(addr, "git@") {
		return true
	}
	return false
}

func (g GitExtensionSource) add(addr string) error {

	if g.GitCloneAddr[addr] == "" {
		dir, err := ioutil.TempDir(os.TempDir(), "*")
		if err != nil {
			logrus.Error("gitExtensionSource tempDir error %v", err)
			return err
		}

		command := exec.Command("sh", "-c", "git clone "+addr)
		command.Dir = dir
		output, err := command.CombinedOutput()
		if err != nil {
			logrus.Errorf("git clone extensions address stderr %v", string(output))
			return err
		}
		g.GitCloneAddr[addr] = dir
	}

	return AddSyncExtension(g.GitCloneAddr[addr])
}

func (g GitExtensionSource) remove(addr string) error {
	err := RemoveSyncExtension(addr)
	if err != nil {
		return err
	}


}

type FileExtensionSource struct {
	watcher *fsnotify.Watcher
	watchAddrList map[string]string
}

func (f FileExtensionSource) match(addr string) bool {
	s, err := os.Stat(addr)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func (f FileExtensionSource) add(addr string) error {
	if f.watchAddrList[addr] == "" {
		f.watchAddrList[addr] = addr
	}

	err := f.watcher.Add(addr)
	if err != nil {
		return err
	}
	return nil
}

func (f FileExtensionSource) remove(addr string) error {
	return f.watcher.Remove(addr)
}



var extensionSources []ExtensionSource

func RegisterExtensionSource(source ExtensionSource)  {
	extensionSources = append(extensionSources, source)
}

func AddSyncExtension(addr string) error {
	for _, source := range extensionSources {
		if source.match(addr) {
			err := source.add(addr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func RemoveSyncExtension(addr string) error {
	for _, source := range extensionSources {
		if source.match(addr) {
			err := source.remove(addr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}



