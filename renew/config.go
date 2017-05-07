// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	_ struct{} // we reserve right to re-order, etc, the fields here

	HTTPStatus        string  // host:port listen spec
	Directories       bool    // whether InputPaths denotes directories or not
	OutputDir         string  // where to place generated OCSP staples
	Extension         string  // filename extension to put on staples
	OutPEM            bool    // use PEM, not DER, for filenames
	TimerT1           float64 // how far through staple validity period to start trying to renew
	Immediate         bool    // renew on start-up, independent of timers
	AllowNonOCSPInDir bool    // just skip any certs which lack OCSP information
	CertExtensions    string  // when scanning dirs, files with one of these extensions is assumed to be a cert
	InputPaths        []string
}

type Renewer struct {
	_ struct{}

	config    Config
	certGlobs []string
}

func New(c Config) (*Renewer, error) {
	r := Renewer{}
	r.config = c

	if len(r.config.InputPaths) == 0 {
		return nil, errors.New("no input paths to examine")
	}

	if 1 <= r.config.TimerT1 && r.config.TimerT1 <= 100 {
		// Handle percentages on cmdline, instead of ratios
		r.config.TimerT1 = r.config.TimerT1 / 100.0
	}
	if r.config.TimerT1 < 0.1 {
		return nil, errors.New("timer T1 set too small (10% minimum)")
	}
	if r.config.TimerT1 > 0.95 {
		return nil, errors.New("timer T1 set too large (95% maximum)")
	}

	if !directoryExists(r.config.OutputDir) {
		return nil, fmt.Errorf("output directory %q does not exist or is not a directory", r.config.OutputDir)
	}

	for _, e := range strings.Fields(r.config.CertExtensions) {
		r.certGlobs = append(r.certGlobs, "*"+e)
	}
	if r.certGlobs == nil {
		r.certGlobs = []string{"*.crt"}
	}

	return &r, nil
}

func (r *Renewer) SetImmediate(i bool) error {
	r.config.Immediate = i
	return nil
}

func directoryExists(d string) bool {
	fi, err := os.Stat(d)
	switch {
	case err != nil:
		return false
	case fi.IsDir():
		return true
	default:
		return false
	}
}
