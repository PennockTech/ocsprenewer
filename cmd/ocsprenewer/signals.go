// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package main // import "go.pennock.tech/ocsprenewer/cmd/ocsprenewer"

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.pennock.tech/ocsprenewer/renew"
)

func setupSignals(r *renew.Renewer) {
	// NB: package Signal DOES NOT BLOCK writing to the channel, so it MUST be buffered.
	// [caught by staticcheck]
	chNormal := make(chan os.Signal, 1)
	chFull := make(chan os.Signal, 1)

	handlerFunc := func(c <-chan os.Signal, full bool) {
		for {
			s := <-c
			r.Logf("received signal %v, triggering forced renew (full=%v)", s, full)
			r.ForceCheckSoon(full)
			time.Sleep(100 * time.Millisecond)
		}
	}
	go handlerFunc(chNormal, false)
	go handlerFunc(chFull, true)

	signal.Notify(chNormal, syscall.SIGHUP, syscall.SIGUSR1)
	signal.Notify(chFull, syscall.SIGUSR2)
}
