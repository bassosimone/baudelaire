package main

import (
	"log"
	"net/http"
)

func rendezvous(w http.ResponseWriter, r *http.Request) {
	log.Printf(r.RemoteAddr)
    w.Write([]byte("Ok"))
}

func main() {
	http.HandleFunc("/rendezvous", rendezvous)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("cannot listen")
	}
}
