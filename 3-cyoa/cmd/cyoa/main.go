package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/marcuscaisey/gophercises/3-cyoa/cyoa"
)

var arcFile = flag.String("story-arcs", "", "JSON file with the map of arc name to story arcs in")
var port = flag.Uint("port", 8080, "Port to serve on")

var stderr = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	flag.Parse()

	if *arcFile == "" {
		exit("-story-arcs must be provided")
	}

	arcFileReader, err := os.Open(*arcFile)
	if err != nil {
		exit("open story arcs file: %s", err)
	}

	arcNameToArc := map[string]cyoa.StoryArc{}
	if err := json.NewDecoder(arcFileReader).Decode(&arcNameToArc); err != nil {
		exit("decode story arcs file from JSON: %s", err)
	}

	handler := cyoa.MustNewHandler(arcNameToArc)

	log.Printf("serving on :%d", *port)
	http.Handle("/", handler)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), handler)
}

func exit(msg string, args ...any) {
	stderr.Printf(msg, args...)
	os.Exit(1)
}
