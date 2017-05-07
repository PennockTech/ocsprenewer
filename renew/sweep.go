// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const NoOCSPExtension = ".noocsp"

var ErrNoOCSPFlagfile = errors.New("a .noocsp flag-file prevented action")
var ErrNoCertsFound = errors.New("no certificate files found in a directory")

// OneShot does a sweep of all candidates and renews if appropriate.
// Appropriateness is a combination of "immediate" and timers.
func (r *Renewer) OneShot() error {
	failed := 0
	for i := range r.config.InputPaths {
		err := r.oneInputPath(r.config.InputPaths[i])
		if err != nil {
			log.Printf("failure: %s", err)
			failed += 1
		}
	}
	if failed > 0 {
		return fmt.Errorf("encountered %d failures", failed)
	}
	return nil
}

func (r *Renewer) oneInputPath(p string) error {
	fi, err := os.Stat(p)
	if err != nil {
		return err
	}
	if r.config.Directories {
		if fi.IsDir() {
			return r.oneInputDirectory(p)
		}
		return fmt.Errorf("not a directory: %q", p)
	}
	if fi.Mode().IsRegular() {
		return r.oneFilename(p)
	}
	return fmt.Errorf("not a regular file: %q", p)
}

func (r *Renewer) oneFilename(p string) error {
	_, err := os.Stat(p + NoOCSPExtension)
	if err == nil {
		return ErrNoOCSPFlagfile
	}

	return errors.New("UNIMPLEMENTED")
}

func (r *Renewer) oneInputDirectory(dirname string) error {
	var candidates []string
	var errList []error

	for _, g := range []string{"*.crt", "*.pem", "*.cert"} {
		m, err := filepath.Glob(filepath.Join(dirname, g))
		if err != nil {
			return err
		}
		if m != nil {
			candidates = append(candidates, m...)
		}
	}
	if candidates == nil {
		return ErrNoCertsFound
	}

	tried := 0
	for _, c := range candidates {
		_, err := os.Stat(c + NoOCSPExtension)
		if err == nil {
			continue
		}
		err = r.oneFilename(c)
		if err != nil {
			errList = append(errList, err)
		}
	}

	if errList != nil {
		return fmt.Errorf("saw %d errors in dir %q: %v", len(errList), dirname, errList)
	}
	if tried == 0 {
		return ErrNoCertsFound // all excluded by NoOCSPExtension is an error for us, I think
	}
	return nil
}
