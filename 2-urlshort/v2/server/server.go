package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/marcuscaisey/gophercises/2-urlshort/v2/errors"
	"github.com/marcuscaisey/gophercises/2-urlshort/v2/errors/codes"
)

type URLRepository interface {
	Create(longURL, shortPath string) error
	Get(shortPath string) (string, error)
}

type Server struct {
	mux     *http.ServeMux
	urlRepo URLRepository
}

func New(urlRepo URLRepository) *Server {
	s := &Server{
		urlRepo: urlRepo,
	}
	return s
}

func (s *Server) Run(port uint) error {
	mux := newErrorHandlingMux()
	mux.Handle(http.MethodPost, "/shorten", s.shorten)
	mux.Handle(http.MethodGet, "/", s.redirect)

	address := fmt.Sprintf(":%d", port)
	log.Printf("Serving on %s.", address)

	err := http.ListenAndServe(address, mux)
	if err != nil {
		return fmt.Errorf("listen and serve on %q: %w", address, err)
	}
	return nil
}

func (s *Server) shorten(w http.ResponseWriter, r *http.Request) error {
	w.Header().Add("Content-Type", "application/json")

	var shortenReq struct {
		ShortPath string `json:"short_path"`
		LongURL   string `json:"long_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&shortenReq); err != nil {
		return errors.New("Request is not valid JSON.", codes.BadRequest, err)
	}

	if shortenReq.LongURL == "" {
		return errors.New(`Request must contain long_url field.`, codes.BadRequest)
	}
	if shortenReq.ShortPath == "" {
		shortenReq.ShortPath = "/" + generateBase64(12)
	} else if shortenReq.ShortPath == "/" {
		return errors.New("short_path must contain at least one character", codes.BadRequest)
	} else if shortenReq.ShortPath[0:1] != "/" {
		shortenReq.ShortPath = "/" + shortenReq.ShortPath
	}

	if err := s.urlRepo.Create(shortenReq.ShortPath, shortenReq.LongURL); err != nil {
		if errors.Code(err) == codes.AlreadyExists {
			return errors.New(fmt.Sprintf("short_path %s has already been taken.", shortenReq.ShortPath), err)
		}
		return fmt.Errorf("create url: %w", err)
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(shortenReq); err != nil {
		return fmt.Errorf("encode response: %+v to JSON: %w", shortenReq, err)
	}

	return nil
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) error {
	shortPath := r.URL.Path
	if shortPath == "/" {
		http.NotFound(w, r)
		return nil
	}

	longURL, err := s.urlRepo.Get(shortPath)
	if err != nil {
		if errors.Code(err) == codes.NotFound {
			return errors.New(fmt.Sprintf("No long URL found for short_path: %s", shortPath), err)
		}
		return fmt.Errorf("get long URL: %w", err)
	}

	http.Redirect(w, r, longURL, http.StatusFound)

	return nil
}

func generateBase64(length int) string {
	randomBytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, randomBytes); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(randomBytes)
}
