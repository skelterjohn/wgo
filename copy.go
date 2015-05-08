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
	//	"fmt"
	"io"
	"os"
	"path/filepath"
)

func copyDir(src, dst string) {
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		finfo, err := os.Stat(path)
		if err != nil {
			//fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			return nil
		}
		if finfo.IsDir() {
			return os.MkdirAll(dstPath, finfo.Mode())
		}
		return copyFile(finfo, path, dstPath)
	}
	filepath.Walk(src, walk)
}

func copyFile(finfo os.FileInfo, src, dst string) error {
	fin, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fin.Close()

	fout, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, finfo.Mode())
	if err != nil {
		return err
	}
	defer fout.Close()

	if _, err = io.Copy(fout, fin); err != nil {
		return err
	}
	if err = fout.Close(); err != nil {
		return err
	}
	return fin.Close()
}
