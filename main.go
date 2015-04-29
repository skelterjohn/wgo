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
	"strings"
)

const (
	ConfigDirName = ".gocfg"
	Debug         = false
)

const usageMessage = `wgo is a tool for managing Go workspaces.

usage: wgo init [--set-primary=PRIMARY_GOPATH] [ADDITIONAL_GOPATH+]
       wgo save
       wgo restore
       wgo <go command>
`

func usage() {
	fmt.Print(usageMessage)
	os.Exit(1)
}

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
		fmt.Println(usageMessage)
		w.shellOutToGo(os.Args)
	}
	var err error
	switch os.Args[1] {
	case "init":
		if err = initWgo(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
	case "restore":
		if len(os.Args) != 2 {
			usage()
		}
		w, err := getCurrentWorkspace()
		orExit(err)
		restore(w)
	case "save":
		if len(os.Args) != 2 {
			usage()
		}
		w, err := getCurrentWorkspace()
		orExit(err)
		save(w)
	default:
		w, err := getCurrentWorkspace()
		orExit(err)
		w.shellOutToGo(os.Args)
	}
}

func initWgo(args []string) error {
	for _, arg := range args {
		if arg != "--set-primary" && strings.HasPrefix(arg, "-") {
			fmt.Fprintf(os.Stderr, "unrecognized flag: %s\n\n", arg)
			usage()
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	w, err := getCurrentWorkspace()
	if err == nil && (w.root != wd || len(args) == 0) {
		return fmt.Errorf("%q is already a workspace", w.root)
	}

	fi, err := os.Stat(wd)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(wd, ConfigDirName), fi.Mode()); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(wd, "src"), fi.Mode()); err != nil {
		return err
	}

	if w, err = getCurrentWorkspace(); err != nil {
		return err
	}

	gopathsPath := filepath.Join(wd, ConfigDirName, "gopaths")
	// if there is no gopaths yet, stick '.' in there.
	if _, err := os.Stat(gopathsPath); err != nil {
		args = append([]string{"."}, args...)
	}

	var fout *os.File
	if fout, err = os.Create(gopathsPath); err != nil {
		return err
	}
	defer fout.Close()

	checkGopath := func(gopath string) {
		if filepath.IsAbs(gopath) {
			fmt.Fprintf(os.Stderr, "%q is not a relative path\n", gopath)
		}
	}

	var gopathArgs []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--set-primary" {
			if i+1 >= len(args) {
				usage()
			}
			checkGopath(args[i+1])
			fmt.Fprintln(fout, args[i+1])
			i++
			continue
		}

		checkGopath(args[i])

		if strings.HasPrefix(args[i], "--set-primary=") {
			fmt.Fprintln(fout, args[i][len("--set-primary="):])
			continue
		}

		gopathArgs = append(gopathArgs, args[i])
	}

	for _, gopath := range w.gopaths {
		fmt.Fprintln(fout, gopath)
	}
	for _, gopath := range gopathArgs {
		fmt.Fprintln(fout, gopath)
	}

	return nil
}

func clone(args []string) error {
	return nil
}
