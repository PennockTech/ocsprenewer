// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	"crypto/x509"
)

func (r *Renewer) timerMatch(cert *x509.Certificate) bool {
	r.Logf("UNIMPLEMENTED timer match for %q", certLabel(cert))
	return false
}
