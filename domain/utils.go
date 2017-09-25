package domain

import "strings"

var hopByHopHeaders = []string {
	"Content-Length",
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"TE",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

func isHopByHopHeader(header string) bool {
	for _, value := range hopByHopHeaders {
		if strings.ToLower(header) == strings.ToLower(value) {
			return true
		}
	}
	return false
}