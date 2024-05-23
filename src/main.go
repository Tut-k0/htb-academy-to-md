package main

import (
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"os"
	"strings"
)

func main() {
	options := getArguments()
	fmt.Println("Authenticating with HackTheBox...")
	session := authenticate(options.email, options.password)
	fmt.Println("Downloading requested module...")
	title, content := getModule(options.moduleUrl, session)
	if options.localImages {
		fmt.Println("Downloading module images...")
		content = getImagesLocally(content)
	}

	markdownContent := htmlToMarkdown(content)

	err := os.WriteFile(title+".md", []byte(markdownContent), 0666)
	if err != nil {
		die(err)
	}
	fmt.Println("Finished downloading module!")
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
