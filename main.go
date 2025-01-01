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
		`)

	case "sitemap":
		if len(os.Args) != 4 {
			color.Red.Println("expecting 4 args. localaddr and publicaddr")
			os.Exit(1)
		}

		localAddr := os.Args[2]
		publicAddr := os.Args[3]

		links, err := getLocalWebsiteLinks(localAddr)
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
		fmt.Println(writer.String())

	case "c404":

	}

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
		return nil, errors.New(fmt.Sprintf("status code error: %d %s", res.StatusCode, res.Status))
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

func getWebsiteLinks(addr string) ([]string, error) {
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
		if !strings.HasPrefix(toWorkOnLink, "/") {
			continue
		}
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

		continue
	}

	return ret, nil
}

func getLocalWebsiteLinks(addr string) ([]string, error) {
	links, err := getWebsiteLinks(addr)
	if err != nil {
		return nil, err
	}

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
