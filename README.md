oscprenewer
===========

*WARNING: this is an early-alpha project with a paucity of tests*

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

```sh
go get -v go.pennock.tech/ocsprenewer/cmd/ocsprenewer
```

### Issues

Probably lots.

Missing almost all tests.  I got something working, needs some TLC.

### Unimplemented

Would be good to have a notify-watch on a directory to automatically pick up
new certs to watch over.

Need to select an appropriate issuer certificate when it's not bundled in the
same file as the end-entity certificate.  My primary use-case is Let's Encrypt
(with certs issued through <https://github.com/xenolf/lego>) so bundling is
the norm for me.  If you need this issuer found and get the `UNIMPLEMENTED`
complaint, file an issue with details.

Should have a periodic sweep of all files, to catch unexpected or
dropped-by-bug things.
