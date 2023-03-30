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
	auth := getLoginToken()
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

func getLoginToken() Auth {
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

func die(err error) {
	fmt.Println(err)
	os.Exit(1)
}
