package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gookit/color"
	"github.com/pkg/errors"
	"github.com/snabb/sitemap"
)

func main() {
	if len(os.Args) < 2 {
		color.Red.Println("expected a command. Open help to view commands.")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--help", "help", "h":
		fmt.Println(`sites135 helps in creating sitemaps and 404 checks.

Example Commands
sites135 sitemap http://127.0.0.1:8080 https://example.com

sites135 c404 http://127.0.0.1:8080
		`)

	case "sitemap":
		if len(os.Args) != 4 {
			color.Red.Println("expecting 4 args. localaddr and publicaddr")
			os.Exit(1)
		}

		localAddr := os.Args[2]
		publicAddr := os.Args[3]

		links, err := getWebsiteLinks(localAddr, true)
		if err != nil {
			color.Red.Println(err.Error())
			os.Exit(1)
		}

		sm := sitemap.New()
		for _, link := range links {
			newAddr, err := url.JoinPath(publicAddr, link)
			if err != nil {
				fmt.Println(err)
				continue
			}
			tt := time.Now()
			sm.Add(&sitemap.URL{Loc: newAddr, LastMod: &tt})
		}

		var outStr string
		writer := bytes.NewBufferString(outStr)
		sm.WriteTo(writer)
		os.WriteFile("sitemap.xml", writer.Bytes(), 0777)

	case "c404":
		if len(os.Args) != 3 {
			color.Red.Println("expecting 3 args. addr")
			os.Exit(1)
		}

		localAddr := os.Args[2]
		links, err := getWebsiteLinks(localAddr, false)
		if err != nil {
			color.Red.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Gotten all links. Total: %d\n", len(links))

		errorCount := 0
		for _, link := range links {
			var newAddr string
			if strings.HasPrefix(link, "/") {
				newAddr, _ = url.JoinPath(localAddr, link)
			} else {
				newAddr = link
			}
			res, err := http.Get(newAddr)
			if err != nil {
				fmt.Println(err.Error())
			}
			defer res.Body.Close()
			if res.StatusCode == 404 {
				fmt.Printf("Not Found: %s\n", newAddr)
				errorCount += 1
			}
		}

		fmt.Printf("404 count: %d\n", errorCount)
	}

}

func isAbsoluteURL(testURL string) bool {
	parsedURL, err := url.Parse(testURL)
	if err != nil {
		return false
	}
	return parsedURL.IsAbs()
}

func getLinksForAPage(addr string) ([]string, error) {
	ret := make([]string, 0)

	// Request the HTML page.
	res, err := http.Get(addr)
	if err != nil {
		return nil, errors.Wrap(err, "http error")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("addr: %s status code error: %d %s", addr, res.StatusCode, res.Status))
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "html error")
	}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		val, ok := s.Attr("href")
		if !ok {
			return
		}
		ret = append(ret, val)
	})

	return ret, nil
}

func getWebsiteLinks(addr string, localOnly bool) ([]string, error) {
	ret, err := getLinksForAPage(addr)
	if err != nil {
		return nil, err
	}

	visited := []string{"/"}

	for {
		toWorkOnLink := ""
		endLoopIndex := 0
		for i, aLink := range ret {
			if !slices.Contains(visited, aLink) {
				toWorkOnLink = aLink
				break
			}
			endLoopIndex = i
		}

		if endLoopIndex == len(ret)-1 {
			break
		}

		visited = append(visited, toWorkOnLink)
		if !isAbsoluteURL(toWorkOnLink) {
			newAddr, err := url.JoinPath(addr, toWorkOnLink)
			if err != nil {
				fmt.Println(err)
				continue
			}
			innerRet, err := getLinksForAPage(newAddr)
			if err != nil {
				fmt.Println(err)
				continue
			}

			for _, aLink2 := range innerRet {
				if !slices.Contains(ret, aLink2) {
					ret = append(ret, aLink2)
				}
			}

		}
	}

	if localOnly {
		return getLocalLinks(addr, ret)
	} else {
		return visited, nil
	}
}

func getLocalLinks(addr string, links []string) ([]string, error) {
	ret := make([]string, 0)
	for _, link := range links {
		if strings.HasPrefix(link, "/") {
			ret = append(ret, link)
		}
		if strings.HasPrefix(link, addr) {
			ret = append(ret, link)
		}
	}

	return ret, nil
}
