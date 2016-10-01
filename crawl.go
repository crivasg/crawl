/*
   Build a Web Crawler in Go
   https://jdanger.com/build-a-web-crawler-in-go.html

   A Simple Web Scraper in Go
   http://schier.co/blog/2015/04/26/a-simple-web-scraper-in-go.html

   Go web page scraper
   http://www.reddit.com/r/golang/comments/37jvaz/go_web_page_scraper/

*/

package main

import (
	"flag"
	"fmt"
	"io"
	//"io/ioutil" // 'ioutil' will help us print pages to the screen
	"golang.org/x/net/html"
	"net/http"
	"os"
)

func Usage() {
	fmt.Fprintf(os.Stderr, "usage: crawl http://example.com/path/file.html\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func CollectLinks(httpBody io.Reader) []string {

	links := make([]string, 0)
	page := html.NewTokenizer(httpBody)
	for {
		tokenType := page.Next()
		if tokenType == html.ErrorToken {
			return links
		}
		token := page.Token()
		if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					links = append(links, attr.Val)
				}
			}
		}
	}

}

func CollectLinks2(httpBody io.Reader) []string {
	// http://golang-examples.tumblr.com/post/47426518779/parse-html

	links := make([]string, 0)
	script_token := 0
	page := html.NewTokenizer(httpBody)
	for {
		tokenType := page.Next()

		if tokenType == html.ErrorToken {
			return links
		}
		token := page.Token()

		switch tokenType {
		case html.StartTagToken: // <tag>
			if token.DataAtom.String() == "script" {
				fmt.Printf("script = %v\n", token)
				script_token = 1
				continue
			}
		case html.TextToken: // text between start and end tag
			if script_token == 1 {
				fmt.Printf("\tattr = %v\n", token)
			}
		case html.EndTagToken: // </tag>
			if script_token == 1 {
				fmt.Printf("----------------------------------------------------------\n")
			}
			script_token = 0
		case html.SelfClosingTagToken: // <tag/>
		}

		/*
			if tokenType == html.StartTagToken && token.DataAtom.String() == "script" {
				fmt.Printf("token = %v\n", token.Data)
				for _, attr := range token.Attr {
					fmt.Printf("\tattr = %v\n", attr)
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}
				}
			}
		*/
	}

}

func RetrieveDataFrom(uri string) (io.Reader, error) {

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return resp.Body, nil

}

func main() {

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		Usage()
	}
	/*
		data, err := RetrieveDataFrom(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving the data from %s\n", args[0])
			os.Exit(1)
		}
		fmt.Println(data)

	*/

	resp, err := http.Get(args[0])
	if err != nil {
		os.Exit(2)
	}

	defer resp.Body.Close()

	links := CollectLinks2(resp.Body)

	for _, link := range links { // 'for' + 'range' in Go is like .each in Ruby or
		fmt.Println(string(link)) // an iterator in many other languages.
	}

}
