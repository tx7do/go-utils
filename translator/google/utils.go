package google

import "net/url"

// javascript "encodeURI()"
// so we embed js to our golang program
func encodeURI(s string) string {
	return url.QueryEscape(s)
}
