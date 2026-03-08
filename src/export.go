package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
)

type userAgentTransport struct {
	Transport http.RoundTripper
	UserAgent string
}

// API Response Structures for HTB Academy 2.0
type ModuleResponse struct {
	Data ModuleData `json:"data"`
}

type ModuleData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SectionsResponse struct {
	Data []SectionGroup `json:"data"`
}

type SectionGroup struct {
	Group    string    `json:"group"`
	Sections []Section `json:"sections"`
}

type Section struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Page  int    `json:"page"`
}

type SectionContentResponse struct {
	Data SectionContent `json:"data"`
}

type SectionContent struct {
	Content string `json:"content"`
}

func authenticateWithCookies(cookies string) *http.Client {
	client, err := newClient(cookies)
	if err != nil {
		die(err)
	}

	// Validates authentication by checking access to the dashboard
	resp, err := client.Get("https://academy.hackthebox.com/app/dashboard")
	if err != nil {
		die(err)
	}

	if resp.StatusCode != 200 {
		fmt.Println("Authentication Failed, refresh your cookies and try again!")
		os.Exit(1)
	}

	return client
}

func (ua *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", ua.UserAgent)
	}
	return ua.Transport.RoundTrip(req)
}

func newClient(cookies string) (*http.Client, error) {
	// For proxy debugging
	//proxy, _ := url.Parse("http://localhost:8080")
	//transport := &userAgentTransport{
	//	Transport: &http.Transport{
	//		Proxy:           http.ProxyURL(proxy),
	//		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//	},
	//	UserAgent: "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
	//}
	transport := &userAgentTransport{
		Transport: http.DefaultTransport,
		UserAgent: "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar:       jar,
		Transport: transport,
	}

	if cookies != "" {
		addCookiesToJar(jar, cookies)
	}

	return client, nil
}

func addCookiesToJar(jar *cookiejar.Jar, cookies string) {
	cookiePairs := strings.Split(cookies, ";")
	cookieList := []*http.Cookie{}

	for _, pair := range cookiePairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			cookieList = append(cookieList, &http.Cookie{
				Name:  strings.TrimSpace(parts[0]),
				Value: strings.TrimSpace(parts[1]),
			})
		}
	}

	u, _ := url.Parse("https://academy.hackthebox.com")
	jar.SetCookies(u, cookieList)
}

func getModule(moduleUrl string, client *http.Client) (string, []string) {
	// Extract module ID from URL (e.g., https://academy.hackthebox.com/module/163/section/1546)
	moduleID := extractModuleID(moduleUrl)
	
	// Normalize URL to use /app/ path format for referer
	refererUrl := normalizeModuleUrl(moduleUrl)
	
	// Fetch module metadata to get the title
	moduleTitle := getModuleMetadata(moduleID, refererUrl, client)
	
	// Fetch all sections for this module
	sections := getModuleSections(moduleID, refererUrl, client)
	
	// Fetch content for each section
	var pagesContent []string
	for _, section := range sections {
		content := getSectionContent(moduleID, section.ID, refererUrl, client)
		pagesContent = append(pagesContent, content)
	}
	
	return moduleTitle, pagesContent
}

