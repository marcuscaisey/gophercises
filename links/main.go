package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/marcuscaisey/gophercises/links/links"
)

var file = flag.String("file", "", "HTML file to parse")

func main() {
	flag.Parse()

	if *file == "" {
		log.Fatalln("-file must be provided")
	}

	f, err := os.Open(*file)
	if err != nil {
		log.Fatalf("open file: %s", err)
	}

	links, err := links.Parse(f)
	if err != nil {
		log.Fatalf("parse links: %s", err)
	}
	for _, link := range links {
		fmt.Println(link)
	}
}
