// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// This file is split out to keep `os` and such like away from main.go and
// make it harder to do things like "exit without safety sleep".

package main

import (
	"fmt"
	"os"
	"time"
)

func exit(val int) {
	if pflags.Persist {
		// avoid being restarted in a tight loop
		time.Sleep(time.Second)
	}
	os.Exit(val)
}

func stderr(spec string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, spec, args...)
}

func stdout(spec string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, spec, args...)
}
