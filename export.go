package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Auth struct {
	loginToken string
	cookies    string
}

func authenticate(email, password string) Auth {
	auth := getLoginTokenAndCookies()
	payload := "_token=" + auth.loginToken + "&email=" + url.QueryEscape(email) + "&password=" + url.QueryEscape(password)

	//proxy, _ := url.Parse("http://localhost:8080")
	//tr := &http.Transport{Proxy: http.ProxyURL(proxy), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	//client.Transport = tr
	req, err := http.NewRequest("POST", "https://academy.hackthebox.com/login", strings.NewReader(payload))
	if err != nil {
		die(err)
	}
	req.Header.Add("Cookie", auth.cookies)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:102.0) Gecko/20100101 Firefox/102.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		die(err)
	}
	cookies := resp.Cookies()

	return Auth{
		loginToken: auth.loginToken,
		cookies:    cookies[0].Name + "=" + cookies[0].Value + "; " + cookies[1].Name + "=" + cookies[1].Value,
	}
}

func getLoginTokenAndCookies() Auth {
	resp, err := http.Get("https://academy.hackthebox.com/login")
	if err != nil {
		die(err)
	}
	cookies := resp.Cookies()
	if len(cookies) != 2 {
		fmt.Printf("WARNING: An unexpected amount of cookies has been sent, expected 2, received %d cookies.\n", len(cookies))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		die(err)
	}
	content := string(body)
	token := parseLoginToken(content)
	return Auth{
		loginToken: token,
		cookies:    cookies[0].Name + "=" + cookies[0].Value + "; " + cookies[1].Name + "=" + cookies[1].Value,
	}
}

func parseLoginToken(htmlText string) string {
	var token string
	tkn := html.NewTokenizer(strings.NewReader(htmlText))

	for {
		tt := tkn.Next()
		switch {

		case tt == html.ErrorToken:
			os.Exit(1)

		case tt == html.StartTagToken:
			t := tkn.Token()
			if t.Data == "input" && t.Attr[0].Val == "hidden" {
				token = t.Attr[2].Val
				return token
			}
		case tt == html.EndTagToken:
			t := tkn.Token()
			if t.Data == "html" {
				fmt.Println("Could not find token on login page, exiting.")
				os.Exit(1)
			}
		}
	}
}

func getModule(moduleUrl string, creds Auth) (string, []string) {
	//proxy, _ := url.Parse("http://localhost:8080")
	//tr := &http.Transport{Proxy: http.ProxyURL(proxy), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{}
	//client.Transport = tr
	req, err := http.NewRequest("GET", moduleUrl, nil)
	if err != nil {
		die(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:102.0) Gecko/20100101 Firefox/102.0")
	req.Header.Add("Cookie", creds.cookies)

	resp, err := client.Do(req)
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
		pagesContent = append(pagesContent, extractPageContent(pageUrl, creds))
	}

	return moduleTitle, pagesContent
}

func getModuleTitle(htmlText string) string {
	var title string
	var isTitle bool
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
				if strings.Contains(c.Attr[0].Val, moduleUrl[:44]) {
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

func extractPageContent(pageUrl string, creds Auth) string {
	var result string
	pageContent := getModulePageContent(pageUrl, creds)

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

func getModulePageContent(pageUrl string, creds Auth) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", pageUrl, nil)
	if err != nil {
		die(err)
	}
	req.Header.Add("Cookie", creds.cookies)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:102.0) Gecko/20100101 Firefox/102.0")

	resp, err := client.Do(req)
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

func die(err error) {
	fmt.Println(err)
	os.Exit(1)
}
