// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew // import "go.pennock.tech/ocsprenewer/renew"

import (
	"log"
	"strconv"
)

func (r *Renewer) Logf(spec string, args ...interface{}) {
	log.Printf(spec, args...)
}

func (r *Renewer) LogAtf(level uint, spec string, args ...interface{}) {
	if r.logLevel >= level {
		r.Logf(spec, args...)
	}
}

func (cr *CertRenewal) CertLogf(spec string, args ...interface{}) {
	t := make([]interface{}, 2, len(args)+2)
	if cr.actionIDStr == "" {
		cr.actionIDStr = strconv.FormatUint(cr.ActionID, 10)
	}
	t[0] = cr.actionIDStr
	t[1] = cr.certLabel()
	t = append(t, args...)
	cr.Logf("[%s] %q: "+spec, t...)
}

func (cr *CertRenewal) CertLogAtf(level uint, spec string, args ...interface{}) {
	if cr.Renewer.logLevel >= level {
		cr.CertLogf(spec, args...)
	}
}
