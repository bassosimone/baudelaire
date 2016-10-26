// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

package collector

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/neubot/baudelaire/common"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"
)

const version_re = "^[0-9A-Za-z_.-]+$"
const name_re = "^[a-zA-Z0-9_ -]+$"
const probe_asn_re = "^AS[0-9]+$"
const probe_cc_re = "^[A-Z]{2}$"
const report_id_re = "^[0-9]{8}T[0-9]{6}Z_AS[0-9]{1,5}_[A-Za-z]{50}$"

type report_create_t struct {
	SoftwareName string `json:"software_name"`
	SoftwareVersion string `json:"software_version"`
	ProbeAsn string `json:"probe_asn"`
	ProbeCc string `json:"probe_cc"`
	TestName string `json:"test_name"`
	TestVersion string `json:"test_version"`
	DataFormatVersion string `json:"data_format_version"`
	TestStartTime string `json:"test_start_time"`
	InputHashes []string `json:"input_hashes"`
	TestHelper string `json:"test_helper"`
	Content string `json:"content"`
	ProbeIp string `json:"probe_ip"`
	Format string `json:"format"`
}

type report_create_response_t struct {
	BackendName string `json:"backend_name"`
	BackendVersion string `json:"backend_version"`
	ReportId string `json:"report_id"`
	TestHelperAddress string `json:"test_helper_address"`
	SupportedFormats []string `json:"supported_formats"`
}

type report_update_t struct {
	Content json.RawMessage `json:"content"`
	Format string `json:"format"`
}

type report_update_response_t struct {
	Status string `json:"status"`
}

func iso8601() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%04d%02d%02dT%02d%02d%02dZ", now.Year(), now.Month(),
		now.Day(), now.Hour(), now.Minute(), now.Second())
}

func matches_regexp(pattern string, s string) bool {
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		log.Printf("cannot compile regular expression: %s", pattern)
		return false
	}
	if !matched {
		log.Printf("regular expression does not match: %s", pattern)
		return false
	}

	return true
}

func Create(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	reader := http.MaxBytesReader(w, r.Body, common.MaximumBodyLength)
	request_body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Printf("can't read request body")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	var request *report_create_t = nil
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

	// XXX: some fields are not validated yet
	if !matches_regexp(name_re, request.SoftwareName) ||
		!matches_regexp(version_re, request.SoftwareVersion) ||
		!matches_regexp(probe_asn_re, request.ProbeAsn) ||
		!matches_regexp(probe_cc_re, request.ProbeCc) ||
		!matches_regexp(name_re, request.TestName) ||
		!matches_regexp(version_re, request.TestVersion) ||
		!matches_regexp(version_re, request.DataFormatVersion) {
		// Note: error already printed by regexp matching function
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	var response report_create_response_t
	response.BackendName = "baudelaire"
	response.BackendVersion = common.Version
	response.ReportId = iso8601() + "_" + request.ProbeAsn + "_" +
		string(common.RandByteMaskingImproved(50))
	response.SupportedFormats = []string{"json"}

	err = os.MkdirAll("data", 0755)
	if err != nil {
		log.Printf("cannot create 'data' directory: %s", err)
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}
	err = ioutil.WriteFile(path.Join("data", response.ReportId),
			[]byte(""), 0644)
	if err != nil {
		log.Printf("cannot create 'data/%s' file: %s", response.ReportId, err)
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	response_body, err := json.Marshal(response)
	if err != nil {
		log.Printf("cannot marshal response body")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	common.WriteResponseJson(w, 200, response_body)
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

func Close(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// XXX this duplicates code in Update()
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

	err = os.MkdirAll("archive", 0755)
	if err != nil {
		log.Printf("cannot create 'archive' directory: %s", err)
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}
	new_fpath := path.Join("archive", report_id)
	err = os.Rename(fpath, new_fpath)
	if err != nil {
		log.Printf("cannot archive the closed report")
		common.WriteResponseJson(w, 500, common.EmptyJson)
		return
	}

	common.WriteResponseJson(w, 200, common.EmptyJson)
}
