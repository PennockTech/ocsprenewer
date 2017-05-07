// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	"crypto/x509"
	"errors"
)

// We're responsible both for the renewal over the wire and for updating any
// staple in filesystem.
func (r *Renewer) renewOneCert(cert *x509.Certificate, path string) error {
	return errors.New("UNIMPLEMENTED for " + certLabel(cert))
}
