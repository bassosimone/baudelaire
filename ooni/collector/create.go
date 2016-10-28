// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

package collector

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/neubot/baudelaire/common"
	"github.com/neubot/bernini"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

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
		string(bernini.RandByteMaskingImproved(50))
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
