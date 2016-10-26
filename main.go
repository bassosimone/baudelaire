// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

package main

import (
	"github.com/neubot/baudelaire/neubot/rendezvous"
	"log"
	"log/syslog"
	"net/http"
	"os"
)

const version = "v0.0.2-dev"

func main() {
	log.SetFlags(0)

	if len(os.Args) == 2 && os.Args[1] == "--version" {
		log.Printf("%s", version)
		os.Exit(0)
	}
	if len(os.Args) != 1 {
		log.Fatal("usage: baudelaire [--version]")
	}

	// See http://technosophos.com/2013/09/14/using-gos-built-logger-log-syslog.html
	log.Print("redirecting logs to the system logger")
	logwriter, err := syslog.New(syslog.LOG_NOTICE, "baudelaire")
	if err != nil {
		log.Fatal("cannot initialize syslog")
	}

	log.SetOutput(logwriter)
	log.Printf("baudelaire neubot master-server %s starting up", version)

	http.HandleFunc("/rendezvous", rendezvous.Handle)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("cannot listen")
	}
}
