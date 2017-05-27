// Copyright © 2017 Pennock Tech, LLC.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package renew // import "go.pennock.tech/ocsprenewer/renew"

import (
	"crypto/rand"
	"math/big"
	"sync/atomic"
)

func (r *Renewer) nextActionID() uint32 {
	return atomic.AddUint32(&r.seqActionID, 1)
}

// We just need something; could return 0, could use math/rand, but for the
// sake of avoiding arguments let's just use something more random.  We're
// going to do enough crypto stuff later that we might as well trigger any
// state initialization.
func seedActionID() uint32 {
	n, err := rand.Int(rand.Reader, big.NewInt(2<<20))
	if err != nil {
		panic(err)
	}
	return uint32(n.Uint64())
}
