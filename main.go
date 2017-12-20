package main

import (
	"flag"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/net/html"
)

var file = flag.String("file", "", "file to scan")

type stackItem struct {
	name string
	raw  string
	line int
}

func main() {
	flag.Parse()

	f, err := os.Open(*file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	stack := []*stackItem{}

	r := html.NewTokenizer(f)
	currentLine := 1
	for {
		r.Next()

		token := r.Token()
		switch token.Type {
		case html.ErrorToken:
			if r.Err() == io.EOF {
				if len(stack) > 0 {
					expected := stack[len(stack) - 1]
					log.Println("Found unclosed tag!")
					log.Println("===================")
					log.Printf("Expected: </%s>\n", expected.name)
					log.Printf("Got: EOF\n")
					log.Printf("Opened in: %s", expected.raw)
					log.Printf("Line: %d", expected.line)
					os.Exit(1)
				}
				return
			}

			log.Fatalf("token failed %s: %s", r.Err(), r.Raw())

		case html.StartTagToken:
			if isSelfClosed(token.Data) {
				continue
			}

			stack = append(stack, &stackItem{
				name: token.Data,
				raw:  string(r.Raw()),
				line: currentLine,
			})

		case html.EndTagToken:
			expected := stack[len(stack)-1]
			if expected.name != token.Data {
				log.Println("Found unclosed tag!")
				log.Println("===================")
				log.Printf("Expected: </%s>\n", expected.name)
				log.Printf("Got: </%s>\n", token.Data)
				log.Printf("Opened in: %s", expected.raw)
				log.Printf("Line: %d", expected.line)
				os.Exit(1)
				return
			}
			stack = stack[:len(stack)-1]
		}
		currentLine += strings.Count(string(r.Raw()), "\n")
	}
}

func isSelfClosed(name string) bool {
	switch name {
	case "link":
		return true
	case "meta":
		return true
	case "img":
		return true
	case "hr":
		return true
	case "br":
		return true
	case "input":
		return true
	}

	return false
}
