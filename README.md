oscprenewer
===========

*WARNING: this is an beta project with a paucity of tests*

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

The script `.compile` will embed this repository's git version information
into the binary, and can be invoked as `./.compile static` to get a static
binary.


### Issues

Probably lots.

Missing almost all tests.  I got something working, needs some TLC.

### Unimplemented

Would be good to have a notify-watch on a directory to automatically pick up
new certs to watch over.  Also replaced (renewed) certs.

Need to select an appropriate issuer certificate when it's not bundled in the
same file as the end-entity certificate.  My primary use-case is Let's Encrypt
(with certs issued through <https://github.com/xenolf/lego>) so bundling is
the norm for me.  If you need this issuer found and get the `UNIMPLEMENTED`
complaint, file an issue with details.

Should have a periodic sweep of all files, to catch unexpected or
dropped-by-bug things.  (Can be done with SIGUSR2 now).

### Invocation

Invoke with `-help` to see flags.

Use `-persist` to set up timers and retry as and when appropriate.  Signal
handlers will also be setup, so that `SIGUSR1` will trigger an immediate
check, per timers, and `SIGUSR2` will trigger a full check, ignoring timers,
forcibly getting new staples.
(`SIGHUP` is also accepted, as per SIGUSR1, but we reserve the right to make
that do Other Things in the future, including full checks and anything else
appropriate).

There's no self-daemon mode.  Instead, run it in the "foreground" under a
keep-alive system, such as `supervise`, or a "modern" init system, or
whatever.

In the `contrib/` sub-directory, there's an `rc.d` script for FreeBSD which
will need adjustment for your installation.  The path to the command, the
place you choose to keep OCSP staples, and where the TLS certificates are
stored are likely to need adjusting.  The script relies upon `chpst` from the
`runit` package.

The core lines are:
```
: ${ocsprenewer_flags="-out-dir /var/cache/exim -cert-extensions .crt -extension .ocsp.der -now -allow-nonocsp-in-dir -dirs -persist /etc/x509/services/exim"}

/usr/sbin/daemon >> "$ocsprenewer_logfile" 2>&1 \
  -c -P "$pidfile" -r \
  /usr/local/sbin/chpst -u "$ocsprenewer_daemon_user" \
  $command ${ocsprenewer_flags}`
```

The `-now` forces a scan on startup, despite `-persist`; we're working on
"every cert in a directory" but accept that some certs are issued without OCSP
information (a private CA) so skip those without erroring.

We tell daemon to switch to the root directory (`-c`), to supervise and
restart if needed (`-r`) and to write its _own_ pid to a pidfile (`-P` instead
of `-p`) so that rc-system signals are sent to daemon, not ocsprenewer,
allowing the service to be shut down.
