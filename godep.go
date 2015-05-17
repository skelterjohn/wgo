/*
Copyright 2015 Google Inc. All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/vcs"
)

type Godeps struct {
	Deps []Dependency
}

type Dependency struct {
	ImportPath string
	Rev        string
}

func (w *workspace) importGodeps() map[string]dirDep {
	dirGs := map[string]Godeps{}
	scanDir := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		if g, err := loadGodepsConfig(path); err == nil {
			dirGs[path] = g
		}
		return nil
	}
	filepath.Walk(w.root, scanDir)
	return w.mergeGodeps(dirGs)
}

func loadGodepsConfig(dir string) (Godeps, error) {
	var g Godeps

	configPath := filepath.Join(dir, "Godeps", "Godeps.json")
	fin, err := os.Open(configPath)
	if err != nil {
		return g, err
	}
	return g, json.NewDecoder(fin).Decode(&g)
}

type dirDep struct {
	srcDir string
	repo   string
	rev    string
	root   string
	kind   string
}

// mergeGodeps will get one master list of revs.
func (w *workspace) mergeGodeps(dirGs map[string]Godeps) map[string]dirDep {
	roots := map[string]dirDep{}
	for dir, g := range dirGs {
		for _, dep := range g.Deps {
			repoRoot, err := vcs.RepoRootForImportPath(dep.ImportPath, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "for %q: %s\n", dep.ImportPath, err)
				continue
			}
			dd := dirDep{
				srcDir: dir,
				rev:    dep.Rev,
				repo:   repoRoot.Repo,
				root:   filepath.Join(w.vendorRootSrc(), repoRoot.Root),
				kind:   repoRoot.VCS.Cmd,
			}
			if orig, ok := roots[repoRoot.Repo]; ok {
				if orig != dd {
					fmt.Fprintf(os.Stderr, "conflict for %q: godeps in %q and %q do not match\n",
						dd.repo, orig.srcDir, dd.srcDir)
					continue
				}
			} else {
				roots[repoRoot.Root] = dd
			}
		}
	}

	// clear out nested
	var rootDirs sort.StringSlice
	for _, dd := range roots {
		rootDirs = append(rootDirs, dd.root)
	}
	sort.Sort(rootDirs)

	var last string
	for _, r := range rootDirs {
		if last != "" && strings.HasPrefix(r, last) {
			delete(roots, r)
			fmt.Fprintf(os.Stderr, "ignoring %q which is managed in %q\n", r, last)
		} else {
			last = r + string(filepath.Separator)
		}
	}

	return roots
}
