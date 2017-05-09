// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/crypto/ocsp"
)

const RetryOnTryLater = 30 * time.Minute

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

	retryAfter := base.Add(time.Duration(float64(expire.Sub(base)) * t1ratio))
	if now.After(retryAfter) {
		cr.CertLogf("timer T1 expired at %s, triggering retry %vx[%s, %s]", retryAfter, t1ratio, base, expire)
		return true
	}

	cr.CertLogf("timer T1 in future, not triggering retry yet (%s from %vx[%s, %s])", retryAfter, t1ratio, base, expire)

	cr.RegisterFutureCheck(cr.certPath, retryAfter)
	return false
}

func (cr *CertRenewal) setRetryTimersFromStaple(staple *ocsp.Response) {
	if cr == nil {
		panic("nil *CertRenewal")
	}
	if !cr.NeedTimers() {
		return
	}

	if staple == nil {
		cr.CertLogf("UNIMPLEMENTED: add %v delay for TryLater response", RetryOnTryLater)
		return
	}

	cr.CertLogf("UNIMPLEMENTED: calculate retry times")
}

func (r *Renewer) RegisterFutureCheck(path string, checkTime time.Time) {
	r.renewMutex.Lock()
	defer r.renewMutex.Unlock()
	existing, ok := r.nextRenew[path]
	if ok && checkTime.After(existing) {
		r.Logf("RegisterFutureCheck(%q): ignoring time extension to %s, keeping %s", path, checkTime, existing)
		return
	}

	r.nextRenew[path] = checkTime
	r.Logf("RegisterFutureCheck(%q): at %s", path, checkTime)
	if r.earliestNextRenew.After(checkTime) {
		r.earliestNextRenew = checkTime
	}
}
