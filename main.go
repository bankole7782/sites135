package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/snabb/sitemap"
)

func main() {
	if len(os.Args) < 2 {
		color.Red.Println("expected a command. Open help to view commands.")
		os.Exit(1)
	}

	rootPath, err := GetRootPath()
	if err != nil {
		color.Red.Println(rootPath)
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

		links, err := getWebsiteLinks(localAddr, true, false)
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
		links, err := getWebsiteLinks(localAddr, false, false)
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
			res, err := http.Head(newAddr)
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

	case "wdl":
		if len(os.Args) != 4 {
			color.Red.Println("expecting 3 args: webaddr and projectname")
			os.Exit(1)
		}

		inputAddr := os.Args[2]
		projectName := os.Args[3]

		links, err := getWebsiteLinks(inputAddr, true, true)
		if err != nil {
			color.Red.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Gotten all links. Total: %d\n", len(links))
		fmt.Println(links)

		baseOutPath := filepath.Join(rootPath, projectName)
		os.MkdirAll(baseOutPath, 0777)
		for _, link := range links {
			var newAddr string
			if strings.HasPrefix(link, "/") {
				newAddr, _ = url.JoinPath(inputAddr, link)
			} else {
				newAddr = link
			}
			newAddrEsc, _ := url.QueryUnescape(newAddr)
			res, err := http.Get(newAddrEsc)
			if err != nil {
				fmt.Println(err.Error())
			}
			defer res.Body.Close()
			if res.StatusCode == 404 {
				fmt.Printf("Not Found: %s\n", newAddrEsc)
				continue
			}

			bodyOut, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Println(err)
				return
			}

			if link == "/" {
				link = "/index"
			}
			toWritePath := filepath.Join(baseOutPath, link)
			if !strings.Contains(toWritePath, ".") {
				toWritePath += ".html"
			}
			lastSym := strings.LastIndex(toWritePath, "?")
			if lastSym != -1 {
				toWritePath = toWritePath[:lastSym]
			}
			err = os.MkdirAll(filepath.Dir(toWritePath), 0777)
			if err != nil {
				fmt.Println(err)
				fmt.Println(newAddr)
				continue
			}

			os.WriteFile(toWritePath, bodyOut, 0777)

			// time.Sleep(1 * time.Second)
		}

	case "wdls":
		if len(os.Args) != 4 {
			color.Red.Println("expecting args: projectname")
			os.Exit(1)
		}
		projectName := os.Args[2]
		port := os.Args[3]
		serveWDl(projectName, port)

	default:
		color.Red.Println("Invalid arguments: ", os.Args[1:])
		os.Exit(1)
	}

}
