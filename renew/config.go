// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew

type Config struct {
	_ struct{} // we reserve right to re-order, etc, the fields here

	HTTPStatus  string  // host:port listen spec
	Directories bool    // whether InputPaths denotes directories or not
	OutputDir   string  // where to place generated OCSP staples
	Extension   string  // filename extension to put on staples
	OutPEM      bool    // use PEM, not DER, for filenames
	TimerT1     float64 // how far through staple validity period to start trying to renew
	Immediate   bool    // renew on start-up, independent of timers
	InputPaths  []string
}

type Renewer struct {
	_ struct{}

	config Config
}

func New(c Config) (*Renewer, error) {
	r := Renewer{}
	r.config = c
	return &r, nil
}

func (r *Renewer) SetImmediate(i bool) error {
	r.config.Immediate = i
	return nil
}
