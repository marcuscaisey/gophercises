package urlshort

import (
	"fmt"
	"net/http"

	"gopkg.in/yaml.v3"
)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToURLs map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mappedURL, ok := pathsToURLs[r.URL.Path]
		if ok {
			http.Redirect(w, r, mappedURL, http.StatusFound)
			return
		}
		fallback.ServeHTTP(w, r)
	}
}

type yamlHandlerConfig []struct {
	Path string `yaml:"path"`
	URL  string `yaml:"url"`
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//     - path: /some-path
//       url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var config yamlHandlerConfig
	if err := yaml.Unmarshal(yml, &config); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}
	pathsToURLs := make(map[string]string, len(config))
	for _, c := range config {
		pathsToURLs[c.Path] = c.URL
	}
	return MapHandler(pathsToURLs, fallback), nil
}
