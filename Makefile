.PHONY: install

prefix = /usr/local
sysconfdir = /etc
bindir = $(prefix)/bin
libdir = $(prefix)/lib

install: baudelaire baudelaire.service rc.local
	install -d $(bindir)
	install -v baudelaire $(bindir)
	install -v -d $(libdir)/systemd/system
	install -v -m644 baudelaire.service $(libdir)/systemd/system
	install -d $(sysconfdir)
	install -v rc.local $(sysconfdir)

baudelaire: main.go
	GOOS=linux GOARCH=amd64 go build

deploy: baudelaire
	scp Makefile baudelaire baudelaire.service rc.local master.neubot.org:
