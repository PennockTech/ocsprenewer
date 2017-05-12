// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew // import "go.pennock.tech/ocsprenewer/renew"

import (
	"log"
	"os"
	"strconv"
)

func (r *Renewer) rawLogf(spec string, args ...interface{}) {
	log.Printf(spec, args...)
}

var thisPid string

func init() {
	thisPid = strconv.Itoa(os.Getpid())
}

func (r *Renewer) Logf(spec string, args ...interface{}) {
	t := make([]interface{}, 1, len(args)+1)
	t[0] = thisPid
	t = append(t, args...)
	r.rawLogf("[pid=%s] "+spec, t...)
}

func (r *Renewer) LogAtf(level uint, spec string, args ...interface{}) {
	if r.logLevel >= level {
		r.Logf(spec, args...)
	}
}

func (cr *CertRenewal) CertLogf(spec string, args ...interface{}) {
	t := make([]interface{}, 3, len(args)+3)
	if cr.actionIDStr == "" {
		cr.actionIDStr = strconv.FormatUint(cr.ActionID, 10)
	}
	t[0] = thisPid
	t[1] = cr.actionIDStr
	t[2] = cr.certLabel()
	t = append(t, args...)
	cr.rawLogf("[pid=%s] [actId=%s] [cert=%q] "+spec, t...)
}

func (cr *CertRenewal) CertLogAtf(level uint, spec string, args ...interface{}) {
	if cr.Renewer.logLevel >= level {
		cr.CertLogf(spec, args...)
	}
}
