// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/ocsp"
)

// https://gist.github.com/sleevi/5efe9ef98961ecfb4da8
// Good notes.  My use of the term "TimerT1" is because I'd already decided
// that a sane approach is based on some kind of renewal model, such as DHCP's
// T1/T2 system.

var (
	ErrCertAlreadyExpired = errors.New("refuse to fetch OCSP staple for expired cert")
	ErrNoIssuer           = errors.New("unable to find an issuer to validate any OCSP response")
)

// We're responsible both for the renewal over the wire and for updating any
// staple in filesystem.
func (cr *CertRenewal) renewOneCertNow(rawRestOfChain []byte) error {

	if len(cr.cert.OCSPServer) < 1 {
		return ErrNoOCSPInCert
	}

	if time.Now().After(cr.cert.NotAfter) {
		return ErrCertAlreadyExpired
	}

	if cr.issuer == nil {
		cr.issuer = cr.tryIssuerInRest(rawRestOfChain)
	}
	if cr.issuer == nil {
		cr.issuer = cr.findIssuer()
	}
	if cr.issuer == nil {
		return ErrNoIssuer
	}

	cr.LogAtf(1, "issuer for %q is %q", certLabel(cr.cert), certLabel(cr.issuer))

	// FIXME: figure out where we'd write to local disk and ensure we can,
	// _before_ we speak remotely.  Unless that's inhibited.

	if !cr.Renewer.permitRemoteComms {
		cr.Logf("remote OCSP renewal inhibited, blocking renew of %q", certLabel(cr.cert))
		return nil
	}

	req, err := ocsp.CreateRequest(cr.cert, cr.issuer, nil)
	if err != nil {
		return err
	}

	return fmt.Errorf("UNIMPLEMENTED FOR %q with REQ %v", cr.certLabel(), req)

	// FIXME: update on local disk
}

func (cr *CertRenewal) tryIssuerInRest(rest []byte) *x509.Certificate {
	if len(rest) == 0 {
		return nil
	}
	block, _ := pem.Decode(rest)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		cr.Logf("remnant data in %q not issuer chain, X509 parse failed: %s", cr.certPath, err)
		return nil
	}
	return cert
}

// move this out to something which manages system pools, any CAs specified in
// cmdline flags, etc
func (cr *CertRenewal) findIssuer() *x509.Certificate {
	cr.Logf("UNIMPLEMENTED findIssuer(%q, %q)", cr.certLabel(), cr.certPath)
	return nil
}
