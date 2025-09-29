package app

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// httpErrorFragment writes a generic error response fragment and logs the error
func httpErrorFragment(w http.ResponseWriter, err error) {
	log.Errorf("error: %v", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<div class="error">Something went wrong. Please try again.</div>`)
}
