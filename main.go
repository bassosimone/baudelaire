package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"strings"
)

var EmptyJson = []byte("{}")

/*-
 * Example:
 *
 * {
 *     "accept": [
 *         "speedtest",
 *         "bittorrent"
 *     ],
 *     "privacy_can_collect": 0,
 *     "privacy_can_share": 0,
 *     "privacy_informed": 0,
 *     "version": "0.4.17.0"
 * }
 *
 */
type Request struct {
	Accept            []string `json:"accept"`
	PrivacyCanCollect int      `json:"privacy_can_collect"`
	PrivacyCanShare   int      `json:"privacy_can_share"`
	PrivacyInformed   int      `json:"privacy_informed"`
	Version           string   `json:"version"`
}

/*-
 *
 * Example:
 *
 * {
 *     "available": {
 *         "bittorrent": [
 *             "http://neubot.mlab.mlab1.mil01.measurement-lab.org:8080/"
 *         ],
 *         "speedtest": [
 *             "http://neubot.mlab.mlab1.mil01.measurement-lab.org:8080/speedtest"
 *         ]
 *     },
 *     "update": {
 *         "uri": "http://neubot.org/",
 *         "version": "0.4.15.6"
 *     }
 * }
 *
 */
type Response struct {
	Update    map[string]map[string]string `json:"update"`
	Available map[string][]string          `json:"available"`
}

/*-
 * Example:
 *
 * {
 *     "city": "Turin",
 *     "url": "http://neubot.mlab.mlab1.trn01.measurement-lab.org:8080",
 *     "ip": ["194.116.85.211", "2001:7f8:23:307::211"],
 *     "fqdn": "neubot.mlab.mlab1.trn01.measurement-lab.org",
 *     "site": "trn01",
 *     "country": "IT"
 * }
 *
 */
type MlabResponse struct {
	Fqdn string `json:"fqdn"`
}

const MaximumBodyLength = 10 * 1024 * 1024

func write_response_json(w http.ResponseWriter, code int, body []byte) error {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json; encoding=utf-8")
	_, err := w.Write(body)
	if err != nil {
		log.Printf("cannot write http response body")
		return err
	}
	return nil
}

func new_response() *Response {
	var response Response
	response.Update = make(map[string]map[string]string)
	response.Available = make(map[string][]string)
	return &response
}

func query_mlabns(remote_addr string, our_response *Response) error {

	vector := strings.SplitN(remote_addr, ":", 2)
	if len(vector) != 2 {
		log.Printf("failed to split remote address")
		return errors.New("error splitting remote address")
	}

	// TODO: add support for IPv6 addresses

	res, err := http.Get("https://mlab-ns.appspot.com/neubot?ip=" + vector[0])
	if err != nil {
		log.Printf("failed to query mlab-ns")
		return err
	}

	reader := io.LimitReader(res.Body, MaximumBodyLength)
	response_body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Printf("error reading mlab's response body")
		return err
	}

	var mlab_response MlabResponse
	err = json.Unmarshal(response_body, &mlab_response)
	if err != nil {
		log.Printf("error umarshalling mlab's response body")
		// TODO: here we should also check for the case where the request body
		// was actually longer than MaximumBodyLength so leading to parse error
		return err
	}

	our_response.Available["speedtest"] = []string{
		"http://" + mlab_response.Fqdn + ":8080/speedtest",
	}
	our_response.Available["bittorrent"] = []string{
		"http://" + mlab_response.Fqdn + ":8080/",
	}
	return nil
}

func rendezvous(w http.ResponseWriter, r *http.Request) {

	reader := http.MaxBytesReader(w, r.Body, MaximumBodyLength)
	request_body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Printf("can't read request body")
		write_response_json(w, 500, EmptyJson)
		return
	}

	var request *Request = nil
	err = json.Unmarshal(request_body, &request)
	if err != nil {
		log.Printf("cannot unmarshal request body")
		write_response_json(w, 500, EmptyJson)
		return
	}

	if request.PrivacyInformed == 0 {
		log.Printf("user did not ready privacy policy")
		write_response_json(w, 500, EmptyJson)
		return
	}
	if request.PrivacyCanCollect == 0 {
		log.Printf("user did not give permission to collect results")
		write_response_json(w, 500, EmptyJson)
		return
	}
	if request.PrivacyCanShare == 0 {
		log.Printf("user did not give permission to publish results")
		write_response_json(w, 500, EmptyJson)
		return
	}

	our_response := new_response()
	err = query_mlabns(r.RemoteAddr, our_response)
	if err != nil {
		// Log message already printed by query_mlabns()
		write_response_json(w, 500, EmptyJson)
		return
	}

	our_response_body, err := json.Marshal(our_response)
	if err != nil {
		log.Printf("cannot marshal response body")
		write_response_json(w, 500, EmptyJson)
		return
	}

	write_response_json(w, 200, our_response_body)
}

func main() {

	// See http://technosophos.com/2013/09/14/using-gos-built-logger-log-syslog.html
	logwriter, err := syslog.New(syslog.LOG_NOTICE, "baudelaire")
	if err != nil {
		log.Fatal("cannot initialize syslog")
	}

	log.SetOutput(logwriter)
	log.SetFlags(0)
	log.Printf("baudelaire neubot master-server v0.0.1 starting up")

	http.HandleFunc("/rendezvous", rendezvous)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("cannot listen")
	}
}
