/*
   https://jdanger.com/build-a-web-crawler-in-go.html

   Build a Web Crawler in Go

*/

package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Please specify start page")
		os.Exit(1)
	}

}
