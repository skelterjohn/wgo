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

/*
The wgo tool is a small wrapper around the go tool. It adds the concept
of a workspace, in addition to that of GOPATH, and several new commands
to manage that workspace.
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
)

var Debug bool

const getFlag = "--go-get"

var usageMessage = fmt.Sprintf(`wgo is a tool for managing Go workspaces.

usage: wgo init [%s=GO_GET_GOPATH] [ADDITIONAL_GOPATH+]
       wgo import
       wgo save
       wgo restore

       wgo <go command>  # run a go command with the workspace's gopaths
`, getFlag)

func usage() {
	fmt.Print(usageMessage)
	os.Exit(1)
}

func init() {
	log.SetFlags(0)
	Debug = os.Getenv("WGO_DEBUG") == "1"
	if !Debug {
		log.SetOutput(ioutil.Discard)
	}
	log.Println("debug mode on")
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
		fmt.Println(usageMessage)
		shellOutToGo(os.Args)
	}
	var err error
	switch os.Args[1] {
	case "init":
		if err = initWgo(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
	case "import":
		if len(os.Args) != 2 {
			usage()
		}
		w, err := getCurrentWorkspace()
		orExit(err)
		importPkgs(w)
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
		if arg != getFlag && strings.HasPrefix(arg, "-") {
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

	if len(args) == 0 {
		args = []string{getFlag, "third_party"}
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

	goGetDir := ""

	var gopathArgs []string
	for i := 0; i < len(args); i++ {
		if args[i] == getFlag {
			if i+1 >= len(args) {
				usage()
			}
			goGetDir = args[i+1]
			checkGopath(goGetDir)
			fmt.Fprintln(fout, goGetDir)
			i++
			continue
		}

		if strings.HasPrefix(args[i], getFlag+"=") {
			goGetDir = args[i][len(getFlag+"="):]

			checkGopath(goGetDir)
			fmt.Fprintln(fout, goGetDir)
			continue
		}

		checkGopath(args[i])
		gopathArgs = append(gopathArgs, args[i])
	}

	alreadyListed := map[string]bool{
		goGetDir: true,
	}

	for _, gopath := range w.gopaths {
		if _, ok := alreadyListed[gopath]; ok {
			continue
		}
		alreadyListed[gopath] = true
		fmt.Fprintln(fout, gopath)
	}
	for _, gopath := range gopathArgs {
		if _, ok := alreadyListed[gopath]; ok {
			continue
		}
		alreadyListed[gopath] = true
		fmt.Fprintln(fout, gopath)
	}

	return nil
}

func clone(args []string) error {
	return nil
}
