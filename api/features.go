package api

import "net/http"

type features struct {
	Search     bool `json:"search,omitempty"`
	Extractor  bool `json:"extractor,omitempty"`
	ProxyHTTP  bool `json:"proxyHTTP,omitempty"`
	Popularity bool `json:"popularity,omitempty"`
}

func featuresHandler(features features) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		args{"features": features}.WriteJSON(w)
	}
}
