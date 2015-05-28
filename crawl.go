/*
   https://jdanger.com/build-a-web-crawler-in-go.html

   Build a Web Crawler in Go

*/

package main

import (
	"flag"
	"fmt"
	"io/ioutil" // 'ioutil' will help us print pages to the screen
	"net/http"
	"os"
)

func Usage() {
	fmt.Fprintf(os.Stderr, "usage: crawl http://example.com/path/file.html\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func RetrieveDataFrom(uri string) (string, error) {

	resp, err := http.Get(uri)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return "", err2
	}

	return string(body), nil

}

func main() {

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		Usage()
	}

	data, err := RetrieveDataFrom(args[0])
	if err != nil {
	    fmt.Fprintf(os.Stderr, "Error retrieving the data from %s\n", args[0])
	    os.Exit(1)
	}
	fmt.Println(data)

}
