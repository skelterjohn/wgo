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
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type workspace struct {
	root string
}

func getCurrentWorkspace() (*workspace, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return getWorkspace(wd)
}

func getWorkspace(start string) (*workspace, error) {
	goCfgPath := filepath.Join(start, ConfigDirName)

	fi, err := os.Stat(goCfgPath)
	if err == nil {
		if fi.IsDir() {
			return &workspace{
				root: start,
			}, nil
		}
	}

	if rune(start[len(start)-1]) == filepath.Separator {
		start = start[:len(start)-1]
	}
	dir, _ := filepath.Split(start)
	if dir == "" {
		err = errors.New("no workspace")
		return nil, err
	}

	return getWorkspace(dir)
}

func (w *workspace) shellOutToGo(args []string) {
	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", fmt.Sprintf("%s%c%s", w.root, filepath.ListSeparator, gopath))

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

func (w *workspace) shellOutToVendor(args []string) {
	vendorJson := filepath.Join(w.root, ConfigDirName, "vendor.json")
	vendorArgs := append(args[2:], vendorJson)

	log.Printf("forking to vendor: %q", vendorArgs)
	cmd := exec.Command("vendor", vendorArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = w.root
	err := cmd.Run()
	if _, ok := err.(*exec.ExitError); ok {
		os.Exit(1)
	}
	os.Exit(0)
}
