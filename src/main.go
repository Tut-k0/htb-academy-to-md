package main

import (
	"fmt"
	"os"
	"strings"
    "strconv"
)

func main() {
	options := getArguments()
	fmt.Println("Authenticating with HackTheBox...")
	session := authenticateWithCookies(options.cookies)
	fmt.Println("Downloading requested module...")
	title, content, groups := getModule(options.moduleUrl, session)

	// Extract module ID for image folder naming
	moduleID := extractModuleID(options.moduleUrl)

	if options.localImages {
		fmt.Println("Downloading module images...")
		content = getImagesLocally(content, moduleID, title, options.dirTree)
	} else {
		// Fix image URLs to be absolute
		content = fixImageUrls(content)
	}
    
    
	markdownContent := cleanMarkdown(content, options.dirTree)
    writeOutput(title, markdownContent, groups, options.dirTree)
	fmt.Println("Finished downloading module!")
}

func cleanMarkdown(sections []string, directoryTree bool) []string {
    // Helpfunction innerCleanFunc
    innerCleanFunc := func(s string) string {
        s = strings.ReplaceAll(s, "shell-session", "shell")
        s = strings.ReplaceAll(s, "powershell-session", "powershell")
        s = strings.ReplaceAll(s, "cmd-session", "shell")
        s = strings.ReplaceAll(s, " [!bash!]$ ", " ")
        s = strings.ReplaceAll(s, "[!bash!]$ ", "")
        s = fixCodeBlockFences(s)
        return s
    }

    // Helpfunction sectionsInSingleString
    sectionsInSingleString := func(sections []string) string {
	    var str string
	    for _, content := range sections {
		    str += content + "\n\n\n"
	    }
        return str
    }
    
   cleanedStrings := []string{}
   if directoryTree {
        for _, section := range sections {
            section = innerCleanFunc(section)
            cleanedStrings = append(cleanedStrings, section)
        }
    } else {
        markdown := sectionsInSingleString(sections)
        markdown = innerCleanFunc(markdown)
        cleanedStrings = append(cleanedStrings, markdown) //array with a len of 1
    }
    return cleanedStrings

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

func writeOutput(title string, markdownContent []string, groups []SectionGroup, directoryTree bool) {

    innerWriter := func(path string, str string ) {
        err := os.WriteFile(path, []byte(str), 0666)
        if err != nil {
            die(err)
        }
    }

    writeTreeDirectory := func(title string, markdownContent []string, groups []SectionGroup) {
        if err := os.MkdirAll(title, 0755); err != nil {
	        fmt.Println("could not create folder")
		    die(err)
	    }

        counter := 0
        for groupCounter, group := range groups {
            group.Group = cleanFilename(group.Group)
        
            groupDirPath := title+"/"+strconv.Itoa(groupCounter+1)+" "+group.Group
            if err := os.MkdirAll(groupDirPath, 0755); err != nil {
	            fmt.Println("could not create folder")
		        die(err)
	        }
            for sectionCounter, section := range group.Sections {
                section.Title = cleanFilename(section.Title)                

                sectionDirPath := groupDirPath+"/"+strconv.Itoa(sectionCounter+1)+" "+section.Title+".md"
                innerWriter(sectionDirPath, markdownContent[counter])
                counter++
            }
        }
    }

    if directoryTree {
        writeTreeDirectory(title, markdownContent, groups)
    } else {
        innerWriter(title+".md",markdownContent[0])
    }


}


