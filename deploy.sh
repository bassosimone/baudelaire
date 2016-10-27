#!/bin/sh
set -ex

prefix=/usr/local
sysconfdir=/etc
bindir=$prefix/bin
libdir=$prefix/lib

baudelaire=baudelaire-linux-amd64

install -d $bindir
install -v $baudelaire $bindir/baudelaire
install -v -d $libdir/systemd/system
install -v -m644 baudelaire.service $libdir/systemd/system
install -d $sysconfdir
install -v rc.local $sysconfdir
/sbin/setcap 'cap_net_bind_service=+ep' $bindir/baudelaire
install -d -v /var/lib/neubot/ooni
chown nobody:nogroup /var/lib/neubot/ooni
