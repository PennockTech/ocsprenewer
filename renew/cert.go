// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ocsp"
)

var (
	ErrEmptyFilename = errors.New("derived an empty filename")
)

type CertRenewal struct {
	*Renewer

	certPath   string
	staplePath string

	cert, issuer *x509.Certificate

	oldStapleRaw []byte
	oldStaple    *ocsp.Response
}

func certLabel(cert *x509.Certificate) string {
	if cert == nil {
		return "<BUG!nil>"
	}
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

func (cr *CertRenewal) certLabel() string {
	return certLabel(cr.cert)
}

func (cr *CertRenewal) findStaple() error {
	fn := filepath.Base(cr.certPath)
	for _, e := range strings.Fields(cr.Renewer.config.CertExtensions) {
		if strings.HasSuffix(fn, e) {
			fn = fn[:len(fn)-len(e)]
		}
	}
	if len(fn) == 0 {
		return ErrEmptyFilename
	}

	cr.staplePath = filepath.Join(cr.Renewer.config.OutputDir, fn+cr.Renewer.config.Extension)

	// All my shell-based tooling stores in DER format, and some quick searches
	// aren't showing anyone using PEM.  This could be a search deficiency.
	// If you need proofs stored in PEM, submit a Pull Request (or open an Issue).

	var err error
	cr.oldStapleRaw, err = ioutil.ReadFile(cr.staplePath)
	if err != nil {
		if os.IsNotExist(err) {
			cr.LogAtf(1, "%q: no existing staple at %q", cr.certLabel(), cr.staplePath)
			return nil
		}
		return err
	}

	cr.LogAtf(1, "%q: found existing staple at %q", cr.certLabel(), cr.staplePath)

	return cr.parseExistingStaple()
}

// we split this out from findStaple because we might grab the issuer later and
// set it in the *CertRenewal, in which case a validation failure becomes
// interesting.
func (cr *CertRenewal) parseExistingStaple() error {
	var err error
	cr.oldStaple, err = ocsp.ParseResponse(cr.oldStapleRaw, cr.issuer)
	return err
}
