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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	ConfigDirName = ".gocfg"
	Debug         = false
)

func init() {
	log.SetFlags(0)
	if !Debug {
		log.SetOutput(ioutil.Discard)
	}
}

func orExit(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}

func main() {
	if len(os.Args) == 1 {
		w, err := getCurrentWorkspace()
		orExit(err)
		w.shellOutToGo(os.Args)
	}
	var err error
	switch os.Args[1] {
	case "init":
		err = initWgo(os.Args[2:])
	case "clone":
		err = clone(os.Args[2:])
	case "restore":
		w, err := getCurrentWorkspace()
		orExit(err)
		w.shellOutToVendor([]string{"wgo", "vendor", "-r"})
	case "save":
		w, err := getCurrentWorkspace()
		orExit(err)
		save(w)
	default:
		w, err := getCurrentWorkspace()
		orExit(err)
		w.shellOutToGo(os.Args)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func initWgo(args []string) error {
	w, err := getCurrentWorkspace()
	if err == nil {
		return fmt.Errorf("%q is already a workspace", w.root)
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	fi, err := os.Stat(wd)
	if err != nil {
		return err
	}

	err = os.Mkdir(filepath.Join(wd, ConfigDirName), fi.Mode())
	if err != nil {
		return err
	}

	// ignore error, src might already be there.
	_ = os.Mkdir(filepath.Join(wd, "src"), fi.Mode())

	return nil
}

func clone(args []string) error {
	return nil
}
