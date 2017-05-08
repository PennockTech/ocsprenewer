// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	"crypto/x509"
	"fmt"
	"time"
)

type RevokedError struct {
	Cert      *x509.Certificate
	RevokedAt time.Time
}

func (re RevokedError) Error() string {
	return fmt.Sprintf("Cert %q revoked at %v", certLabel(re.Cert), re.RevokedAt)
}

type UnknownAtCAError struct {
	Cert *x509.Certificate
	URL  string
}

func (uace UnknownAtCAError) Error() string {
	return fmt.Sprintf("Cert %q not recognized as issued by OCSP responder at %q", certLabel(uace.Cert), uace.URL)
}
