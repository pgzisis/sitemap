package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pgzisis/html-link-parser/htmlparser"
)

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

type urlset struct {
	Urls  []loc  `xml:"url"`
	Xmlns string `xml:"xmlns,attr"`
}

type loc struct {
	Value string `xml:"loc"`
}

func main() {
	urlName := flag.String("url", "https://gophercises.com", "url to create sitemap")
	maxDepth := flag.Int("depth", 3, "the maximum number of links deep to traverse")
	flag.Parse()

	pages := bfs(*urlName, *maxDepth)
	toXml := urlset{
		Xmlns: xmlns,
	}
	for _, p := range pages {
		toXml.Urls = append(toXml.Urls, loc{p})
	}

	fmt.Print(xml.Header)
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("", "  ")
	if err := enc.Encode(toXml); err != nil {
		panic(err)
	}
	fmt.Println()
}

func bfs(urlStr string, maxDepth int) []string {
	seen := make(map[string]struct{})
	var q map[string]struct{}
	nq := map[string]struct{}{
		urlStr: {},
	}

	for i := 0; i < maxDepth; i++ {
		q, nq = nq, make(map[string]struct{})
		for url, _ := range q {
			if _, ok := seen[url]; ok {
				continue
			}
			seen[url] = struct{}{}
			for _, link := range getPages(&url) {
				nq[link] = struct{}{}
			}
		}
	}

	var result []string
	for url, _ := range seen {
		result = append(result, url)
	}

	return result
}

func getPages(urlName *string) []string {
	resp, err := http.Get(*urlName)
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()

	reqUrl := resp.Request.URL
	baseUrl := &url.URL{
		Scheme: reqUrl.Scheme,
		Host:   reqUrl.Host,
	}
	base := baseUrl.String()

	hrefs := getHrefs(resp, base)
	filteredHrefs := filter(*urlName, hrefs)

	return filteredHrefs
}

func getHrefs(resp *http.Response, base string) []string {
	links, _ := htmlparser.Parse(resp.Body)

	var hrefs []string
	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			hrefs = append(hrefs, base+l.Href)
		case strings.HasPrefix(l.Href, "http"):
			hrefs = append(hrefs, l.Href)
		}
	}

	return hrefs
}

func filter(base string, links []string) []string {
	var result []string
	for _, l := range links {
		if strings.HasPrefix(l, base) {
			result = append(result, l)
		}
	}

	return result
}
