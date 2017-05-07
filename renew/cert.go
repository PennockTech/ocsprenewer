// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	"crypto/x509"
	"fmt"
)

func certLabel(cert *x509.Certificate) string {
	if len(cert.Subject.CommonName) > 0 {
		return cert.Subject.CommonName
	}
	if len(cert.DNSNames) > 0 {
		return cert.DNSNames[0]
	}
	if len(cert.Subject.Country) > 0 && len(cert.Subject.Organization) > 0 {
		label := cert.Subject.Country[0] + " " + cert.Subject.Organization[0]
		if len(cert.Subject.OrganizationalUnit) > 0 {
			label += " " + cert.Subject.OrganizationalUnit[0]
		}
		return label
	}
	return fmt.Sprintf("%v", cert.Subject)
}
