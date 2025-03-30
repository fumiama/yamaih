package main

import (
	"flag"
	"fmt"

	"github.com/fumiama/yamaih"
)

func main() {
	ep := flag.String("l", "127.0.0.1:6783", "listening endpoint")
	df := flag.String("f", "log.db", "log database file path")
	api := flag.String("api", "v1beta", "api version")
	flag.Parse()
	g := yamaih.NewGemini(*ep, *df, *api)
	fmt.Println("Fatal err:", g.RunBlocking())
}
