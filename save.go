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
	"bytes"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/skelterjohn/vendor/vend"
)

func (w *workspace) getOutsidePackages(targets []string) map[string]string {
	os.Setenv("GOPATH", w.gopath(true))

	for _, gopath := range w.gopaths {
		target := "./" + gopath + "/src/..." // filepath.Join() doesn't like a leading dot.
		targets = append(targets, target)
	}

	goListTestArgs := []string{"list", "-e", "-f", "{{range .TestImports}}{{.}}\n{{end}}"}
	goListTestArgs = append(goListTestArgs, targets...)
	// fmt.Printf("%q\n", goListTestArgs)
	var testBuf bytes.Buffer
	cmd := exec.Command("go", goListTestArgs...)
	cmd.Dir = w.root
	cmd.Stdout = &testBuf
	orExit(cmd.Run())
	for _, pkg := range strings.Split(testBuf.String(), "\n") {
		targets = append(targets, pkg)
	}

	goListArgs := []string{"list", "-e", "-f", "{{.ImportPath}}\n{{range .Deps}}{{.}}\n{{end}}"}
	goListArgs = append(goListArgs, targets...)
	// fmt.Printf("%q\n", goListArgs)
	var buf bytes.Buffer
	cmd = exec.Command("go", goListArgs...)
	cmd.Dir = w.root
	cmd.Stdout = &buf
	orExit(cmd.Run())

	goroot := runtime.GOROOT()
	build.Default.GOPATH = w.gopath(true)

	pkgs := map[string]string{}
	for _, pkg := range strings.Split(buf.String(), "\n") {
		if pkg == "" {
			continue
		}
		p, err := build.Import(pkg, w.root, build.FindOnly)
		if err != nil {
			continue
		}
		if x, err := filepath.Rel(goroot, p.Dir); err == nil && !strings.HasPrefix(x, "..") {
			continue
		}
		pkgs[pkg] = p.Dir
	}

	for pkg, dir := range pkgs {
		if !filepath.IsAbs(dir) {
			continue
		}
		if x, err := filepath.Rel(w.root, dir); err == nil && !strings.HasPrefix(x, "..") {
			continue
		}
		pkgs[pkg] = dir
	}

	return pkgs
}

func save(w *workspace, args []string) {

	var targets []string
	godeps := false
	for _, t := range args {
		if t == "--godeps" {
			godeps = true
		} else {
			targets = append(targets, t)
		}
	}

	pkgs := w.getOutsidePackages(targets)

	addonMapping := map[string]string{}
	for pkg, dir := range pkgs {
		destination := filepath.Join(w.vendorRootSrc(), pkg)
		// if it's already in here, vendor will pick it up
		if !filepath.IsAbs(dir) {
			continue
		}
		if x, err := filepath.Rel(w.root, dir); err == nil && !strings.HasPrefix(x, "..") {
			continue
		}
		addonMapping[destination] = dir
	}

	var addons []string
	for destination, dir := range addonMapping {
		addons = append(addons, destination+"="+dir)
	}

	var rgits, rhgs []string
	if godeps {
		for _, dd := range w.importGodeps() {
			rarg := dd.root + "=" + dd.repo + "@" + dd.rev
			switch dd.kind {
			case "git":
				rgits = append(rgits, rarg)
			case "hg":
				rhgs = append(rhgs, rarg)
			default:
				fmt.Fprintf(os.Stderr, "unsupported VCS %q\n", dd.kind)
			}
		}
	}

	ignoreDirs := []string{".git", ".hg", ".gocfg"}
	for _, gopath := range w.gopaths {
		ignoreDirs = append(ignoreDirs,
			filepath.Join(gopath, "pkg"),
			filepath.Join(gopath, "bin"))
	}
	ignored := map[string]bool{}
	for _, dir := range ignoreDirs {
		ignored[dir] = true
	}

	cfgPath := filepath.Join(w.root, ConfigDirName, "vendor.json")

	vend.Save(w.root, cfgPath, addons, rgits, rhgs, ignored, true)
}

func vendor(w *workspace, targets []string) {
	pkgs := w.getOutsidePackages(targets)

	firstGopath := "."
	if len(w.gopaths) != 0 {
		firstGopath = w.gopaths[0]
	}

	for pkg, dir := range pkgs {
		destination := filepath.Join(firstGopath, "src", pkg)
		// if it's already in here, vendor will pick it up
		if !filepath.IsAbs(dir) {
			continue
		}
		if x, err := filepath.Rel(w.root, dir); err == nil && !strings.HasPrefix(x, "..") {
			continue
		}
		if _, err := os.Stat(destination); err == nil {
			continue
		}
		fmt.Println(pkg)
		copyDir(dir, destination)
	}
}

func restore(w *workspace) {
	cfgPath := filepath.Join(w.root, ConfigDirName, "vendor.json")

	vend.Restore(w.root, cfgPath)
}
