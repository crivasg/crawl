// Usage ./crawl podcast --feed http://feeds.wnyc.org/onthemedia
// Usage ./crawl podcast -f http://feeds.wnyc.org/onthemedia

package main

import (
	"encoding/xml"
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"net/http"
)

type FeedRSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr,omitempty"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	PubDate       string `xml:"pubDate"`
	Items         []Item `xml:"item"`
	LastBuildDate string `xml:"lastBuildDate"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Guid        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Author      string `xml:"author"`
	Description string `xml:"description"`
	Encl        []Encl `xml:"enclosure"`
}

type Encl struct {
	Url    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

func podcastCommand() cli.Command {

	command := cli.Command{
		Name:    "podcast",
		Aliases: []string{"p"},
		Usage:   "Grabs podcast episodes",
		Action:  actionPodcast,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "feed, f",
				Usage: "The `URL` to be parsed",
			},
		},
	}
	return command

}

func actionPodcast(ctx *cli.Context) {

	feed := ctx.String("feed")

	if len(feed) > 0 {
		channel, _ := getPodcastData(feed)

		for _, item := range channel.Items {
			fmt.Printf("# %s\n%s\n", item.Title, item.Description)
			for _, encl := range item.Encl {
				fmt.Printf("%s\n", encl.Url)
			}
		}

	}

}

func getPodcastData(feed_url string) (Channel, error) {

	res, err := http.Get(feed_url)
	if err != nil {
		return Channel{}, err
	}

	if res.StatusCode != http.StatusOK {
		return Channel{}, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Channel{}, err
	}

	var feed FeedRSS
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		return Channel{}, err
	}

	return feed.Channel, nil
}
