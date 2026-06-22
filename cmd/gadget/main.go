package main

import (
	"fmt"
	"os"
	"strings"

	"zgo.at/gadget"
)

func main() {
	for _, a := range os.Args[1:] {
		var full, short string
		if strings.Contains(a, "~") {
			full, short = gadget.UnshortenUA(a), a
		} else {
			full, short = a, gadget.ShortenUA(a)
		}
		ua := gadget.ParseUA(full)
		fmt.Printf("%s\t%s\t%s\t%s\n", ua.Browser(), ua.OS(), short, full)
	}
}
