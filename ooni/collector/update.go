// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

package collector

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/neubot/baudelaire/common"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

type report_update_t struct {
	Content json.RawMessage `json:"content"`
	Format string `json:"format"`
}

type report_update_response_t struct {
	Status string `json:"status"`
}

func Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	report_id := ps.ByName("id")
	if !matches_regexp(report_id_re, report_id) {
		// Log message already printed by the matches_regexp() func
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}
	fpath := path.Join("data", report_id)
	statbuf, err := os.Lstat(fpath)
	if err != nil {
		log.Printf("cannot stat report")
		common.WriteResponseJson(w, 404, common.EmptyJson)
		return
	}
	if !statbuf.Mode().IsRegular() {
		log.Printf("not a regular file")
		common.WriteResponseJson(w, 404, common.EmptyJson)
		return
	}

	reader := http.MaxBytesReader(w, r.Body, common.MaximumBodyLength)
	request_body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Printf("can't read request body")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	var request *report_update_t = nil
	err = json.Unmarshal(request_body, &request)
	if err != nil {
		log.Printf("cannot unmarshal request body")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	if request.Format != "json" {
		log.Printf("unsupported request format")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	serialized, err := request.Content.MarshalJSON()
	if err != nil {
		log.Printf("cannot marshal the content field")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	fileptr, err := os.OpenFile(fpath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("cannot open the report file")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}
	defer fileptr.Close()
	if _, err = fileptr.Write(serialized); err != nil {
		log.Printf("cannot write entry to the report file")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}
	if _, err = fileptr.WriteString("\n"); err != nil {
		log.Printf("cannot write newline to the report file")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}
	if err = fileptr.Sync(); err != nil {
		log.Printf("cannot flush the report file")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	var response report_update_response_t
	response.Status = "success"

	response_body, err := json.Marshal(response)
	if err != nil {
		log.Printf("cannot marshal response body")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	common.WriteResponseJson(w, 200, response_body)
}
