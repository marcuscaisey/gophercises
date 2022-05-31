package sitemap

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/marcuscaisey/gophercises/4-link/links"
)

type extractResult struct {
	link *url.URL
	err  error
}

func Generate(rawURL string, numWorkers int, rate int) (Sitemap, error) {
	rootURL, err := url.Parse(rawURL)
	if err != nil {
		return Sitemap{}, fmt.Errorf("parse URL: %s", err)
	}
	if rootURL.Scheme == "" {
		log.Printf("%s has no scheme, defaulting to http", rootURL)
		rootURL, _ = rootURL.Parse("http://" + rawURL)
	}

	done := make(chan struct{})
	defer close(done)

	urlsToCrawl := make(chan *url.URL)
	var activeURLCounter sync.WaitGroup

	scheduledURLs := schedule(done, urlsToCrawl)
	extractResults := extract(done, scheduledURLs, &activeURLCounter, numWorkers, rate)
	return collect(rootURL, extractResults, urlsToCrawl, &activeURLCounter)
}

func schedule(done <-chan struct{}, in <-chan *url.URL) <-chan *url.URL {
	scheduledURLs := make(chan *url.URL)
	queueChan := make(chan *url.URL)
	queueingDone := make(chan struct{})

	go func() {
		for url := range in {
			select {
			case scheduledURLs <- url:
			case queueChan <- url:
			case <-done:
				return
			}
		}
		close(scheduledURLs)
		close(queueingDone)
	}()

	go func() {
		queue := &linkedQueue[*url.URL]{}
		var nextURL *url.URL
		for {
			if !queue.Empty() && nextURL == nil {
				nextURL = queue.Dequeue()
			}
			if nextURL != nil {
				select {
				case scheduledURLs <- nextURL:
					nextURL = nil
				case <-done:
					return
				default:
				}
			}
			select {
			case url := <-queueChan:
				queue.Enqueue(url)
			case <-done:
			case <-queueingDone:
				return
			default:
			}
		}
	}()

	return scheduledURLs
}

func extract(done chan struct{}, urls <-chan *url.URL, activeURLCounter *sync.WaitGroup, numWorkers int, rate int) <-chan extractResult {
	extractResults := make(chan extractResult)
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		workerNum := i + 1
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(time.Second / time.Duration(rate))
			for url := range urls {
				log.Printf("[Worker %d] Crawling %s.", workerNum, url)
				links, err := crawl(url)
				if err != nil {
					select {
					case extractResults <- extractResult{err: fmt.Errorf("crawl: %s", err)}:
					case <-done:
					}
					return
				}
				activeURLCounter.Add(len(links))
				for _, link := range links {
					select {
					case extractResults <- extractResult{link: link}:
					case <-done:
						return
					}
				}
				activeURLCounter.Done()
				select {
				case <-ticker.C:
				case <-done:
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(extractResults)
	}()

	return extractResults
}

func crawl(pageURL *url.URL) ([]*url.URL, error) {
	resp, err := http.Get(pageURL.String())
	if err != nil {
		return nil, fmt.Errorf("GET page: %s", err)
	}
	defer resp.Body.Close()
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		log.Printf("non-2xx status code for %s: %s", pageURL, resp.Status)
		return nil, nil
	}

	links, err := links.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse links from %s: %s", pageURL, err)
	}

	var filteredLinks []*url.URL
	for _, link := range links {
		parsedLink, err := pageURL.Parse(link.Href)
		if err != nil {
			log.Printf("Error parsing link href from %s: %s", pageURL, err)
			continue
		}
		parsedLink.Fragment = ""
		parsedLink.Path = strings.TrimSuffix(parsedLink.Path, "/")

		sameHost := parsedLink.Host == pageURL.Host
		hasHTTPScheme := strings.HasPrefix(parsedLink.Scheme, "http")
		if sameHost && hasHTTPScheme {
			filteredLinks = append(filteredLinks, parsedLink)
		}
	}

	return filteredLinks, nil
}

func collect(rootURL *url.URL, extractResults <-chan extractResult, urlsToCrawl chan<- *url.URL, activeURLCounter *sync.WaitGroup) (Sitemap, error) {
	seenURLs := newHashSet[string]()
	seenURLs.Add(rootURL.String())
	sitemap := Sitemap{
		URLs: []URL{
			{Loc: rootURL.String()},
		},
	}

	activeURLCounter.Add(1)
	urlsToCrawl <- rootURL

	go func() {
		activeURLCounter.Wait()
		close(urlsToCrawl)
	}()

	for res := range extractResults {
		if res.err != nil {
			return Sitemap{}, res.err
		}
		if seenURLs.Exists(res.link.String()) {
			activeURLCounter.Done()
			continue
		}
		seenURLs.Add(res.link.String())
		sitemap.URLs = append(sitemap.URLs, URL{Loc: res.link.String()})
		urlsToCrawl <- res.link
	}

	return sitemap, nil
}
