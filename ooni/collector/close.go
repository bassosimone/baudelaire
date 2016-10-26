// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

package collector

import (
	"github.com/julienschmidt/httprouter"
	"github.com/neubot/baudelaire/common"
	"log"
	"net/http"
	"os"
	"path"
)

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
