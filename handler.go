package main

import (
	"encoding/json"
	"net/http"
)

type APIHandler struct {
	Handler func(w http.ResponseWriter, r *http.Request) error
}

func (ah APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := ah.Handler(w, r)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		encoder := json.NewEncoder(w)

		switch e := err.(type) {
		case Error:
			w.WriteHeader(e.Status())
			encoder.Encode(struct {
				Status  int    `json:"code"`
				Message string `json:"message"`
			}{Status: e.Status(), Message: e.Error()})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(struct {
				Status  int    `json:"code"`
				Message string `json:"message"`
			}{Status: 500, Message: http.StatusText(http.StatusInternalServerError)})
		}
	}
}
