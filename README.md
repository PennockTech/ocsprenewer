oscprenewer
===========

The <abbr title='Online Certificate Status Protocol'>OCSP</abbr> Renewer
renews OCSP Staples on local disk.

An X.509 PKIX setup can use checks that a certificate is still valid, in the
form of OCSP proofs.  These can be requested of the CA's servers by the
clients, but this is Bad in many ways.  The proofs can also be requested by
the server using the certificate which needs to be proved still valid, and
delivered over TLS to the client.  In this form, they're called "staples".

Each staple has a short lifetime, typically on the order of days, and must be
renewed to be valid.  Ideally, something on the server side starts trying to
renew before the end of the current validity period and keeps trying until
successful.

For simple standalone servers this functionality is best if just embedded in
the server.  For some setups, it's best to have one system with the
responsibility of managing this renewal and distributing that out to the
components which need it.

The OCSP Renewer tooling provides a library suitable for use by Golang
programs to manage this flow, and a command-line tool suitable for use in
fairly arbitrary setups.  The command-line tool can run as a daemon, or a
one-shot "try".

This author uses the Exim MTA which is able to serve staples as long as
they're provided to it on local filesystem storage; Exim does nothing to try
to renew staples, but will just use what it's given.  `ocsprenewer` was
written with that as a key usage model.


### Installation

TBD, probably `go get go.pennock.tech/ocsprenewer/...`


### Issues

Probably lots.

