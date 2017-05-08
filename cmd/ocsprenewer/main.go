// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package main

import (
	"flag"
	"fmt"
	"os"

	"go.pennock.tech/ocsprenewer/renew"
)

const (
	HTTPUserAgent = "ocsprenewer/0.1 (Pennock Tech OCSP Renewer)"
)

// We don't use "daemon" because we don't auto-fork into background, but
// instead make ourselves easy to supervise.  If there are complaints that
// daemonization is too hard, we can consider log-file redirection and
// self-daemonization as later features.
var pflags struct {
	Persist   bool
	IfNeeded  bool
	Verbose   bool
	NotReally bool
}

var renewerConfig renew.Config

func init() {
	flag.BoolVar(&pflags.Persist, "persist", false, "run in a loop, renewing as needed")
	flag.BoolVar(&pflags.Persist, "p", false, "short form of -persist")
	flag.BoolVar(&pflags.IfNeeded, "if-needed", false, "if not persisting, only renew those which need renewal per timers")
	flag.BoolVar(&pflags.Verbose, "verbose", false, "be more verbose")
	flag.BoolVar(&pflags.Verbose, "v", false, "short form of -verbose")
	flag.BoolVar(&pflags.NotReally, "not-really", false, "don't talk to remote servers, do everything else")
	flag.BoolVar(&pflags.NotReally, "n", false, "short form of -not-really")

	flag.BoolVar(&renewerConfig.Immediate, "now", false, "renew immediately in persist mode")
	flag.StringVar(&renewerConfig.HTTPStatus, "http", "", "start an HTTP status service, on given host:port spec")
	flag.BoolVar(&renewerConfig.Directories, "dirs", false, "arguments are directories containing certs")
	flag.StringVar(&renewerConfig.OutputDir, "out-dir", "./", "place files into given directory")
	flag.StringVar(&renewerConfig.Extension, "extension", ".ocsp", "create proofs in files with this extension")
	flag.Float64Var(&renewerConfig.TimerT1, "timer-t1", 0.5, "how far through staple validity period to start trying to renew")
	flag.BoolVar(&renewerConfig.AllowNonOCSPInDir, "allow-nonocsp-in-dir", false, "do not error on certs missing OCSP info")
	flag.StringVar(&renewerConfig.CertExtensions, "cert-extensions", ".crt .cert .pem", "files in dir-scan with these extensions should be certs")
}

func main() {
	flag.Parse()
	renewerConfig.InputPaths = flag.Args()
	renewerConfig.HTTPUserAgent = HTTPUserAgent

	renewer, err := renew.New(renewerConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuring OCSP renewer failed: %s", err)
		os.Exit(1)
	}
	if pflags.Verbose {
		renewer.SetLogLevel(1)
	}
	if pflags.NotReally {
		renewer.SetNotReally(true)
	}

	if pflags.Persist {
		// Should not return until exiting
		ok := renewer.Start()
		if ok {
			os.Exit(0)
		}
		os.Exit(1)
	}

	if pflags.IfNeeded {
		renewer.SetImmediate(false)
	} else {
		renewer.SetImmediate(true)
	}

	err = renewer.OneShot()
	if err != nil {
		renewer.Logf("renewing failed: %s", err)
		os.Exit(1)
	}
}
