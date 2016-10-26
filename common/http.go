// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

package common

import (
	"log"
	"net/http"
)

var EmptyJson = []byte("{}")

func WriteResponseJson(w http.ResponseWriter, code int, body []byte) error {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json; encoding=utf-8")
	_, err := w.Write(body)
	if err != nil {
		log.Printf("cannot write http response body")
		return err
	}
	return nil
}
