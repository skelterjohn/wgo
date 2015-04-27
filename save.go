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
	gopath := os.Getenv("GOPATH")
	newgopath := fmt.Sprintf("%s%c%s", w.root, filepath.ListSeparator, gopath)
	os.Setenv("GOPATH", newgopath)
	cmd := exec.Command("go", "list", "-f", "{{range .Deps}}{{.}}\n{{end}}", "./src/...")
	cmd.Dir = w.root
	cmd.Stdout = &buf
	orExit(cmd.Run())

	goroot := runtime.GOROOT()
	build.Default.GOPATH = newgopath

	pkgs := map[string]string{}
	for _, pkg := range strings.Split(buf.String(), "\n") {
		if pkg == "" {
			continue
		}
		p, err := build.Import(pkg, w.root, build.FindOnly)
		if err != nil {
			continue
		}
		if strings.HasPrefix(p.Dir, goroot+"/") {
			continue
		}
		pkgs[pkg] = p.Dir
	}

	var addonArgs []string

	for pkg, dir := range pkgs {
		addonArgs = append(addonArgs, "-a", filepath.Join("src", pkg)+"="+dir)

	}
	w.shellOutToVendor(
		append([]string{"wgo", "vendor", "-s"}, addonArgs...))
}

func restore(w *workspace) {
	ensureVendor()

	w.shellOutToVendor([]string{"wgo", "vendor", "-r"})
}
