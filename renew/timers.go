// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew // import "go.pennock.tech/ocsprenewer/renew"

import (
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"golang.org/x/crypto/ocsp"
)

// Retry times always have jitter adjustments to avoid phase lock
// synchronization of requests.
const (
	// If we get told TryLater by an OCSP server, how long that is
	RetryOnTryLater = 30 * time.Minute

	// If an OCSP staple is missing any timers, how often we'll retry instead
	RetryMissingTimers = 24 * time.Hour

	// If a newly issued staple appear to have already expired, how long until
	// we try again
	RetryOnAlreadyExpired = 15 * time.Minute

	// If we're after T1 timer, how often to retry
	RetryAfterT1 = time.Hour

	// How long we wait between renew checks if we somehow failed to find any timers
	SweepIntervalTimerless = 24 * time.Hour
)

func (cr *CertRenewal) timerMatch() bool {
	raw, err := ioutil.ReadFile(cr.staplePath)
	if err != nil {
		if os.IsNotExist(err) {
			cr.CertLogf("no staple found reduces timer-match to 'yes'")
			// have no staple, therefore should fetch staple
			return true
		}
		// This is a dangerous scenario, as a system misconfiguration can lead
		// to failure to renew.  For now, logging with caps ERROR to stand out.
		// I can see an argument that this should result in `true`, but also
		// would then want a way to change permission restoration so that we're
		// not _always_ hitting this?  Although after first fetch, within
		// process lifetime, should be running off data received so fine.
		cr.CertLogf("ERROR reading staple to determine timer: %s", err)
		return false
	}

	// We want to ignore any issuer stuff here, we're just after the timers in
	// the current staple, whether valid or not.
	resp, err := ocsp.ParseResponseForCert(raw, cr.cert, nil)
	if err != nil {
		cr.CertLogf("error parsing existing staple, decreeing timer-match=yes: %s", err)
		return true
	}

	// thisUpdate: latest time known to have been good
	// producedAt: when response generated
	// Let's say we want ((nextUpdate - producedAt) * TimerT1) + producedAt as
	// the time to start retrying then.  It might be that we want thisUpdate
	// instead ... experience will tell.
	//
	// If changing this, also check setRetryTimersFromStaple()
	base := resp.ProducedAt
	expire := resp.NextUpdate
	now := time.Now()
	t1ratio := cr.Renewer.config.TimerT1
	if expire.IsZero() {
		cr.CertLogf("OCSP staple missing expiry time, need update")
		return true
	}
	if now.After(expire) {
		cr.CertLogf("OCSP staple already expired [%s], need update", expire)
		return true
	}
	if base.IsZero() {
		cr.CertLogf(
			"OCSP staple missing initial validity time; assuming need update [producedAt %s] [thisUpdate %s] [nextUpdate %s]",
			resp.ProducedAt, resp.ThisUpdate, resp.NextUpdate)
		return true
	}

	// NB: retryJitter can have us trying before T1, so this relies upon us not
	// doing this timer check when doing a "check because told to check at this
	// time".  We can reconsider and perhaps use "T+n" instead of "T±n" for
	// this jitter.
	retryAfter := base.Add(retryJitter(time.Duration(float64(expire.Sub(base)) * t1ratio)))

	if now.After(retryAfter) {
		cr.CertLogf("timer T1 expired at %s, triggering retry %vx[%s, %s]", retryAfter, t1ratio, base, expire)
		return true
	}

	cr.CertLogf("timer T1 in future, not triggering retry yet (%s from %vx[%s, %s])", retryAfter, t1ratio, base, expire)

	cr.RegisterFutureCheck(cr.certPath, retryAfter)
	return false
}

// not happy at the duplication of knowledge from timerMatch() but I value the
// precise log-messages therein, so am accepting it.
func (cr *CertRenewal) setRetryTimersFromStaple(staple *ocsp.Response) {
	if cr == nil {
		panic("nil *CertRenewal")
	}
	if !cr.NeedTimers() {
		return
	}

	atOffset := func(offset time.Duration) {
		cr.RegisterFutureCheck(cr.certPath, time.Now().Add(retryJitter(offset)))
	}

	if staple == nil {
		atOffset(RetryOnTryLater)
		return
	}

	// See equivalent roughly matching logic in timerMatch
	base := staple.ProducedAt
	expire := staple.NextUpdate
	now := time.Now()
	t1ratio := cr.Renewer.config.TimerT1
	if expire.IsZero() {
		atOffset(RetryMissingTimers)
		return
	}
	if now.After(expire) {
		atOffset(RetryOnAlreadyExpired)
		return
	}
	if base.IsZero() {
		atOffset(RetryMissingTimers)
		return
	}

	retryAfter := base.Add(retryJitter(time.Duration(float64(expire.Sub(base)) * t1ratio)))
	if now.After(retryAfter) {
		atOffset(RetryAfterT1)
		return
	}

	cr.RegisterFutureCheck(cr.certPath, retryAfter)
}

func (r *Renewer) RegisterFutureCheck(path string, checkTime time.Time) {
	if !r.NeedTimers() {
		return
	}
	r.renewMutex.Lock()
	defer r.renewMutex.Unlock()
	existing, ok := r.nextRenew[path]
	now := time.Now()
	if ok && checkTime.After(existing) && now.Before(existing) {
		r.Logf("RegisterFutureCheck(%q): ignoring time extension to %s, keeping %s", path, checkTime, existing)
		return
	}

	r.nextRenew[path] = checkTime
	r.Logf("RegisterFutureCheck(%q): at %s", path, checkTime)
	if r.earliestNextRenew.IsZero() || r.earliestNextRenew.After(checkTime) {
		r.earliestNextRenew = checkTime
	}
}

func retryJitter(base time.Duration) time.Duration {
	b := float64(base)
	// 10% +/-
	offsetFactor := rand.Float64()*0.2 - 0.1
	return time.Duration(b + offsetFactor*b)
}
