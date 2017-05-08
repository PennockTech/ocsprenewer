// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	// "crypto/x509"
	"time"

	"golang.org/x/crypto/ocsp"
)

const RetryOnTryLater = 30 * time.Minute

func (cr *CertRenewal) timerMatch() bool {
	cr.Logf("UNIMPLEMENTED timer match for %q", cr.certLabel())
	// find the staple, populate path if needed, extract times from it, etc
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
		cr.Logf("UNIMPLEMENTED: %q: add %v delay for TryLater response", cr.certLabel(), RetryOnTryLater)
		return
	}

	cr.Logf("UNIMPLEMENTED: %q: calculate retry times", cr.certLabel())
}
