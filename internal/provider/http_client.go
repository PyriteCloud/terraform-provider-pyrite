package provider

import (
	"net/http"
)

type AuthTransport struct {
	token   string
	wrapped http.RoundTripper
}

func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Api-Key", t.token)
	return t.wrapped.RoundTrip(req)
}
