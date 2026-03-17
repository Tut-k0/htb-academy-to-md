package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	options := getArguments()
	fmt.Println("Authenticating with HackTheBox...")
	session := authenticateWithCookies(options.cookies)
	fmt.Println("Downloading requested module...")
	title, content := getModule(options.moduleUrl, session)

	// Extract module ID for image folder naming
	moduleID := extractModuleID(options.moduleUrl)

	if options.localImages {
		fmt.Println("Downloading module images...")
		content = getImagesLocally(content, moduleID)
	} else {
		// Fix image URLs to be absolute
		content = fixImageUrls(content)
	}

	markdownContent := cleanMarkdown(content)

	outputPath := title + ".md"
	if options.outputDir != "" {
		if err := os.MkdirAll(options.outputDir, 0755); err != nil {
			die(err)
		}
		outputPath = filepath.Join(options.outputDir, title+".md")
	}

	err := os.WriteFile(outputPath, []byte(markdownContent), 0666)
	if err != nil {
		die(err)
	}
	fmt.Println("Finished downloading module!")
}

func cleanMarkdown(sections []string) string {
	var markdown string
	for _, content := range sections {
		markdown += content + "\n\n\n"
	}

	// Strip some content for proper code blocks.
	markdown = strings.ReplaceAll(markdown, "shell-session", "shell")
	markdown = strings.ReplaceAll(markdown, "powershell-session", "powershell")
	markdown = strings.ReplaceAll(markdown, "cmd-session", "shell")
	// Remove bash prompts - handle both with and without leading space
	markdown = strings.ReplaceAll(markdown, " [!bash!]$ ", " ")
	markdown = strings.ReplaceAll(markdown, "[!bash!]$ ", "")

	// Fix malformed code block closures (remove leading/trailing spaces from lines with only backticks)
	markdown = fixCodeBlockFences(markdown)

	return markdown
}

func fixCodeBlockFences(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// If the line is only backticks (with optional language identifier)
		if strings.HasPrefix(trimmed, "```") {
			lines[i] = trimmed
		}
	}
	return strings.Join(lines, "\n")
}
