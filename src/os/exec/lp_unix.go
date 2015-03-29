// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package exec

import (
	"errors"
	"os"
	"strings"
)

// ErrNotFound is the error resulting if a path search failed to find an executable file.
var ErrNotFound = errors.New("executable file not found in $PATH")

func findExecutable(file string) error {
	d, err := os.Stat(file)
	if err != nil {
		return err
	}
	if m := d.Mode(); !m.IsDir() && m&0111 != 0 {
		return nil
	}
	return os.ErrPermission
}

// LookPath searches for an executable binary named file
// in the directories named by the PATH environment variable.
// If file contains a slash, it is tried directly and the PATH is not consulted.
// The result may be an absolute path or a path relative to the current directory.
func LookPath(file string) (string, error) {
	bin_sl, err := lookPaths(file, true)
	if 0 == len(bin_sl) {
		return "", err
	}
	return bin_sl[0], nil
}

// LookPath searches for all executables binary named file
// which can be found in the directories named by the PATH environment variable.
// If file contains a slash, it is tried directly and the PATH is not consulted.
// The result may be a slice of an absolute path or a path relative to the current directory.
func LookPaths(file string) ([]string, error) {
	return lookPaths(file, false)
}

func lookPaths(file string, stop_at_first bool) ([]string, error) {
	// NOTE(rsc): I wish we could use the Plan 9 behavior here
	// (only bypass the path if file begins with / or ./ or ../)
	// but that would not match all the Unix shells.

	if strings.Contains(file, "/") {
		err := findExecutable(file)
		if err == nil {
			return []string{file}, nil
		}
		return []string{}, &Error{file, err}
	}
	pathenv := os.Getenv("PATH")
	if pathenv == "" {
		return []string{}, &Error{file, ErrNotFound}
	}

	dir_slice := strings.Split(pathenv, ":")
	var ret []string
	if !stop_at_first {
		ret = make([]string, 0, len(dir_slice))
	} else {
		ret = make([]string, 0, 1)
	}

	for _, dir := range dir_slice {
		if dir == "" {
			// Unix shell semantics: path element "" means "."
			dir = "."
		}
		path := dir + "/" + file
		if err := findExecutable(path); err == nil {
			ret = append(ret, path)
			if stop_at_first {
				break
			}
		}
	}

	var err error
	if 0 == len(ret) {
		err = &Error{file, ErrNotFound}
	} else {
		err = nil
	}
	return ret, err
}
