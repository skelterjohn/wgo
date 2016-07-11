/*
Copyright 2016 Google Inc. All rights reserved.
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
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Purge directories in the indicated gopaths if they to not contain source
// referenced from a non-indicated gopath.
func purge(w *workspace, args []string) {
	confirmed := false
	var gopaths []string
	for _, a := range args {
		if a == "--confirm" {
			confirmed = true
			continue
		}
		gopaths = append(gopaths, a)
	}
	bctx := build.Default
	bctx.GOPATH = w.Gopath(false)

	if len(gopaths) == 0 {
		gopaths = w.Gopaths[:1] // By default, the first one is vendor/.
	}
	if len(gopaths) == 0 {
		fmt.Fprintf(os.Stderr, "must purge at least one GOPATH\n")
		os.Exit(1)
	}
	wgps := map[string]bool{}
	for _, wpg := range w.Gopaths {
		wgps[wpg] = true
	}
	pgs := map[string]bool{}
	for _, pg := range gopaths {
		pgs[pg] = true
		if !wgps[pg] {
			fmt.Fprintf(os.Stderr, "unknown GOPATH %q", pg)
			os.Exit(1)
		}
	}
	if len(gopaths) == len(w.Gopaths) {
		fmt.Fprintf(os.Stderr, "cannot purge all GOPATHs; try 'rm -r' instead\n")
		os.Exit(1)
	}

	// Collect a set of safe directories that will not get purged.
	safeDirs := []string{}
	// Keep them in a map too, to protect against loops.
	safeDirsAll := map[string]bool{}
	// Start by adding all directories in the non-purged gopaths.
	for _, wpg := range w.Gopaths {
		if pgs[wpg] {
			// skip the purged ones
			continue
		}
		// Only deal with abs paths from here on out, to make checking easier.
		if !filepath.IsAbs(wpg) {
			wpg = filepath.Join(w.Root, wpg)
		}
		filepath.Walk(filepath.Join(wpg, "src"), func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				safeDirs = append(safeDirs, path)
				safeDirsAll[path] = true
			}
			return nil
		})
	}

	// Go through each safe dir and add its subsafedirs to the end of the list.
	for i := 0; i < len(safeDirs); i++ {
		deps, err := getDepDirs(bctx, safeDirs[i])
		if err != nil {
			fmt.Fprintf(os.Stderr, "problem inspecting %s: %v\n", safeDirs[i], err)
			os.Exit(1)
		}
		for _, d := range deps {
			if safeDirsAll[d] {
				// cut off cycles
				continue
			}
			safeDirs = append(safeDirs, d)
			safeDirsAll[d] = true
		}
	}

	// Expand the list of safe dirs to be all parents of safe dirs, to make checking easier later.
	for dir := range safeDirsAll {
		for _, parent := range getAllParents(dir) {
			safeDirsAll[parent] = true
		}
	}

	// Armed with a list of safe dirs, go through gopaths and purge those that aren't safe.
	dirsToPurge := map[string]bool{}
	for _, pg := range gopaths {
		if !filepath.IsAbs(pg) {
			pg = filepath.Join(w.Root, pg)
		}
		filepath.Walk(filepath.Join(pg, "src"), func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				return nil
			}
			// If this directory is safe, or is the parent of somethinge safe, we keep it.
			// We stored all the parents of the safe directories so we only need to do a single check here.
			if safeDirsAll[path] {
				return nil
			}
			// Not safe, will be purged.
			dirsToPurge[path] = true
			return nil
		})
	}

	for d := range dirsToPurge {
		// If we're already removing a parent, forget about this directory.
		pds := getAllParents(d)
		for _, pd := range pds {
			if dirsToPurge[pd] {
				dirsToPurge[d] = false
			} else {
				//fmt.Println("no parent", pd, d)
			}
		}
	}

	sortedPurge := []string{}
	for d, rm := range dirsToPurge {
		if rm {
			rd, err := filepath.Rel(w.Root, d)
			if err != nil {
				fmt.Fprintf(os.Stderr, "problem with relative path for %q: %v\n", d, err)
				os.Exit(1)
			}
			if strings.HasPrefix(rd, "..") {
				fmt.Fprintf(os.Stderr, "skipping path outside of workspace %q", d)
				continue
			}
			sortedPurge = append(sortedPurge, rd)
		}
	}
	sort.Strings(sortedPurge)

	if !confirmed {
		fmt.Fprintln(os.Stderr, "Directories containing no imported source:")
		for _, d := range sortedPurge {
			fmt.Println(d)
		}
		fmt.Fprintln(os.Stderr, "To delete the listed directories, run this command again with '--confirm'.")
		os.Exit(0)
	}

	for _, d := range sortedPurge {
		if err := os.RemoveAll(d); err != nil {
			fmt.Println("Error removing %q: %v\n", d, err)
		}
	}

}

func getDepDirs(bctx build.Context, dir string) ([]string, error) {
	pkg, err := bctx.ImportDir(dir, 0)
	if err != nil {
		if _, ok := err.(*build.NoGoError); !ok {
			return nil, err
		}
	}
	depDirs := []string{}
	for _, imp := range pkg.Imports {
		depPkg, err := bctx.Import(imp, dir, 0)
		if err == nil {
			depDirs = append(depDirs, depPkg.Dir)
		}
	}
	return depDirs, nil
}

func getAllParents(dir string) []string {
	var parents []string
	for {
		parent, _ := filepath.Split(dir)
		parent = filepath.Clean(parent)
		if parent == dir {
			return parents
		}
		parents = append(parents, parent)
		dir = parent
	}
}
