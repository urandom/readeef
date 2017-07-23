package api

import "net/http"

type features struct {
	I18N       bool
	Search     bool
	Extractor  bool
	ProxyHTTP  bool
	Popularity bool
}

func featuresHandler(features features) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		args{"features": features}.WriteJSON(w)
	}
}
