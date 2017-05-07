// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const NoOCSPExtension = ".noocsp"
const MaxCertFileSize = 1024 * 1024 // not processing a cert file larger than 1MB

var (
	ErrNoOCSPFlagfile   = errors.New("a .noocsp flag-file prevented action")
	ErrNoCertsFound     = errors.New("no certificate files found in a directory")
	ErrCertFileTooLarge = errors.New("certificate file too large")
	ErrNotCertificate   = errors.New("no certificate found in file")
	ErrNoOCSPInCert     = errors.New("certificate lacks OCSP information")
)

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

func (r *Renewer) oneInputDirectory(dirname string) error {
	var candidates []string
	var errCount int

	for _, g := range r.certGlobs {
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
		if !r.oneFilenameSuccess(c) {
			errCount += 1
		}
	}

	if errCount > 0 {
		return fmt.Errorf("saw %d errors in dir %q", errCount, dirname)
	}
	if tried == 0 {
		return ErrNoCertsFound // all excluded by NoOCSPExtension is an error for us, I think
	}
	return nil
}

// oneFilenameSuccess should only be used when scanning directories and is
// allowed to suppress errors on that basis
func (r *Renewer) oneFilenameSuccess(p string) bool {
	err := r.oneFilename(p)
	if err == nil {
		return true
	}
	if r.config.AllowNonOCSPInDir && err == ErrNoOCSPInCert {
		// TODO: if gain verboseness, log something here
		return true
	}
	log.Printf("failed on %q: %s", p, err)
	return false
}

func (r *Renewer) oneFilename(p string) error {
	fi, err := os.Stat(p + NoOCSPExtension)
	if err == nil {
		return ErrNoOCSPFlagfile
	}

	fi, err = os.Stat(p)
	if err != nil {
		return err
	}
	if fi.Size() > MaxCertFileSize {
		return ErrCertFileTooLarge
	}

	data, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}

	// We currently _only_ handle PEM input, and we only look at the first cert
	// in a file, ignoring any chain.  We ignore any PEM headers.

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return ErrNotCertificate
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}
	if len(cert.OCSPServer) < 1 {
		return ErrNoOCSPInCert
	}

	for i := range cert.OCSPServer {
		log.Printf("cert %q OCSP server %q", p, cert.OCSPServer[i])
	}

	if r.config.Immediate {
		return r.renewOneCert(cert, p)
	}
	if r.timerMatch(cert) {
		return r.renewOneCert(cert, p)
	}

	// TODO: this should be verboseness-constrained
	log.Printf("cert %q (%s) skipping for not within timer", p, certLabel(cert))
	return nil
}
