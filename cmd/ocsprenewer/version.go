// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package main // import "go.pennock.tech/ocsprenewer/cmd/ocsprenewer"

import (
	"runtime"
	"runtime/debug"
)

var ProjectName = "OCSP Renewer"
var Version = "0.1.9"

func showVersion() {
	stdout("%s version %s\n", ProjectName, Version)
	stdout("%s: Golang: Runtime: %s\n", ProjectName, runtime.Version())

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		stdout("%s: no repo version details available\n", ProjectName)
		return
	}

	type versionLine struct {
		path, version, sum string
		replaced           bool
	}
	lines := make([]versionLine, 0, 10)
	addVersion := func(p, v, sum string, replaced bool) {
		lines = append(lines, versionLine{p, v, sum, replaced})
	}

	m := &buildInfo.Main
	topVersion := m.Version
	if Version != "" {
		topVersion = Version
	}
	addVersion(m.Path, topVersion, m.Sum, m.Replace != nil)
	for m.Replace != nil {
		m = m.Replace
		addVersion(m.Path, m.Version, m.Sum, m.Replace != nil)
	}

	for _, m := range buildInfo.Deps {
		addVersion(m.Path, m.Version, m.Sum, m.Replace != nil)
		for m.Replace != nil {
			m = m.Replace
			addVersion(m.Path, m.Version, m.Sum, m.Replace != nil)
		}
	}

	headers := []string{"Path", "Version", "Checksum", "Replaced"}
	maxP, maxV, maxS := len(headers[0]), len(headers[1]), len(headers[2])
	for _, l := range lines {
		if len(l.path) > maxP {
			maxP = len(l.path)
		}
		if len(l.version) > maxV {
			maxV = len(l.version)
		}
		if len(l.sum) > maxS {
			maxS = len(l.sum)
		}
	}
	stdout("%-*s %-*s %-*s %s\n", maxP, headers[0], maxV, headers[1], maxS, headers[2], headers[3])
	for _, l := range lines {
		stdout("%-*s %-*s %-*s %v\n", maxP, l.path, maxV, l.version, maxS, l.sum, l.replaced)
	}
}

// We expect Version to be overridable at the linker, perhaps with git
// information, so it might be more than just a tuple of digits joined with
// dots.
// In HTTP, any "token" can be used as the "product-version"
func httpVersion(ver string) string {
	for i := 0; i < len(ver); i++ {
		// see comment after func with relevant grammar
		// the RFC7230/RFC2616 approaches should be identical in result, the 2616 is simpler to code here
		if ver[i] > 126 {
			// DEL or bit 7 set, so UTF-8 sequence, non-CHAR
			return ver[:i]
		}
		switch ver[i] {
		case 0, '(', ')', '<', '>', '@', ',', ';', ':', '\\', '"', '/', '[', ']', '?', '=', '{', '}', ' ', '\t':
			return ver[:i]
		}
	}
	return ver
}

/* RFC 2616 or 7230/7231

2616:
       token          = 1*<any CHAR except CTLs or separators>
       separators     = "(" | ")" | "<" | ">" | "@"
                      | "," | ";" | ":" | "\" | <">
                      | "/" | "[" | "]" | "?" | "="
                      | "{" | "}" | SP | HT
       CTL            = <any US-ASCII control character
                        (octets 0 - 31) and DEL (127)>
       CHAR           = <any US-ASCII character (octets 0 - 127)>

7230:
   Most HTTP header field values are defined using common syntax
   components (token, quoted-string, and comment) separated by
   whitespace or specific delimiting characters.  Delimiters are chosen
   from the set of US-ASCII visual characters not allowed in a token
   (DQUOTE and "(),/:;<=>?@[\]{}").

     token          = 1*tchar

     tchar          = "!" / "#" / "$" / "%" / "&" / "'" / "*"
                    / "+" / "-" / "." / "^" / "_" / "`" / "|" / "~"
                    / DIGIT / ALPHA
                    ; any VCHAR, except delimiters
*/
