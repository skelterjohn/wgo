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
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/skelterjohn/wgo/workspaces"
)

type workspace struct {
	workspaces.Workspace
}

func getCurrentWorkspace() (*workspace, error) {
	if w, err := workspaces.GetCurrentWorkspace(); err != nil {
		return nil, err
	} else {
		return &workspace{*w}, nil
	}
}

func getWorkspace(start string) (*workspace, error) {
	if w, err := workspaces.GetWorkspace(start); err != nil {
		return nil, err
	} else {
		return &workspace{*w}, nil
	}
}

func guessGoCommand(args []string) string {
	if len(args) < 1 {
		return ""
	}
	return args[1]
}

func (w *workspace) shellOutToGo(args []string) {
	// we want to fetch new code directly into the workspace, for convenience
	gopath := w.Gopath(guessGoCommand(args) != "get")
	os.Setenv("GOPATH", gopath)
	log.Printf("using GOPATH=%s", gopath)
	shellOutToGo(args)
}

func (w *workspace) vendorRootSrc() string {
	firstGopath := "."
	if len(w.Gopaths) != 0 {
		firstGopath = w.Gopaths[0]
	}
	return filepath.Join(firstGopath, "src")
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
