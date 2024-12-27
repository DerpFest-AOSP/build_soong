// Copyright 2024 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elf

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func UpdateBuildIdDir(path string) error {
	path = filepath.Clean(path)
	buildIdPath := path + "/.build-id"

	// Collect the list of files and build-id symlinks. If the symlinks are
	// up to date (newer than the symbol files), there is nothing to do.
	var buildIdFiles, symbolFiles []string
	var buildIdMtime, symbolsMtime time.Time
	filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if entry == nil || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		mtime := info.ModTime()
		if strings.HasPrefix(path, buildIdPath) {
			if buildIdMtime.Compare(mtime) < 0 {
				buildIdMtime = mtime
			}
			buildIdFiles = append(buildIdFiles, path)
		} else {
			if symbolsMtime.Compare(mtime) < 0 {
				symbolsMtime = mtime
			}
			symbolFiles = append(symbolFiles, path)
		}
		return nil
	})
	if symbolsMtime.Compare(buildIdMtime) < 0 {
		return nil
	}

	// Collect build-id -> file mapping from ELF files in the symbols directory.
	concurrency := 8
	done := make(chan error)
	buildIdToFile := make(map[string]string)
	var mu sync.Mutex
	for i := 0; i != concurrency; i++ {
		go func(paths []string) {
			for _, path := range paths {
				id, err := Identifier(path, true)
				if err != nil {
					done <- err
					return
				}
				if id == "" {
					continue
				}
				mu.Lock()
				oldPath := buildIdToFile[id]
				if oldPath == "" || oldPath > path {
					buildIdToFile[id] = path
				}
				mu.Unlock()
			}
			done <- nil
		}(symbolFiles[len(symbolFiles)*i/concurrency : len(symbolFiles)*(i+1)/concurrency])
	}

	// Collect previously generated build-id -> file mapping from the .build-id directory.
	// We will use this for incremental updates. If we see anything in the .build-id
	// directory that we did not expect, we'll delete it and start over.
	prevBuildIdToFile := make(map[string]string)
out:
	for _, buildIdFile := range buildIdFiles {
		if !strings.HasSuffix(buildIdFile, ".debug") {
			prevBuildIdToFile = nil
			break
		}
		buildId := buildIdFile[len(buildIdPath)+1 : len(buildIdFile)-6]
		for i, ch := range buildId {
			if i == 2 {
				if ch != '/' {
					prevBuildIdToFile = nil
					break out
				}
			} else {
				if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
					prevBuildIdToFile = nil
					break out
				}
			}
		}
		target, err := os.Readlink(buildIdFile)
		if err != nil || !strings.HasPrefix(target, "../../") {
			prevBuildIdToFile = nil
			break
		}
		prevBuildIdToFile[buildId[0:2]+buildId[3:]] = path + target[5:]
	}
	if prevBuildIdToFile == nil {
		err := os.RemoveAll(buildIdPath)
		if err != nil {
			return err
		}
		prevBuildIdToFile = make(map[string]string)
	}

	// Wait for build-id collection from ELF files to finish.
	for i := 0; i != concurrency; i++ {
		err := <-done
		if err != nil {
			return err
		}
	}

	// Delete old symlinks.
	for id, _ := range prevBuildIdToFile {
		if buildIdToFile[id] == "" {
			symlinkDir := buildIdPath + "/" + id[:2]
			symlinkPath := symlinkDir + "/" + id[2:] + ".debug"
			if err := os.Remove(symlinkPath); err != nil {
				return err
			}
		}
	}

	// Add new symlinks and update changed symlinks.
	for id, path := range buildIdToFile {
		prevPath := prevBuildIdToFile[id]
		if prevPath == path {
			continue
		}
		symlinkDir := buildIdPath + "/" + id[:2]
		symlinkPath := symlinkDir + "/" + id[2:] + ".debug"
		if prevPath == "" {
			if err := os.MkdirAll(symlinkDir, 0755); err != nil {
				return err
			}
		} else {
			if err := os.Remove(symlinkPath); err != nil {
				return err
			}
		}

		target, err := filepath.Rel(symlinkDir, path)
		if err != nil {
			return err
		}
		if err := os.Symlink(target, symlinkPath); err != nil {
			return err
		}
	}
	return nil
}
