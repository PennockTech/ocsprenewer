// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew // import "go.pennock.tech/ocsprenewer/renew"

import (
	"os"
)

// BasicChecks does whatever checks the renewer library considers worthwhile
// sanity checks to try before starting any persistent run.
func (r *Renewer) BasicChecks() error {
	fh, err := os.CreateTemp(r.config.OutputDir, "startup-check")
	if err != nil {
		return err
	}
	if err = fh.Close(); err != nil {
		return err
	}
	if err = os.Remove(fh.Name()); err != nil {
		return err
	}

	// Any other checks?

	return nil
}
