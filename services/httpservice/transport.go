package httpservice

import (
	"net/http"
)

type MasterTransport struct {
	// Transport is the underlying http.RoundTripper that is actually used to make the request
	Transport http.RoundTripper
}

func (t *MasterTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", defaultUserAgent)

	return t.Transport.RoundTrip(req)
}
