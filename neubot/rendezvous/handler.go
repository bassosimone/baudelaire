// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

package rendezvous

import (
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"github.com/neubot/baudelaire/common"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// See README.md
type request_t struct {
	Accept            []string `json:"accept"`
	PrivacyCanCollect int      `json:"privacy_can_collect"`
	PrivacyCanShare   int      `json:"privacy_can_share"`
	PrivacyInformed   int      `json:"privacy_informed"`
	Version           string   `json:"version"`
}

// See README.md
type response_t struct {
	Update    map[string]map[string]string `json:"update"`
	Available map[string][]string          `json:"available"`
}

// See README.md
type mlab_response_t struct {
	Fqdn string `json:"fqdn"`
	Country string `json:"country"`
}

func new_response() *response_t {
	var response response_t
	response.Update = make(map[string]map[string]string)
	response.Available = make(map[string][]string)
	return &response
}

func query_mlabns(remote_addr string, our_response *response_t) error {

	log.Printf("/rendezvous request from %s", remote_addr)

	if strings.HasPrefix(remote_addr, "[") {
		// Note: master.neubot.org is IPv4-only therefore by excluding this
		// case we are only striving for implementation correctness.
		//
		// In the hypothesis we need to do IPv6, the transformation we would
		// need is like: [12::3:4]:567 => [12::3:4] => 12::3:4
		log.Printf("we do not support IPv6 input")
		return errors.New("IPv6 not supported")
	}

	vector := strings.SplitN(remote_addr, ":", 2)
	if len(vector) != 2 {
		log.Printf("failed to split remote address")
		return errors.New("error splitting remote address")
	}

	res, err := http.Get("https://mlab-ns.appspot.com/neubot?ip=" + vector[0])
	if err != nil {
		log.Printf("failed to query mlab-ns")
		return err
	}

	reader := io.LimitReader(res.Body, common.MaximumBodyLength)
	response_body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Printf("error reading mlab's response body")
		return err
	}

	var mlab_response mlab_response_t
	err = json.Unmarshal(response_body, &mlab_response)
	if err != nil {
		log.Printf("error umarshalling mlab's response body")
		// TODO: here we should also check for the case where the request body
		// was actually longer than MaximumBodyLength so leading to parse error
		return err
	}

	log.Printf("rendezvous_server: %s[%s] -> %s", remote_addr,
			mlab_response.Country, mlab_response.Fqdn)

	our_response.Available["speedtest"] = []string{
		"http://" + mlab_response.Fqdn + ":8080/speedtest",
	}
	our_response.Available["bittorrent"] = []string{
		"http://" + mlab_response.Fqdn + ":8080/",
	}
	return nil
}

func Handle(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	reader := http.MaxBytesReader(w, r.Body, common.MaximumBodyLength)
	request_body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Printf("can't read request body")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	var request *request_t = nil
	err = json.Unmarshal(request_body, &request)
	if err != nil {
		log.Printf("cannot unmarshal request body")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	if request.PrivacyInformed == 0 {
		log.Printf("user did not ready privacy policy")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}
	if request.PrivacyCanCollect == 0 {
		log.Printf("user did not give permission to collect results")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}
	if request.PrivacyCanShare == 0 {
		log.Printf("user did not give permission to publish results")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	our_response := new_response()
	err = query_mlabns(r.RemoteAddr, our_response)
	if err != nil {
		// Log message already printed by query_mlabns()
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	our_response_body, err := json.Marshal(our_response)
	if err != nil {
		log.Printf("cannot marshal response body")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	common.WriteResponseJson(w, 200, our_response_body)
}
