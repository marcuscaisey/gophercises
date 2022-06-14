package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/marcuscaisey/gophercises/sitemap"
)

var (
	numWorkers = flag.Int("workers", 16, "Number of parallel workers.")
	rate       = flag.Int("rate", 1, "Maximum rate of requests per second for each worker.")
)

func usage() {
	usage := `
Usage:
  sitemap [options] url

Options:`
	fmt.Fprintln(flag.CommandLine.Output(), usage)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(flag.CommandLine.Output(), "error: no URL provided")
		usage()
		os.Exit(2)
	}
	url := flag.CommandLine.Arg(0)

	sitemap, err := sitemap.Generate(url, *numWorkers, *rate)
	if err != nil {
		log.Fatalf("Failed to generate sitemap: %s", err)
	}

	sitemapXML, err := xml.MarshalIndent(sitemap, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal generated sitemap into XML: %s", err)
	}
	fmt.Printf("\n%s%s", xml.Header, sitemapXML)
}
