package main

import "fmt"

func main() {
	options := getArguments()
	session := authenticate(options.email, options.password)
	fmt.Println(session)
}
