// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package main // import "go.pennock.tech/ocsprenewer/cmd/ocsprenewer"

import (
	"runtime"
	"strings"
)

var ProjectName = "OCSP Renewer"
var Version = "0.1.9"
var RepoVersionString = ""

func showVersion() {
	stdout("%s version %s\n", ProjectName, Version)
	stdout("%s: Golang: Runtime: %s\n", ProjectName, runtime.Version())

	if RepoVersionString != "" {
		// Linker cmdline hack: use Unit Separator US (0x1F) within Record Separator terminated records (RS, 0x1E)
		// And whitespace replaced with Substitute (0x1A)
		for _, l := range strings.Split(RepoVersionString, "\x1E") {
			if l != "" {
				l = strings.Replace(l, "\x1A", " ", -1)
				units := strings.SplitN(l, "\x1F", 2)
				if len(units) == 1 {
					units = append(units, "<unknown>")
				}
				stdout("%s: repo %q: %q\n", ProjectName, units[0], units[1])
			}
		}
	} else {
		stdout("%s: no repo version details available\n", ProjectName)
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
