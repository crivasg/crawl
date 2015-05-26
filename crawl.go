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

func retrieve(uri string) (string, error) {

	resp, err := http.Get(uri)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	return string(body), nil

}

func main() {

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Please specify start page")
		os.Exit(1)
	}

	data, _ := retrieve(args[0])

	fmt.Println(data)

}
