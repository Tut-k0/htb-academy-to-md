package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
	"io"
	"math/rand"
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

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	IntendedRoute string `json:"intended_route"`
}

func authenticateWithCookies(cookies string) *http.Client {
	client, err := newClient(cookies)
	if err != nil {
		die(err)
	}

	resp, err := client.Get("https://academy.hackthebox.com/dashboard")
	if err != nil {
		die(err)
	}

	if resp.StatusCode != 200 {
		fmt.Println("Authentication Failed, refresh your cookies and try again!")
		os.Exit(1)
	}

	return client
}

func authenticate(email, password string) *http.Client {
	client, err := newClient("")
	if err != nil {
		die(err)
	}

	// Get academy cookies cookies
	resp, err := client.Get("https://academy.hackthebox.com/login")
	if err != nil {
		die(err)
	}

	// Head over to the SSO login
	resp, err = client.Get("https://academy.hackthebox.com/sso/redirect")
	if err != nil {
		die(err)
	}

	// Get CSRF-cookie for logging in
	resp, err = client.Get("https://account.hackthebox.com/api/v1/csrf-cookie")
	if err != nil {
		die(err)
	}

	// Login to HackTheBox
	credentials := Credentials{
		Email:    email,
		Password: password,
	}
	jsonData, err := json.Marshal(credentials)
	if err != nil {
		die(err)
	}
	req, err := http.NewRequest("POST", "https://account.hackthebox.com/api/v1/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		die(err)
	}
	xsrfToken := getXSRFToken(client, "https://account.hackthebox.com/")
	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("X-Xsrf-Token", xsrfToken)
	req.Header.Set("Origin", "https://account.hackthebox.com")
	req.Header.Set("Referer", "https://account.hackthebox.com/login")
	// Perform the request
	resp, err = client.Do(req)
	if err != nil {
		die(err)
	}

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		fmt.Println("Authenticating to HackTheBox failed.")
		fmt.Printf("Unexpected status code: %d\n", resp.StatusCode)
		fmt.Println("Response body:", string(body))
		os.Exit(1)
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		die(err)
	}

	// Make a request to the intended route
	resp, err = client.Get(loginResp.IntendedRoute)
	if err != nil {
		die(err)
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

func getXSRFToken(client *http.Client, urlStr string) string {
	u, _ := url.Parse(urlStr)
	cookies := client.Jar.Cookies(u)
	for _, cookie := range cookies {
		if cookie.Name == "XSRF-TOKEN" {
			rawToken, err := url.QueryUnescape(cookie.Value)
			if err != nil {
				die(err)
			}
			return rawToken
		}
	}
	return ""
}

func getModule(moduleUrl string, client *http.Client) (string, []string) {
	resp, err := client.Get(moduleUrl)
	if err != nil {
		die(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		die(err)
	}
	content := string(body)
	moduleTitle := getModuleTitle(content)
	pageUrls := getModulePages(content, moduleUrl)

	var pagesContent []string
	for _, pageUrl := range pageUrls {
		pagesContent = append(pagesContent, extractPageContent(pageUrl, client))
	}

	return moduleTitle, pagesContent
}

func getModuleTitle(htmlText string) string {
	var title string
	var isTitle bool
	badChars := []string{"/", "\\", "?", "%", "*", ":", "|", "\"", "<", ">"}
	tkn := html.NewTokenizer(strings.NewReader(htmlText))

	for {
		tt := tkn.Next()
		switch {

		case tt == html.ErrorToken:
			os.Exit(1)

		case tt == html.StartTagToken:
			t := tkn.Token()
			if t.Data == "title" {
				isTitle = true
			}
		case tt == html.TextToken:
			t := tkn.Token()

			if isTitle {
				title = t.Data
				for _, badChar := range badChars {
					title = strings.ReplaceAll(title, badChar, "-")
				}
				return title
			}
		case tt == html.EndTagToken:
			t := tkn.Token()
			if t.Data == "html" {
				fmt.Println("Could not find title on module page, exiting.")
				os.Exit(1)
			}
		}
	}
}

func getModulePages(htmlText string, moduleUrl string) []string {
	var modulePages []string

	doc, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		die(err)
	}

	var traverse func(n *html.Node) *html.Node
	traverse = func(n *html.Node) *html.Node {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Data == "a" {
				// Pass on <a> tags that do not have any attributes.
				if len(c.Attr) == 0 {
				} else if strings.Contains(c.Attr[0].Val, moduleUrl[:44]) {
					modulePages = append(modulePages, c.Attr[0].Val)
				}
			}
			res := traverse(c)
			if res != nil {
				return res
			}
		}
		return nil
	}
	traverse(doc)

	return modulePages[1:]
}