func extractModuleID(moduleUrl string) string {
	// Parse URL like: https://academy.hackthebox.com/module/163/section/1546
	// or: https://academy.hackthebox.com/app/module/163/section/1546
	parts := strings.Split(moduleUrl, "/")
	for i, part := range parts {
		if part == "module" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	fmt.Println("Could not extract module ID from URL:", moduleUrl)
	os.Exit(1)
	return ""
}

func normalizeModuleUrl(moduleUrl string) string {
	// Ensure URL uses /app/module/ format
	// Convert: https://academy.hackthebox.com/module/163/... 
	// To: https://academy.hackthebox.com/app/module/163/...
	if strings.Contains(moduleUrl, "/app/module/") {
		return moduleUrl
	}
	return strings.Replace(moduleUrl, "/module/", "/app/module/", 1)
}

func getModuleMetadata(moduleID string, refererUrl string, client *http.Client) string {
	apiUrl := fmt.Sprintf("https://academy.hackthebox.com/api/v2/modules/%s", moduleID)
	
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		die(err)
	}
	
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", refererUrl)
	
	resp, err := client.Do(req)
	if err != nil {
		die(err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		die(err)
	}
	
	if resp.StatusCode != 200 {
		fmt.Printf("Failed to fetch module metadata. Status: %d\n", resp.StatusCode)
		fmt.Println("Response:", string(body))
		os.Exit(1)
	}
	
	var moduleResp ModuleResponse
	if err := json.Unmarshal(body, &moduleResp); err != nil {
		die(err)
	}
	
	// Clean the title for use as a filename
	title := moduleResp.Data.Name
	badChars := []string{"/", "\\", "?", "%", "*", ":", "|", "\"", "<", ">"}
	for _, badChar := range badChars {
		title = strings.ReplaceAll(title, badChar, "-")
	}
	
	return title
}

func getModuleSections(moduleID string, refererUrl string, client *http.Client) []Section {
	apiUrl := fmt.Sprintf("https://academy.hackthebox.com/api/v3/modules/%s/sections", moduleID)
	
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		die(err)
	}
	
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", refererUrl)
	
	resp, err := client.Do(req)
	if err != nil {
		die(err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		die(err)
	}
	
	if resp.StatusCode != 200 {
		fmt.Printf("Failed to fetch module sections. Status: %d\n", resp.StatusCode)
		fmt.Println("Response:", string(body))
		os.Exit(1)
	}
	
	var sectionsResp SectionsResponse
	if err := json.Unmarshal(body, &sectionsResp); err != nil {
		die(err)
	}
	
	// Flatten all sections from all groups and sort by page number
	var allSections []Section
	for _, group := range sectionsResp.Data {
		allSections = append(allSections, group.Sections...)
	}
	
	// Sort sections by page number
	for i := 0; i < len(allSections); i++ {
		for j := i + 1; j < len(allSections); j++ {
			if allSections[i].Page > allSections[j].Page {
				allSections[i], allSections[j] = allSections[j], allSections[i]
			}
		}
	}
	
	return allSections
}

func getSectionContent(moduleID string, sectionID int, refererUrl string, client *http.Client) string {
	apiUrl := fmt.Sprintf("https://academy.hackthebox.com/api/v2/modules/%s/sections/%d", moduleID, sectionID)
	
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		die(err)
	}
	
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", refererUrl)
	
	resp, err := client.Do(req)
	if err != nil {
		die(err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		die(err)
	}
	
	if resp.StatusCode != 200 {
		fmt.Printf("Failed to fetch section %d. Status: %d\n", sectionID, resp.StatusCode)
		fmt.Println("Response:", string(body))
		os.Exit(1)
	}
	
	var contentResp SectionContentResponse
	if err := json.Unmarshal(body, &contentResp); err != nil {
		die(err)
	}
	
	// The content is already in markdown format with \r\n line endings
	// Normalize line endings to \n
	content := strings.ReplaceAll(contentResp.Data.Content, "\r\n", "\n")
	
	return content
}

func fixImageUrls(sections []string) []string {
	var result []string
	for _, section := range sections {
		// Replace relative image paths with absolute URLs
		updatedSection := fixRelativeImageUrls(section)
		result = append(result, updatedSection)
	}
	return result
}

func fixRelativeImageUrls(content string) string {
	result := content
	
	// Find all markdown images: ![alt](path) and make URLs absolute
	searchPos := 0
	for {
		start := strings.Index(result[searchPos:], "![")
		if start == -1 {
			break
		}
		start += searchPos
		
		// Find the closing ]
		altEnd := strings.Index(result[start:], "](")
		if altEnd == -1 {
			break
		}
		altEnd += start
		
		// Find the closing )
		pathStart := altEnd + 2
		pathEnd := strings.Index(result[pathStart:], ")")
		if pathEnd == -1 {
			break
		}
		pathEnd += pathStart
		
		imagePath := result[pathStart:pathEnd]
		
		// If path is relative, make it absolute
		if !strings.HasPrefix(imagePath, "http") {
			newPath := "https://academy.hackthebox.com" + imagePath
			result = result[:pathStart] + newPath + result[pathEnd:]
			searchPos = pathStart + len(newPath)
		} else {
			searchPos = pathEnd + 1
		}
	}
	
	return result
}

func getImagesLocally(sections []string, moduleTitle string, moduleID string) []string {
	// Create images directory if it doesn't exist
	imgDir := "images"
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		die(err)
	}
	
	var result []string
	imageCounter := 0
	
	for _, section := range sections {
		updatedSection, newCounter := replaceImagePathsInSectionWithCounter(section, moduleID, imageCounter)
		imageCounter = newCounter
		result = append(result, updatedSection)
	}
	
	return result
}

