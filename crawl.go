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
	"encoding/xml"
	"fmt"
	"github.com/urfave/cli"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

//------------------------------TEMPLATES-------------------------------------------------

var funcMap = template.FuncMap{
	"basenameURL": basenameURL,
	"cleanURL":    cleanURL,
}

const templ = `wget -O {{basenameURL .URL}} {{cleanURL .URL}}
`

//------------------------------MODELS----------------------------------------------------

type Rss2 struct {
	Feed Feed `xml:"feed"`
}

type Feed struct {
	Id      string  `xml:"id"`
	Title   string  `xml:"title"`
	Updated string  `xml:"updated"`
	Links   []Link  `xml:"link"`
	Icon    string  `xml:"icon"`
	Entries []Entry `xml:"entry"`
}

type Entry struct {
	Updated string `xml:"updated"`
	Id      string `xml:"id"`
	Title   string `xml:"title"`
	Summary string `xml:"summary"`
	Link    Link   `xml:"link"`
	Content string `xml:"content"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}

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

func (i AudioData) FormatAudioUrl() (string, error) {

	u, err := url.Parse(i.AudioUrl)
	if err != nil {
		return "", err
	}

	slice1 := strings.Split(u.Path, "/")
	filename := slice1[len(slice1)-1]

	result := fmt.Sprintf("wget -O %s %s", filename, u.Scheme+"://"+u.Host+u.Path)
	return result, nil
}

func (i AudioData) Basename() (string, error) {

	u, err := url.Parse(i.AudioUrl)
	if err != nil {
		return "", err
	}

	slice1 := strings.Split(u.Path, "/")
	filename := slice1[len(slice1)-1]

	return filename, nil

}

type Enclosure struct {
	URL    string
	Length string
	Type   string
}

func (i Enclosure) String() string {
	return fmt.Sprintf(" URL: %s\nLength: %s\nType: %s", i.URL, i.Length, i.Type)
}

//------------------------------FUNCTIONS-------------------------------------------------

func cleanURL(i string) string {

	u, err := url.Parse(i)
	if err != nil {
		return ""
	}

	return u.Scheme + "://" + u.Host + u.Path

}

func basenameURL(i string) string {

	u, err := url.Parse(i)
	if err != nil {
		return ""
	}

	slice1 := strings.Split(u.Path, "/")
	filename := slice1[len(slice1)-1]

	return filename

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

}

func parseJSON(b string) {

	program := new(Program)
	reader := strings.NewReader(b)

	err := json.NewDecoder(reader).Decode(program)
	if err != nil {
		fmt.Printf("%v\n\n\n", err)
	}

	for _, episode := range program.AudioData {
		cmd, _ := episode.FormatAudioUrl()
		fmt.Printf("%s\n", cmd)
	}

	basename, _ := program.AudioData[0].Basename()
	//fmt.Printf("\n%v\n", basename)

	for _, episode := range program.AudioData[1:] {
		basename1, _ := episode.Basename()
		fmt.Printf("cat %s >> %s\nrm %s\n", basename1, basename, basename1)
	}

}

func CollectLinksRadiolab(httpBody io.Reader) []Enclosure {
	// http://golang-examples.tumblr.com/post/47426518779/parse-html

	r, _ := regexp.Compile("\"http(s://|://).*mp3\"")

	links := make([]Enclosure, 0)
	scriptToken := 0
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
				scriptToken = 1
				continue
			}
		case html.TextToken: // text between start and end tag
			if scriptToken == 1 {
				match, _ := regexp.MatchString("embed_audio_buttons", token.Data)
				if match == true {
					//fmt.Printf("%s\n", strings.Trim(token.Data, "\t\n "))
					link := r.FindString(strings.Trim(token.Data, "\t\n "))
					link = strings.Replace(link, "\"", "", -1)
					//fmt.Printf("-> %s\n",link)
					links = append(links, Enclosure{link, "0", "mp3"})
				}
			}
		case html.EndTagToken: // </tag>
			scriptToken = 0
		case html.SelfClosingTagToken: // <tag/>
			scriptToken = 0
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

func initApp() *cli.App {

	app := cli.NewApp()
	app.Name = "crawl"
	app.Version = "0.0.0"
	app.Usage = "A command line client written in golang to scrape radiolab and all things considered website for mp3 files."

	app.Commands = []cli.Command{
		radiolabCommand(),
		atcCommand(),
		iTunesCommand(),
		// Add more sub-commands ...
	}

	return app

}

func radiolabCommand() cli.Command {
	command := cli.Command{
		Name:   "radiolab",
		Usage:  "Grabs radiolab episodes from its website, http://radiolab.org",
		Action: actionRadiolab,
	}
	return command

}

func actionRadiolab(ctx *cli.Context) {

	resp, err := http.Get("http://radiolab.org")
	if err != nil {
		return
	}

	defer resp.Body.Close()

	links := CollectLinksRadiolab(resp.Body)
	for _, link := range links { // 'for' + 'range' in Go is like .each in Ruby or
		fmt.Printf("%s\n", link.URL)
	}

}

func atcCommand() cli.Command {
	command := cli.Command{
		Name:      "all-things-considered",
		ShortName: "atc",
		Usage:     "Grabs All things considered episodes from its website, http://www.npr.org/programs/all-things-considered/",
		Action:    actionAtc,
	}
	return command
}

func actionAtc(ctx *cli.Context) {
	fmt.Printf("%s\n", "http://www.npr.org/programs/all-things-considered/")

	resp, err := http.Get("http://www.npr.org/programs/all-things-considered/")
	if err != nil {
		return
	}

	defer resp.Body.Close()

	links := CollectLinksATC(resp.Body)
	for _, link := range links { // 'for' + 'range' in Go is like .each in Ruby or
		fmt.Printf("%v\n", link)
	}

}

func iTunesCommand() cli.Command {
	command := cli.Command{
		Name:   "itunes",
		Usage:  "Query iTunes for top podcasts",
		Action: actionItunes,
	}
	return command
}

func actionItunes(ctx *cli.Context) {

	size := 10
	if ctx.NArg() > 0 {
		tmp, _ := strconv.Atoi(ctx.Args()[0])
		size = tmp
	}

	URL := fmt.Sprintf("https://itunes.apple.com/us/rss/toppodcasts/limit=%d/explicit=true/xml", size)
	resp, err := http.Get(URL)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var feed Feed
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		return
	}

	fmt.Printf("# %s\n\n", feed.Title)

	for _, entry := range feed.Entries {
		fmt.Printf("- %s\n%s\n%s\n\n", entry.Title, entry.Summary, entry.Link.Href)
	}

}

func main() {

	app := initApp()
	app.Run(os.Args)

}
