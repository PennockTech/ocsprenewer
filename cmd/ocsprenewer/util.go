// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// This file is split out to keep `os` and such like away from main.go and
// make it harder to do things like "exit without safety sleep".

package main // import "go.pennock.tech/ocsprenewer/cmd/ocsprenewer"

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
)

func exit(val int) {
	if pflags.Persist {
		// avoid being restarted in a tight loop
		time.Sleep(time.Second)
	}
	os.Exit(val)
}

func argvQuoted() string {
	b := &bytes.Buffer{}
	r := regexp.MustCompile(`^[A-Za-z0-9./_-]+$`)
	for _, arg := range os.Args {
		if r.MatchString(arg) {
			b.WriteString(arg)
		} else {
			b.WriteString(strconv.Quote(arg))
		}
		b.WriteRune(' ')
	}
	return b.String()
}

func stderr(spec string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, spec, args...)
}

func stdout(spec string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, spec, args...)
}
