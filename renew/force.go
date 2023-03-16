// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew // import "go.pennock.tech/ocsprenewer/renew"

import (
	"time"
)

type sweepReq struct {
	T    time.Time
	Full bool
}

// We can be interrupted by a forced full sweep; if so, then we set
// forcedSweepAt to persist in the renewer object that this has happened, to
// make it easier to decide later what should happen.
func (r *Renewer) sleepUnlessInterrupted(dur time.Duration) {
	sleeper := time.NewTimer(dur)
	select {
	case <-sleeper.C:
		return
	case req := <-r.forceSweepReqs:
		r.forceAddCheck(req)
		if !sleeper.Stop() {
			<-sleeper.C
		}
		return
	}
}

// Interrupt the current sleep, force a sweep soon.
func (r *Renewer) ForceCheckSoon(full bool) {
	r.forceSweepReqs <- sweepReq{T: time.Now(), Full: full}
}

// If a sweep has been requested, return the time/identity for that.
// We could probably use a counter just as easily.  If no reset, return
// a zero time.  The boolean indicates if a full reset was requested.
func (r *Renewer) forcedSweepCheck() (time.Time, bool) {
	r.renewMutex.Lock()
	defer r.renewMutex.Unlock()
	return r.forcedSweepAt, r.forcedFull
}

// Record that we want a new check.  If multiple requests come in, we coalesce;
// the action to check notes the value it got (from forcedSweepCheck) and uses
// that to reset, so we shouldn't lose requests.
// If some requests are for Full and some aren't, then coalesce into full-needed.
// LIMITATION:
//
//	We only reset the Full flag if we successfully clear any pending reset.
//	If one or more non-full requests come in when we're in the middle of
//	dealing with a full, then we'll do full checks for the next invocation
//	too, and so on, until we clear any pending check.
//	This should be fairly innocuous and I don't think handling it is worth more
//	complexity.
func (r *Renewer) forceAddCheck(s sweepReq) {
	r.renewMutex.Lock()
	defer r.renewMutex.Unlock()
	r.forcedSweepAt = s.T
	if s.Full {
		r.forcedFull = true
	}
}

// Reset for the sequence id (time) given ... or any earlier values.
func (r *Renewer) forcedSweepResetFor(t time.Time) {
	r.renewMutex.Lock()
	defer r.renewMutex.Unlock()

	if r.forcedSweepAt.After(t) {
		return
	}
	r.forcedSweepAt = time.Time{}
	r.forcedFull = false
}
