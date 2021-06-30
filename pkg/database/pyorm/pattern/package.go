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

package pattern

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	packageName = ".erda_migration_in_py"
)
const EntryFilename = "entry.py"

type TextFile interface {
	GetName() string
	GetData() []byte
}

type Package struct {
	DeveloperScript TextFile
	Requirements    TextFile
	Settings        Settings
	Entrypoint      Entrypoint
	Commit          bool
}

func (p *Package) Make() (err error) {
	if err = os.Mkdir(".erda_migration_in_py", 0644); err != nil {
		return errors.Wrap(err, "failed to make temp python package for migration")
	}
	defer func() {
		if err != nil {
			_ = p.Remove()
		}
	}()

	msg := "failed to make python file for migration"

	if err = p.writeDeveloperScript(); err != nil {
		return errors.Wrap(err, msg)
	}

	if err = p.writeEntrypoint(); err != nil {
		return errors.Wrap(err, msg)
	}

	if err = p.writeRequirements(); err != nil {
		return errors.Wrap(err, msg)
	}

	return nil
}

func (p *Package) writeDeveloperScript() error {
	filename := "erda_" + filepath.Base(p.DeveloperScript.GetName())
	filename = filepath.Join(packageName, filename)

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = GenSettings(f, p.Settings); err != nil {
		return err
	}
	if _, err = f.Write(p.DeveloperScript.GetData()); err != nil {
		return err
	}
	return nil
}

func (p *Package) writeEntrypoint() error {
	f, err := os.OpenFile(filepath.Join(packageName, EntryFilename), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return GenEntrypoint(f, p.Entrypoint, p.Commit)
}

func (p *Package) writeRequirements() error {
	f, err := os.OpenFile(filepath.Join(packageName, RequirementsFilename), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(p.Requirements.GetData())
	return err
}

func (p *Package) Remove() error {
	return os.RemoveAll(packageName)
}

func (p *Package) Run() error {
	cmd := exec.Command("pip", "install", "-r", filepath.Join(packageName, RequirementsFilename), "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stdout, os.Stderr)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}

	cmd = exec.Command("python", filepath.Join(packageName, EntryFilename))
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stdout, os.Stderr)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
