// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

package collector

import (
	"fmt"
	"log"
	"regexp"
	"time"
)

const version_re = "^[0-9A-Za-z_.-]+$"
const name_re = "^[a-zA-Z0-9_ -]+$"
const probe_asn_re = "^AS[0-9]+$"
const probe_cc_re = "^[A-Z]{2}$"
const report_id_re = "^[0-9]{8}T[0-9]{6}Z_AS[0-9]{1,5}_[A-Za-z]{50}$"

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
