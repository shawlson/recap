package main

import (
	"flag"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"os"
)

func main() {
	var docType string
	flag.StringVar(&docType, "type", "html", "document type being minified - html/css")
	flag.Parse()

	var mediatype string
	switch docType {
	case "html":
		mediatype = "text/html"
	case "css":
		mediatype = "text/css"
	}

	m := minify.New()
	m.Add("text/html", &html.Minifier{
		KeepDocumentTags: true,
		KeepEndTags:      true,
		KeepWhitespace:   false,
	})
	m.AddFunc("text/css", css.Minify)

	if err := m.Minify(mediatype, os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
