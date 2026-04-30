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

package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/erda-project/erda/tools/cli/release"
)

func packageArtifact(target publishTarget) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "erda-cli-release-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }

	archivePath := filepath.Join(tmpDir, filepath.Base(release.ArtifactObjectName(target.goos, target.goarch, target.version)))
	switch target.goos {
	case "windows":
		err = packageZipArchive(archivePath, release.ExecutableFileName(target.goos), target.file)
	default:
		err = packageTarGzArchive(archivePath, release.ExecutableFileName(target.goos), target.file)
	}
	if err != nil {
		cleanup()
		return "", nil, err
	}
	return archivePath, cleanup, nil
}

func packageTarGzArchive(archivePath, entryName, sourcePath string) (err error) {
	src, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	dst, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := dst.Close()
		if err == nil {
			err = closeErr
		}
	}()

	gw := gzip.NewWriter(dst)
	defer func() {
		closeErr := gw.Close()
		if err == nil {
			err = closeErr
		}
	}()

	tw := tar.NewWriter(gw)
	defer func() {
		closeErr := tw.Close()
		if err == nil {
			err = closeErr
		}
	}()

	header := &tar.Header{
		Name: entryName,
		Mode: 0o755,
		Size: info.Size(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err = io.Copy(tw, src)
	return err
}

func packageZipArchive(archivePath, entryName, sourcePath string) (err error) {
	src, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	dst, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := dst.Close()
		if err == nil {
			err = closeErr
		}
	}()

	zw := zip.NewWriter(dst)
	defer func() {
		closeErr := zw.Close()
		if err == nil {
			err = closeErr
		}
	}()

	writer, err := zw.Create(entryName)
	if err != nil {
		return err
	}
	_, err = writer.Write(src)
	return err
}
