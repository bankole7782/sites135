package main

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

func main() {
	links, err := getWebsiteLinks("http://127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	fmt.Println(links)
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

		newAddr, err := url.JoinPath(addr, toWorkOnLink)
		if err != nil {
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