func getModulePageContent(pageUrl string, client *http.Client) string {
	resp, err := client.Get(pageUrl)
	if err != nil {
		die(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		die(err)
	}
	content := string(body)

	return content
}

func extractPageContent(pageUrl string, client *http.Client) string {
	var result string
	pageContent := getModulePageContent(pageUrl, client)

	doc, err := html.Parse(strings.NewReader(pageContent))
	if err != nil {
		die(err)
	}

	trainingContent := findDivByClass(doc, "training-module")
	if trainingContent != nil {
		var tc strings.Builder
		if err := html.Render(&tc, trainingContent); err != nil {
			die(err)
		}
		result = tc.String()
	} else {
		fmt.Printf("Parsing training content failed, HTML dump: %s", pageContent)
		os.Exit(1)
	}

	currentDoc, err := html.Parse(strings.NewReader(result))
	if err != nil {
		die(err)
	}

	contentToRemove := findDivByClassOrId(currentDoc, "vpn-switch-card", "screen")
	if contentToRemove != nil {
		var htmlBuilder strings.Builder
		if err := html.Render(&htmlBuilder, contentToRemove); err != nil {
			die(err)
		}
		indexToRemove := strings.Index(result, htmlBuilder.String())
		result = result[:indexToRemove]
	}

	curDoc, err := html.Parse(strings.NewReader(result))
	if err != nil {
		die(err)
	}
	fixImgs(curDoc)
	var r strings.Builder
	if err := html.Render(&r, curDoc); err != nil {
		die(err)
	}
	result = r.String()

	return result
}

func findDivByClass(n *html.Node, className string) *html.Node {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, a := range n.Attr {
			if a.Key == "class" && strings.Contains(a.Val, className) {
				return n
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if div := findDivByClass(c, className); div != nil {
			return div
		}
	}
	return nil
}

func findDivByClassOrId(n *html.Node, className string, id string) *html.Node {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, a := range n.Attr {
			if (a.Key == "class" && strings.Contains(a.Val, className)) ||
				(a.Key == "id" && strings.Contains(a.Val, id)) {
				return n
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if div := findDivByClassOrId(c, className, id); div != nil {
			return div
		}
	}
	return nil
}

func fixImgs(node *html.Node) {
	if node.Type == html.ElementNode && node.Data == "img" {
		for i, attr := range node.Attr {
			if attr.Key == "src" && !startsWith(attr.Val, "https://") {
				node.Attr[i].Val = "https://academy.hackthebox.com" + attr.Val
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		fixImgs(child)
	}
}

func startsWith(str, prefix string) bool {
	return len(str) >= len(prefix) && str[0:len(prefix)] == prefix
}

func getImagesLocally(htmlPages []string) []string {
	var result []string
	for _, page := range htmlPages {
		doc, err := html.Parse(strings.NewReader(page))
		if err != nil {
			die(err)
		}
		replaceImgs(doc)
		var htmlBulder strings.Builder
		if err := html.Render(&htmlBulder, doc); err != nil {
			die(err)
		}
		result = append(result, htmlBulder.String())
	}
	return result
}

func replaceImgs(node *html.Node) {
	if node.Type == html.ElementNode && node.Data == "img" {
		for i, attr := range node.Attr {
			if attr.Key == "src" {
				fileName := downloadImage(node.Attr[i].Val)
				node.Attr[i].Val = fileName
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		replaceImgs(child)
	}
}

func downloadImage(fileUrl string) string {
	fileName := randomFileName()
	resp, err := http.Get(fileUrl)
	if err != nil {
		die(err)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		die(err)
	}
	if isPNG(content) {
		fileName = fileName + ".png"
		err := os.WriteFile(fileName, content, 0666)
		if err != nil {
			die(err)
		}
	} else if isJPEG(content) {
		fileName = fileName + ".jpg"
		err := os.WriteFile(fileName, content, 0666)
		if err != nil {
			die(err)
		}
	} else if isGIF(content) {
		fileName = fileName + ".gif"
		err := os.WriteFile(fileName, content, 0666)
		if err != nil {
			die(err)
		}
	} else {
		err := os.WriteFile(fileName, content, 0666)
		if err != nil {
			die(err)
		}
	}

	return fileName
}

func randomFileName() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, 12)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
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
