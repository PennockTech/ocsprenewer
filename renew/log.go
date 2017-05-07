// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	"log"
)

func (r *Renewer) Logf(spec string, args ...interface{}) {
	log.Printf(spec, args...)
}

func (r *Renewer) LogAtf(level uint, spec string, args ...interface{}) {
	if r.logLevel >= level {
		log.Printf(spec, args...)
	}
}
