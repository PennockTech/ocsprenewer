// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	_ struct{} // we reserve right to re-order, etc, the fields here

	HTTPStatus        string  // host:port listen spec
	Directories       bool    // whether InputPaths denotes directories or not
	OutputDir         string  // where to place generated OCSP staples
	Extension         string  // filename extension to put on staples
	TimerT1           float64 // how far through staple validity period to start trying to renew
	Immediate         bool    // renew on start-up, independent of timers
	AllowNonOCSPInDir bool    // just skip any certs which lack OCSP information
	CertExtensions    string  // when scanning dirs, files with one of these extensions is assumed to be a cert
	HTTPUserAgent     string  // HTTP User-Agent to send
	InputPaths        []string
}

type Renewer struct {
	_ struct{}

	// Modify HTTPClient if your application requires that; it defaults to http.DefaultClient
	HTTPClient *http.Client

	config    Config
	certGlobs []string
	logLevel  uint

	// these are currently controlled via the -not-really flag but could be
	// more fine-grained, thus the split.  Probably makes sense to block file
	// updates in most tests.
	permitRemoteComms bool
	permitFileUpdate  bool

	// used in persist/etc modes, to indicate timers are wanted
	needTimers bool

	// Could probably do with a more efficient and scalable data structure if
	// there are more than a dozen certs to be renewed, but a map which is
	// walked to find appropriate times is acceptable for the currently
	// envisioned scale.
	//
	// If someone wants to renew thousands of certs with this tool, we can
	// revisit this at that time.  We'd also need to renew concurrently, with
	// concurrency limits, instead of one-at-a-time as we are currently.
	nextRenew map[string]time.Time
}

func New(c Config) (*Renewer, error) {
	r := Renewer{
		config:            c,
		nextRenew:         make(map[string]time.Time),
		permitRemoteComms: true,
		permitFileUpdate:  true,
		HTTPClient:        http.DefaultClient,
	}

	if r.config.HTTPUserAgent == "" {
		return nil, errors.New("you must take accountability with an HTTP User-Agent")
	}

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

func (r *Renewer) SetLogLevel(lvl uint) {
	r.logLevel = lvl
}

func (r *Renewer) SetImmediate(i bool) error {
	r.config.Immediate = i
	return nil
}

func (r *Renewer) SetNotReally(nr bool) {
	r.permitRemoteComms = !nr
	r.permitFileUpdate = !nr
}

func (r *Renewer) NeedTimers() bool {
	return r.needTimers
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

func (r *Renewer) httpDo(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", r.config.HTTPUserAgent)
	return r.HTTPClient.Do(req)
}
