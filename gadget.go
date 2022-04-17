package gadget

import "net/http"

// Parse attempts to retrieve the browser and system name from the set of
// headers.
func Parse(h http.Header) UserAgent {
	return ParseUA(h.Get("User-Agent"))
}
