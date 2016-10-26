// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/neubot/baudelaire/common"
	"github.com/neubot/baudelaire/neubot/rendezvous"
	"github.com/neubot/baudelaire/ooni/collector"
	"log"
	"log/syslog"
	"net/http"
	"os"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) == 2 && os.Args[1] == "--version" {
		log.Printf("%s", common.Version)
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
	log.Printf("baudelaire neubot master-server %s starting up", common.Version)

	router := httprouter.New()

	router.POST("/rendezvous", rendezvous.Handle)
	router.GET("/rendezvous",
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			log.Printf("received GET request for /rendezvous")
			rendezvous.Handle(w, r, ps)
		})

	router.POST("/report/:id", collector.Update)
	router.POST("/report/:id/close", collector.Close)
	router.POST("/report", collector.Create)

	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("cannot listen")
	}
}
