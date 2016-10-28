// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/neubot/baudelaire/common"
	"github.com/neubot/baudelaire/neubot/rendezvous"
	"github.com/neubot/baudelaire/ooni/collector"
	"github.com/neubot/bernini"
	"github.com/pborman/getopt"
	"log"
	"net/http"
	"os"
)

const usage = `usage: baudelaire [-d directory] [-p port] [[-p port] ...]
       baudelaire [--version]
       baudelaire [--help]`

func main() {
	bernini.InitRng()
	bernini.InitLogger()

	bernini.GetoptVersionAndHelp(common.Version, usage)

	ports := getopt.List('p', "Set ports where to list")
	run_path := getopt.String('d', "", "Set runtime directory")

	if err := getopt.Getopt(nil); err != nil {
		fmt.Printf("%s\n", usage)
		os.Exit(1)
	}
	optarg := getopt.Args()
	if len(optarg) != 0 {
		fmt.Printf("%s\n", usage)
		os.Exit(1)
	}

	if *run_path != "" {
		err := os.Chdir(*run_path)
		if err != nil {
			log.Fatal("cannot change directory")
		}
	}

	bernini.UseSyslogOrDie("baudelaire")
	log.Printf("baudelaire server %s starting up", common.Version)

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

	if len(*ports) <= 0 {
		*ports = []string{"8080"}
	}
	channel := make(chan error)
	for i := 0; i < len(*ports); i += 1 {
		port := (*ports)[i]
		go func() {
			err := http.ListenAndServe(":" + port, router)
			if err != nil {
				channel <- err
				return
			}
			log.Printf("listening at :%s", port)
			// Note: do not emit anything here such that the main loop
			// is now blocked on the channel and we loop forever
		}()
	}
	for err := range channel {
		if err != nil {
			log.Printf("error: %s", err)
			os.Exit(1)
		}
	}
}
