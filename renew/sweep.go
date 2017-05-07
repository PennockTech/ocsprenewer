// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

import (
	"errors"
)

// OneShot does a sweep of all candidates and renews if appropriate.
// Appropriateness is a combination of "immediate" and timers.
func (r *Renewer) OneShot() error {
	return errors.New("unimplemented")
}
