// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew // import "go.pennock.tech/ocsprenewer/renew"

import (
	"sort"
	"time"
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

	if len(r.nextRenew) < 1 {
	ContingencyPlanB:
		for {
			d := retryJitter(SweepIntervalTimerless)
			r.Logf("BAD: no scheduled renew checks found; sleeping %v", d)
			time.Sleep(d)

			err := r.OneShot()
			if err != nil {
				r.Logf("partial sweep errored: %s", err)
			}
			if len(r.nextRenew) >= 1 {
				break ContingencyPlanB
			}
		}
	}

	previousLoopStartTime := time.Now()
	for {
		r.renewMutex.Lock()
		firstRenewal := r.earliestNextRenew
		r.renewMutex.Unlock()
		now := time.Now()

		if now.Sub(previousLoopStartTime) < time.Second {
			// An extra sleep on the first pass is acceptable.
			// Mostly avoiding busy loops in the event something goes badly wrong.
			time.Sleep(time.Second)
			now = time.Now()
		}
		previousLoopStartTime = now

		if firstRenewal.After(now) {
			// FUTURE: with a service HTTP port and requested exit, or better
			// signal handling, or dynamic pickup via notify watch, we'd have a
			// select{} block here, but for now keep it simple.
			d := firstRenewal.Sub(now)
			r.Logf("persist-sleep: next renewal at %s, sleeping %s", firstRenewal, d)
			time.Sleep(d)
		}

		err := r.runTimerBasedChecks()
		// Use a sentinel error to request exit?
		if err != nil {
			r.Logf("timer-based sweep errored: %s", err)
		}
		continue
	}

	return false
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
	foundEnd := false
	for i := range timePaths {
		if timePaths[i].T.After(now) {
			foundEnd = true
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

	err := r.oneShotOverPaths(paths)

	if !foundEnd {
		timePaths := r.getTimePaths()
		sort.Slice(timePaths, func(i, j int) bool { return timePaths[i].T.Before(timePaths[j].T) })
		r.renewMutex.Lock()
		if timePaths[0].T.Before(r.earliestNextRenew) {
			r.earliestNextRenew = timePaths[0].T
		}
		r.renewMutex.Unlock()
	}

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
