.PHONY: install

prefix = /usr/local
sysconfdir = /etc
bindir = $(prefix)/bin
libdir = $(prefix)/lib

baudelaire: main.go
	go build

install: baudelaire baudelaire.service rc.local
	install -d $(bindir)
	install -v baudelaire $(bindir)
	install -v -d $(libdir)/systemd/system
	install -v -m644 baudelaire.service $(libdir)/systemd/system
	install -d $(sysconfdir)
	install -v rc.local $(sysconfdir)
