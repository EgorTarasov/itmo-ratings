package scrapper

import "net/http"

type (
	client interface {
		Do(req *http.Request) (*http.Response, error)
	}
)
