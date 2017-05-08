// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
// "crypto/x509"
)

func (cr *CertRenewal) timerMatch() bool {
	cr.Logf("UNIMPLEMENTED timer match for %q", cr.certLabel())
	// find the staple, populate path if needed, extract times from it, etc
	return false
}
