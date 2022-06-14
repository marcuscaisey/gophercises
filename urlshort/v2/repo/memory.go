package repo

import (
	"github.com/marcuscaisey/gophercises/urlshort/v2/errors"
	"github.com/marcuscaisey/gophercises/urlshort/v2/errors/codes"
)

type InMemoryURLRepository struct {
	shortPathToLongURL map[string]string
}

func NewInMemoryURLRepository() InMemoryURLRepository {
	return InMemoryURLRepository{
		shortPathToLongURL: map[string]string{},
	}
}

func (r InMemoryURLRepository) Create(shortPath, longURL string) error {
	_, found := r.shortPathToLongURL[shortPath]
	if found {
		return errors.New(codes.AlreadyExists)
	}
	r.shortPathToLongURL[shortPath] = longURL
	return nil
}

func (r InMemoryURLRepository) Get(shortPath string) (string, error) {
	longURL, found := r.shortPathToLongURL[shortPath]
	if !found {
		return "", errors.New(codes.NotFound)
	}
	return longURL, nil
}
