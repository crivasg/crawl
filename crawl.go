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
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"io/ioutil" // 'ioutil' will help us print pages to the screen
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type Program struct {
	Type      string      `json:"type"`
	Full      []string    `json:"full"`
	Segments  []string    `json:"segments"`
	AudioData []AudioData `json:"audioData"`
}

type AudioData struct {
	Uid         string `json:"uid"`
	Available   bool   `json:"available"`
	Duration    int    `json:"duration"`
	Title       string `json:"title"`
	AudioUrl    string `json:"audioUrl"`
	StoryUrl    string `json:"storyUrl"`
	Slug        string `json:"slug"`
	Program     string `json:"program"`
	Affiliation string `json:"affiliation"`
	Song        string `json:"song"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	Track       int    `json:"track"`
	Type        string `json:"type"`
	Subtype     string `json:"subtype"`
}

func (i AudioData) String() string {

	return fmt.Sprintf("Title: %s\nURL: %s\nAudio URL: %s", i.Title, i.StoryUrl, i.AudioUrl)
	//return fmt.Sprintf("%s", i.AudioUrl)

}

//----------------------------------------------------------------------------------------

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

func CollectLinksATC(httpBody io.Reader) []string {
	// ATC == All things considered

	full_show_token := false

	links := make([]string, 0)
	page := html.NewTokenizer(httpBody)
	for {
		tokenType := page.Next()

		if tokenType == html.ErrorToken {
			return links
		}
		token := page.Token()

		switch tokenType {
		case html.StartTagToken: // <tag>
			if token.DataAtom.String() == "div" {
				for _, attr := range token.Attr {
					if attr.Key == "id" && attr.Val == "full-show" {
						full_show_token = true
					}
				}
			}
			if token.DataAtom.String() == "b" && full_show_token == true {
				// class="full-show-unavailable"
				for _, attr := range token.Attr {
					if attr.Key == "class" && attr.Val == "full-show-unavailable" {
						fmt.Printf("-------------> %s\n%s\n", attr.Key, attr.Val)
						full_show_token = false
					}
				}
				if full_show_token == false {
					return links
				}
				for _, attr := range token.Attr {
					//fmt.Printf("%s\n%s\n", attr.Key, attr.Val)
					//links = append(links, attr.Val)
					parseJSON(attr.Val)
				}
				full_show_token = false
			}
		case html.TextToken: // text between start and end tag
		case html.EndTagToken: // </tag>
		case html.SelfClosingTagToken: // <tag/>
		}
	}

	program := new(Program)
	reader := strings.NewReader(strings.Join(links, "\n"))

	err := json.NewDecoder(reader).Decode(program)
	if err != nil {
		fmt.Printf("error!!")
	}

	for _, episode := range program.AudioData {
		fmt.Printf("%v\n", episode)
	}

	return links

}

func parseJSON(b string) {

	program := new(Program)
	reader := strings.NewReader(b)

	err := json.NewDecoder(reader).Decode(program)
	if err != nil {
		fmt.Printf("%v\n\n\n", err)
	}

	for _, episode := range program.AudioData {
		fmt.Printf("%s\n", episode)
	}

}

func CollectLinksRadiolab(httpBody io.Reader) []string {
	// http://golang-examples.tumblr.com/post/47426518779/parse-html

	r, _ := regexp.Compile("\"http(s://|://).*mp3\"")

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
				//fmt.Printf("script = %v\n", token)
				script_token = 1
				continue
			}
		case html.TextToken: // text between start and end tag
			if script_token == 1 {
				match, _ := regexp.MatchString("embed_audio_buttons", token.Data)
				if match == true {
					//fmt.Printf("%s\n", strings.Trim(token.Data, "\t\n "))
					link := r.FindString(strings.Trim(token.Data, "\t\n "))
					link = strings.Replace(link, "\"", "", -1)
					//fmt.Printf("-> %s\n",link)
					links = append(links, link)
				}
			}
		case html.EndTagToken: // </tag>
			script_token = 0
		case html.SelfClosingTagToken: // <tag/>
			script_token = 0
		}
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

func test_atc_json(filename string) {
	//
	b, _ := ioutil.ReadFile(filename)
	//fmt.Printf("-> %s\n", b)

	program := new(Program)
	reader := strings.NewReader(string(b))

	err := json.NewDecoder(reader).Decode(program)
	if err != nil {
		fmt.Printf("%v\n\n\n", err)
	}

	for _, episode := range program.AudioData {
		fmt.Printf("%s\n", episode)
	}

}

func main() {

	//test_atc_json("/Users/crivas/Desktop/atc.json")
	//return

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

	//links := CollectLinksRadiolab(resp.Body)
	links := CollectLinksATC(resp.Body)

	for _, link := range links { // 'for' + 'range' in Go is like .each in Ruby or
		fmt.Println(string(link)) // an iterator in many other languages.
	}

}
