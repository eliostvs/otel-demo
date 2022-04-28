package web

import (
	"net/http"
	"strings"
)

type FilterURLs []string

func (f FilterURLs) Use(r *http.Request) bool {
	for _, url := range f {
		if r.URL.Path == url || strings.HasPrefix(r.URL.Path, url) {
			return false
		}
	}

	return true
}
