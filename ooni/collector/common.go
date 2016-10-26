// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

package collector

import (
	"errors"
	"fmt"
	"log"
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

func map_report_id_to_path(report_id string) (string, error) {
	if !matches_regexp(report_id_re, report_id) {
		// Log message already printed by the matches_regexp() func
		return "", errors.New("matches_regexp() failed")
	}
	fpath := path.Join("data", report_id)
	statbuf, err := os.Lstat(fpath)
	if err != nil {
		log.Printf("cannot stat report")
		return "", err
	}
	if !statbuf.Mode().IsRegular() {
		log.Printf("not a regular file")
		return "", err
	}

	return fpath, nil
}