func replaceImagePathsInSectionWithCounter(content string, moduleID string, startCounter int) (string, int) {
	imageCounter := startCounter
	result := content
	
	// Find all markdown images: ![alt](path)
	// Process them from the end backwards to avoid offset issues
	var imageMatches []struct {
		start    int
		altEnd   int
		pathStart int
		pathEnd  int
		imagePath string
	}
	
	searchPos := 0
	for {
		start := strings.Index(result[searchPos:], "![")
		if start == -1 {
			break
		}
		start += searchPos
		
		// Find the closing ]
		altEnd := strings.Index(result[start:], "](")
		if altEnd == -1 {
			break
		}
		altEnd += start
		
		// Find the closing )
		pathStart := altEnd + 2
		pathEnd := strings.Index(result[pathStart:], ")")
		if pathEnd == -1 {
			break
		}
		pathEnd += pathStart
		
		imagePath := result[pathStart:pathEnd]
		
		imageMatches = append(imageMatches, struct {
			start    int
			altEnd   int
			pathStart int
			pathEnd  int
			imagePath string
		}{start, altEnd, pathStart, pathEnd, imagePath})
		
		searchPos = pathEnd + 1
	}
	
	// Process images from the end backwards
	for i := len(imageMatches) - 1; i >= 0; i-- {
		match := imageMatches[i]
		imagePath := match.imagePath
		
		var newPath string
		imageCounter++
		
		fullUrl := imagePath
		if !strings.HasPrefix(imagePath, "http") {
			fullUrl = "https://academy.hackthebox.com" + imagePath
		}
		
		// Extract original filename from path
		pathParts := strings.Split(imagePath, "/")
		originalName := pathParts[len(pathParts)-1]
		
		// Create a meaningful filename and download
		newPath = downloadImageToFile(fullUrl, moduleID, imageCounter, originalName)
		
		// Replace the image path in the result
		result = result[:match.pathStart] + newPath + result[match.pathEnd:]
	}
	
	return result, imageCounter
}

func downloadImageToFile(fileUrl string, moduleID string, counter int, originalName string) string {
	// Create filename: images/module-{id}-{counter}-{original}.ext
	ext := ""
	if idx := strings.LastIndex(originalName, "."); idx != -1 {
		ext = originalName[idx:]
	}
	
	fileName := fmt.Sprintf("images/module-%s-%03d%s", moduleID, counter, ext)
	
	resp, err := http.Get(fileUrl)
	if err != nil {
		fmt.Printf("Warning: Failed to download image %s: %v\n", fileUrl, err)
		return fileUrl // Return original URL on failure
	}
	defer resp.Body.Close()
	
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Warning: Failed to read image %s: %v\n", fileUrl, err)
		return fileUrl
	}
	
	// If extension can't be determined from URL, detect from content
	if ext == "" {
		if isPNG(content) {
			fileName = fileName + ".png"
		} else if isJPEG(content) {
			fileName = fileName + ".jpg"
		} else if isGIF(content) {
			fileName = fileName + ".gif"
		}
	}
	
	err = os.WriteFile(fileName, content, 0666)
	if err != nil {
		fmt.Printf("Warning: Failed to write image %s: %v\n", fileName, err)
		return fileUrl
	}
	
	return fileName
}

func isPNG(b []byte) bool {
	return len(b) >= 8 && string(b[0:8]) == "\x89PNG\r\n\x1a\n"
}

func isJPEG(b []byte) bool {
	return len(b) >= 2 && string(b[0:2]) == "\xff\xd8"
}

func isGIF(data []byte) bool {
	if len(data) < 6 {
		return false
	}
	if string(data[:3]) != "GIF" {
		return false
	}
	if string(data[3:6]) != "87a" && string(data[3:6]) != "89a" {
		return false
	}
	return true
}

func die(err error) {
	fmt.Println(err)
	os.Exit(1)
}
