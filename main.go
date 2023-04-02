package main

import (
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"os"
)

func main() {
	// options := getArguments()
	// session := authenticate(options.email, options.password)
	// title, content := getModule(options.moduleUrl, session)
	textBytes, err := os.ReadFile("test-1-page.html")
	if err != nil {
		die(err)
	}
	html := string(textBytes)
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(html)
	if err != nil {
		die(err)
	}

	fmt.Println(markdown)
}
