package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

func GetRootPath() (string, error) {
	hd, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "os error")
	}

	dd := os.Getenv("SNAP_USER_COMMON")
	if strings.HasPrefix(dd, filepath.Join(hd, "snap", "go")) || dd == "" {
		dd = filepath.Join(hd, "sites135")
		os.MkdirAll(dd, 0777)
	}

	return dd, nil
}

func isAbsoluteURL(testURL string) bool {
	parsedURL, err := url.Parse(testURL)
	if err != nil {
		return false
	}
	return parsedURL.IsAbs()
}

func getLinksForAPage(addr string, withAssetsLinks bool) ([]string, error) {
	ret := make([]string, 0)
	// check if this is not a html page
	parsedURL, err := url.Parse(addr)
	if err != nil {
		return ret, err
	}
	notNeededExtensions := []string{".png", ".jpeg", ".gif", ".mp4"}
	for _, ext := range notNeededExtensions {
		if strings.HasSuffix(parsedURL.Path, ext) {
			return ret, nil
		}
	}

	addrEsc, _ := url.QueryUnescape(addr)

	// Request the HTML page.
	res, err := http.Get(addrEsc)
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

	if withAssetsLinks {
		doc.Find("img").Each(func(i int, s *goquery.Selection) {
			val, ok := s.Attr("src")
			if !ok {
				return
			}
			ret = append(ret, val)
		})

		doc.Find("link").Each(func(i int, s *goquery.Selection) {
			val, ok := s.Attr("href")
			if !ok {
				return
			}
			ret = append(ret, val)
		})

		doc.Find("script").Each(func(i int, s *goquery.Selection) {
			val, ok := s.Attr("src")
			if !ok {
				return
			}
			ret = append(ret, val)
		})

	}

	return ret, nil
}

func getWebsiteLinks(addr string, localOnly, withAssetsLinks bool) ([]string, error) {
	ret, err := getLinksForAPage(addr, withAssetsLinks)
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
			innerRet, err := getLinksForAPage(newAddr, withAssetsLinks)
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
		return ret, nil
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
