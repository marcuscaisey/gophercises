package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/marcuscaisey/gophercises/urlshort/v2/errors"
	"github.com/marcuscaisey/gophercises/urlshort/v2/errors/codes"
)

type errorHandlingMux struct {
	serveMux *http.ServeMux
}

func newErrorHandlingMux() *errorHandlingMux {
	return &errorHandlingMux{serveMux: http.NewServeMux()}
}

func (m *errorHandlingMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.serveMux.ServeHTTP(w, r)
}

func (m *errorHandlingMux) Handle(allowedMethod string, pattern string, handler func(http.ResponseWriter, *http.Request) error) {
	f := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowedMethod {
			w.Header().Add("Allow", allowedMethod)
			w.WriteHeader(http.StatusMethodNotAllowed)
			writeError(w, fmt.Sprintf("Method %s is not allowed, use %s.", r.Method, allowedMethod))
			return
		}
		if err := handler(w, r); err != nil {
			handleError(w, err)
		}
	}
	m.serveMux.Handle(pattern, http.HandlerFunc(f))
}

func handleError(w http.ResponseWriter, err error) {
	code := errors.Code(err)

	switch code {
	case codes.Internal:
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	case codes.AlreadyExists:
		w.WriteHeader(http.StatusConflict)
	case codes.NotFound:
		w.WriteHeader(http.StatusNotFound)
	case codes.BadRequest:
		w.WriteHeader(http.StatusBadRequest)
	}

	var msg string
	if code == codes.Internal {
		msg = "An internal server error has occurred."
	} else {
		msg = errors.Message(err)
	}

	writeError(w, msg)
}

func writeError(w http.ResponseWriter, msg string) {
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: msg,
	})
}
