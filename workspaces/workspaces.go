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

package workspaces

import (
	"bufio"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	ConfigDirName = ".gocfg"
)

type Workspace struct {
	Root    string
	Gopaths []string
}

func GetCurrentWorkspace() (*Workspace, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return GetWorkspace(wd)
}

func GetWorkspace(start string) (*Workspace, error) {
	goCfgPath := filepath.Join(start, ConfigDirName)

	if fi, err := os.Stat(goCfgPath); err == nil && fi.IsDir() {
		w := &Workspace{
			Root: start,
		}
		if cfgFile, err := os.Open(filepath.Join(goCfgPath, "gopaths")); err == nil {
			sc := bufio.NewScanner(cfgFile)
			for sc.Scan() {
				gopath := sc.Text()
				w.Gopaths = append(w.Gopaths, strings.TrimSpace(gopath))
			}
		}
		return w, nil
	}

	if rune(start[len(start)-1]) == filepath.Separator {
		start = start[:len(start)-1]
	}
	dir, _ := filepath.Split(start)
	if dir == start {
		return nil, errors.New("no workspace")
	}

	return GetWorkspace(dir)
}

func (w *Workspace) Gopath(external bool) string {
	var oldgopath string
	if external {
		oldgopath = os.Getenv("GOPATH")
	}
	var absGoPaths []string
	for _, gopath := range w.Gopaths {
		absGoPaths = append(absGoPaths, filepath.Join(w.Root, gopath))
	}
	newgopath := strings.Join(absGoPaths, string(filepath.ListSeparator))
	newgopath = strings.Join([]string{newgopath, oldgopath}, string(filepath.ListSeparator))
	return newgopath
}

func guessGoCommand(args []string) string {
	if len(args) < 1 {
		return ""
	}
	return args[1]
}

func (w *Workspace) ShellOutToGo(args []string) {
	// we want to fetch new code directly into the workspace, for convenience
	gopath := w.Gopath(guessGoCommand(args) != "get")
	os.Setenv("GOPATH", gopath)
	log.Printf("using GOPATH=%s", gopath)
	shellOutToGo(args)
}

func shellOutToGo(args []string) {
	log.Printf("forking to go: %q", args[1:])
	cmd := exec.Command("go", args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if _, ok := err.(*exec.ExitError); ok {
		os.Exit(1)
	}
	os.Exit(0)
}
