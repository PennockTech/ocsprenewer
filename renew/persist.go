// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew // import "go.pennock.tech/ocsprenewer/renew"

import (
	"sort"
	"time"
)

const (
	minSpinningLoopBackoff      = time.Second
	maxSpinningLoopBackoff      = 5 * time.Minute
	georatioSpinningLoopBackoff = 2.0
)

// Start creates a persisting process which keeps renewing all OCSP staples forever.
// It exits with a bool which indicates whether exit was expected or not.
// The HTTP interface might in future provide a means to request a clean expected exit.
func (r *Renewer) Start() (status bool) {
	status = false
	defer func() {
		r.Logf("exiting persistent sweep, no more timer-based renews")
	}()

	r.needTimers = true

	err := r.OneShot()
	if err != nil {
		r.Logf("First sweep errored: %s", err)
	}

	r.config.Immediate = false

	previousLoopStartTime := time.Now()
	var spinningLoopBackoff time.Duration
	for {
		r.renewMutex.Lock()
		firstRenewal := r.earliestNextRenew
		r.renewMutex.Unlock()
		now := time.Now()
		emptyTimers := false
		safetyPause := false

		if firstRenewal.IsZero() {
			emptyTimers = true
			d := retryJitter(SweepIntervalTimerless)
			if r.permitRemoteComms {
				r.Logf("BAD: no scheduled renew checks found; will sleep for %v", d)
			} else {
				r.Logf("remote comms disabled, unable to get timers; assuming you're testing; will sleep for %v", d)
			}
			firstRenewal = now.Add(d)
		}

		// Left as zero for first pass to prevent sleeping and logging on startup
		if spinningLoopBackoff == 0 {
			spinningLoopBackoff = minSpinningLoopBackoff
		} else if now.Sub(previousLoopStartTime) < spinningLoopBackoff {
			r.Logf("CPU-protection: sleeping for %v", spinningLoopBackoff)
			time.Sleep(spinningLoopBackoff)
			now = time.Now()
			spinningLoopBackoff *= georatioSpinningLoopBackoff
			if spinningLoopBackoff > maxSpinningLoopBackoff {
				spinningLoopBackoff = maxSpinningLoopBackoff
			}
		} else {
			spinningLoopBackoff = minSpinningLoopBackoff
		}
		previousLoopStartTime = now

		if firstRenewal.After(now) {
			// FUTURE: with a service HTTP port and requested exit, or better
			// signal handling, or dynamic pickup via notify watch, we'd have a
			// select{} block here, but for now keep it simple.
			d := firstRenewal.Sub(now)
			r.Logf("persist-sleep: next renewal at %s, sleeping %s", firstRenewal, d)
			r.sleepUnlessInterrupted(d)
		}
	SAFETY_RESTART:
		if safetyPause {
			// explained below, just before the Evil Goto
			r.sleepUnlessInterrupted(2 * minSpinningLoopBackoff)
		}

		if t, full := r.forcedSweepCheck(); !t.IsZero() {
			if full {
				r.config.Immediate = true
			}
			err := r.OneShot()
			if err != nil {
				r.Logf("forced full sweep errored: %s", err)
			}
			r.config.Immediate = false
			r.forcedSweepResetFor(t)
		} else if emptyTimers {
			err := r.OneShot()
			if err != nil {
				r.Logf("timerless fallback full sweep errored: %s", err)
			}
		} else if safetyPause {
			safetyPause = false
		} else {
			err := r.runTimerBasedChecks()
			// Use a sentinel error to request exit?
			if err != nil {
				r.Logf("timer-based sweep errored: %s", err)
			} else {
				// We think everything is fine, after a time-based run, so we
				// have a choice about the CPU-protection sleep: let it log
				// normally, or avoid logging for a "normal" case, while still
				// backing off a little.  I think the best thing to do is to
				// sleep for twice the minimum, so that when things are
				// well-behaved, we don't sleep again, but if things had spun a
				// bit to ramp the backoff up past this, we'll still sleep
				// again and get protection.
				// None of this is ideal or exact; it's heuristic to protect
				// against bugs in our own logic but avoid noise.
				// We also want to handle signals, even in this pause, so we
				// need all of this post-sleep logic to apply again.
				r.Logf("pausing and cleaning up")
				safetyPause = true
				cleanUpState()
				goto SAFETY_RESTART
			}
		}
		continue
	}
}

type timePath struct {
	T time.Time
	P string
}

// We don't guarantee that we'll run all the checks whose start-time will have
// passed by the time we finish, but do guarantee that r.earliestNextRenew will
// be the earliest of those we _don't_ handle, so that the next sweep should
// pick those up immediately.
func (r *Renewer) runTimerBasedChecks() error {
	timePaths := r.getTimePaths()
	sort.Slice(timePaths, func(i, j int) bool { return timePaths[i].T.Before(timePaths[j].T) })
	r.LogAtf(2, "ordered check list: %#v", timePaths)

	now := time.Now()
	for i := range timePaths {
		if timePaths[i].T.After(now) {
			r.renewMutex.Lock()
			r.earliestNextRenew = timePaths[i].T
			r.renewMutex.Unlock()
			timePaths = timePaths[:i]
		}
	}

	paths := make([]string, len(timePaths))
	for i := range timePaths {
		paths[i] = timePaths[i].P
	}

	err := r.sweepOverPaths(paths, r.oneFilename)

	// We don't know if the sweep will have registered new checks before the
	// earliest of any remaining checks, since OCSP leases can be for varying
	// durations.  So we _always_ look at all now-current timepaths to
	// determine the earliest, regardless of whether or not we just looked at
	// all of them.

	timePaths = r.getTimePaths()
	sort.Slice(timePaths, func(i, j int) bool { return timePaths[i].T.Before(timePaths[j].T) })
	r.renewMutex.Lock()
	if timePaths[0].T.Before(r.earliestNextRenew) {
		r.earliestNextRenew = timePaths[0].T
	}
	r.renewMutex.Unlock()

	return err
}

// unsorted but complete list of times
func (r *Renewer) getTimePaths() []timePath {
	r.renewMutex.Lock()
	defer r.renewMutex.Unlock()
	tp := make([]timePath, len(r.nextRenew))
	i := 0
	for p, t := range r.nextRenew {
		tp[i].T = t
		tp[i].P = p
		i++
	}
	return tp
}
