// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

// Start creates a persisting process which keeps renewing all OCSP staples forever.
// It exits with a bool which indicates whether exit was expected or not.
// The HTTP interface might in future provide a means to request a clean expected exit.
func (r *Renewer) Start() bool {
	return false
	// Use OneShot, then set `r.config.Immediate = false`, then use OneShot based upon calculated timers
	// (which are stored in renewer)
}
