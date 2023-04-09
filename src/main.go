package main

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"os"
	"strings"
)

func main() {
	options := getArguments()
	session := authenticate(options.email, options.password)
	title, content := getModule(options.moduleUrl, session)
	if options.localImages {
		content = getImagesLocally(content)
	}

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

	// Strip some content for proper code blocks.
	markdown = strings.ReplaceAll(markdown, "shell-session", "shell")
	markdown = strings.ReplaceAll(markdown, "powershell-session", "powershell")
	markdown = strings.ReplaceAll(markdown, "[!bash!]$ ", "")

	return markdown
}
