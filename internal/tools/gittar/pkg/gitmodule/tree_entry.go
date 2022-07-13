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

//go:build !codeanalysis
// +build !codeanalysis

package gitmodule

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"sort"

	git "github.com/libgit2/git2go/v33"
)

type EntryMode string

// There are only a few file modes in Git. They look like unix file modes, but they can only be
// one of these.
const (
	ENTRY_MODE_BLOB    EntryMode = "0100644"
	ENTRY_MODE_EXEC    EntryMode = "0100755"
	ENTRY_MODE_SYMLINK EntryMode = "0120000"
	ENTRY_MODE_COMMIT  EntryMode = "0160000"
	ENTRY_MODE_TREE    EntryMode = "0040000"
)

type TreeEntry struct {
	ID        string     `json:"id"`
	Type      ObjectType `json:"type"`
	Url       string     `json:"url,omitempty"`
	Name      string     `json:"name"`
	Mode      EntryMode  `json:"mode"`
	PtrTree   *Tree      `json:"-"`
	path      string
	repo      *Repository
	Commit    *Commit `json:"commit"`
	EntrySize int64   `json:"size"`
	sized     bool    `json:"sized"`
}

func (te *TreeEntry) ListEntries() (Entries, error) {
	tree := &Tree{
		ID:   te.ID,
		Repo: te.repo,
	}
	return tree.innerListEntries()
}

func (te *TreeEntry) Size() int64 {
	if te.IsDir() {
		return 0
	} else if te.sized {
		return te.EntrySize
	}
	err := Setting.ObjectSizeCache.Get(te.ID, &te.EntrySize)
	if err == nil {
		te.sized = true
		return te.EntrySize
	}
	repository, err := te.repo.GetRawRepo()
	if err != nil {
		return 0
	}
	oid, err := git.NewOid(te.ID)
	if err != nil {
		return 0
	}
	blob, err := repository.LookupBlob(oid)

	if err != nil {
		return 0
	}

	te.sized = true
	te.EntrySize = blob.Size()
	Setting.ObjectSizeCache.Set(te.ID, te.EntrySize)
	return te.EntrySize
}

func (te *TreeEntry) IsSubModule() bool {
	return te.Mode == ENTRY_MODE_COMMIT
}

func (te *TreeEntry) IsDir() bool {
	return te.Mode == ENTRY_MODE_TREE
}

func (te *TreeEntry) IsLink() bool {
	return te.Mode == ENTRY_MODE_SYMLINK
}

func (te *TreeEntry) Blob() *Blob {
	return &Blob{
		repo:      te.PtrTree.Repo,
		TreeEntry: te,
	}
}

type Entries []*TreeEntry

var sorter = []func(t1, t2 *TreeEntry) bool{
	func(t1, t2 *TreeEntry) bool {
		return (t1.IsDir() || t1.IsSubModule()) && !t2.IsDir() && !t2.IsSubModule()
	},
	func(t1, t2 *TreeEntry) bool {
		return t1.Name < t2.Name
	},
}

func (tes Entries) Len() int      { return len(tes) }
func (tes Entries) Swap(i, j int) { tes[i], tes[j] = tes[j], tes[i] }
func (tes Entries) Less(i, j int) bool {
	t1, t2 := tes[i], tes[j]
	var k int
	for k = 0; k < len(sorter)-1; k++ {
		sort := sorter[k]
		switch {
		case sort(t1, t2):
			return true
		case sort(t2, t1):
			return false
		}
	}
	return sorter[k](t1, t2)
}

func (tes Entries) Sort() {
	sort.Sort(tes)
}

var defaultConcurrency = runtime.NumCPU()

type commitInfo struct {
	entryName string
	infos     []interface{}
	err       error
}

// GetCommitsInfo takes advantages of concurrency to speed up getting information
// of all commits that are corresponding to these entries. This method will automatically
// choose the right number of goroutine (concurrency) to use related of the host CPU.
func (tes Entries) GetCommitsInfo(commit *Commit, treePath string) ([][]interface{}, error) {
	return tes.GetCommitsInfoWithCustomConcurrency(commit, treePath, 0)
}

// GetCommitsInfoWithCustomConcurrency takes advantages of concurrency to speed up getting information
// of all commits that are corresponding to these entries. If the given maxConcurrency is negative or
// equal to zero:  the right number of goroutine (concurrency) to use will be choosen related of the
// host CPU.
func (tes Entries) GetCommitsInfoWithCustomConcurrency(commit *Commit, treePath string, maxConcurrency int) ([][]interface{}, error) {
	if len(tes) == 0 {
		return nil, nil
	}

	if maxConcurrency <= 0 {
		maxConcurrency = defaultConcurrency
	}

	// Length of taskChan determines how many goroutines (subprocesses) can run at the same time.
	// The length of revChan should be same as taskChan so goroutines whoever finished job can
	// exit as early as possible, only store data inside channel.
	taskChan := make(chan bool, maxConcurrency)
	revChan := make(chan commitInfo, maxConcurrency)
	doneChan := make(chan error)

	// Receive loop will exit when it collects same number of data pieces as tree entries.
	// It notifies doneChan before exits or notify early with possible error.
	infoMap := make(map[string][]interface{}, len(tes))
	go func() {
		i := 0
		for info := range revChan {
			if info.err != nil {
				doneChan <- info.err
				return
			}

			infoMap[info.entryName] = info.infos
			i++
			if i == len(tes) {
				break
			}
		}
		doneChan <- nil
	}()

	for i := range tes {
		// When taskChan is idle (or has empty slots), put operation will not block.
		// However when taskChan is full, code will block and wait any running goroutines to finish.
		taskChan <- true

		if tes[i].Type != OBJECT_COMMIT {
			go func(i int) {
				cinfo := commitInfo{entryName: tes[i].Name}
				c, err := commit.GetCommitByPath(filepath.Join(treePath, tes[i].Name))
				if err != nil {
					cinfo.err = fmt.Errorf("GetCommitByPath (%s/%s): %v", treePath, tes[i].Name, err)
				} else {
					cinfo.infos = []interface{}{tes[i], c}
				}
				revChan <- cinfo
				<-taskChan // Clear one slot from taskChan to allow new goroutines to start.
			}(i)
			continue
		}

		// Handle submodule
		go func(i int) {
			cinfo := commitInfo{entryName: tes[i].Name}
			sm, err := commit.GetSubModule(path.Join(treePath, tes[i].Name))
			if err != nil && !IsErrNotExist(err) {
				cinfo.err = fmt.Errorf("GetSubModule (%s/%s): %v", treePath, tes[i].Name, err)
				revChan <- cinfo
				return
			}

			smURL := ""
			if sm != nil {
				smURL = sm.URL
			}

			c, err := commit.GetCommitByPath(filepath.Join(treePath, tes[i].Name))
			if err != nil {
				cinfo.err = fmt.Errorf("GetCommitByPath (%s/%s): %v", treePath, tes[i].Name, err)
			} else {
				cinfo.infos = []interface{}{tes[i], NewSubModuleFile(c, smURL, tes[i].ID)}
			}
			revChan <- cinfo
			<-taskChan
		}(i)
	}

	if err := <-doneChan; err != nil {
		return nil, err
	}

	commitsInfo := make([][]interface{}, len(tes))
	for i := 0; i < len(tes); i++ {
		commitsInfo[i] = infoMap[tes[i].Name]
	}
	return commitsInfo, nil
}
