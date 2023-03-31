package main

import "fmt"

func main() {
	options := getArguments()
	session := authenticate(options.email, options.password)
	title, content := getModule(options.moduleUrl, session)
	fmt.Println(title, content)
}
