// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

package collector

import (
	"github.com/julienschmidt/httprouter"
	"github.com/neubot/baudelaire/common"
	"net/http"
)

func Create(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	common.WriteResponseJson(w, 500, common.EmptyJson)
}

func Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	common.WriteResponseJson(w, 500, common.EmptyJson)
}

func Close(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	common.WriteResponseJson(w, 500, common.EmptyJson)
}
