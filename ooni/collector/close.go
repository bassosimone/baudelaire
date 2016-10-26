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

	report_id := ps.ByName("id")
	fpath, err := map_report_id_to_path(report_id)
	if err != nil {
		// Log message already printed by the above function
		common.WriteResponseJson(w, 500, common.EmptyJson)
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
