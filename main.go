package main

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"os"
)

func main() {
	options := getArguments()
	session := authenticate(options.email, options.password)
	title, content := getModule(options.moduleUrl, session)
	markdownContent := htmlToMarkdown(content)

	err := os.WriteFile(title+".md", []byte(markdownContent), 0666)
	if err != nil {
		die(err)
	}
}

func htmlToMarkdown(html []string) string {
	converter := md.NewConverter("", true, nil)
	converter.Use(plugin.GitHubFlavored())
	var markdown string
	for _, content := range html {
		m, err := converter.ConvertString(content)
		if err != nil {
			die(err)
		}
		markdown += m + "\n\n\n"
	}

	return markdown
}
