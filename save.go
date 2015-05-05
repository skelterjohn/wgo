/*
Copyright 2014 Google Inc. All rights reserved.
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
)

func ensureVendor() {
	var buf bytes.Buffer
	cmd := exec.Command("vendor")
	cmd.Stderr = &buf
	cmd.Run()
	if !strings.HasPrefix(buf.String(), "Usage: vendor") {
		fmt.Fprintln(os.Stderr, "The save/restore functionality uses 'vendor'.")
		fmt.Fprintln(os.Stderr, "To install vendor, 'go get github.com/skelterjohn/vendor'.")
		os.Exit(1)
	}
}

func save(w *workspace) {
	ensureVendor()

	var buf bytes.Buffer
	os.Setenv("GOPATH", w.gopath())
	for _, gopath := range w.gopaths {
		cmd := exec.Command("go", "list", "-f", "{{range .Deps}}{{.}}\n{{end}}", "./"+gopath+"/...")
		cmd.Dir = w.root
		cmd.Stdout = &buf
		orExit(cmd.Run())
	}

	goroot := runtime.GOROOT()
	build.Default.GOPATH = w.gopath()

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

	firstGopath := "."
	if len(w.gopaths) != 0 {
		firstGopath = w.gopaths[0]
	}

	addonMapping := map[string]string{}
	for pkg, dir := range pkgs {
		destination := filepath.Join(firstGopath, "src", pkg)
		// if it's already in here, vendor will pick it up
		if !filepath.IsAbs(dir) {
			continue
		}
		if x, err := filepath.Rel(w.root, dir); err == nil && !strings.HasPrefix(x, "..") {
			continue
		}
		addonMapping[destination] = dir
	}

	var addonArgs []string
	for destination, dir := range addonMapping {
		addonArgs = append(addonArgs, "-a", destination+"="+dir)
	}

	ignoreDirs := []string{".git", ".hg", ".gocfg"}
	for _, gopath := range w.gopaths {
		ignoreDirs = append(ignoreDirs,
			filepath.Join(gopath, "pkg"),
			filepath.Join(gopath, "bin"))
	}
	ignoreDirsStr := strings.Join(ignoreDirs, string(filepath.ListSeparator))
	os.Setenv("VENDOR_IGNORE_DIRS", ignoreDirsStr)

	w.shellOutToVendor(
		append([]string{"wgo", "vendor", "-s"}, addonArgs...))
}

func restore(w *workspace) {
	ensureVendor()

	w.shellOutToVendor([]string{"wgo", "vendor", "-r"})
}