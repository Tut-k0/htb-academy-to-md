package main

import (
	"flag"
	"fmt"
	"os"
)

type Args struct {
	moduleUrl   string
	email       string
	password    string
	localImages bool
}

func getArguments() Args {
	var mFlag = flag.String("m", "", "(REQUIRED) Academy Module URL to the first page.")
	var eFlag = flag.String("e", "", "(REQUIRED) Email for your HTB Academy account.")
	var pFlag = flag.String("p", "", "(REQUIRED) Password for your HTB Academy account.")
	var imgFlag = flag.Bool("images", false, "Save images locally rather than referencing the static URL.")
	flag.Parse()
	arg := Args{
		moduleUrl:   *mFlag,
		email:       *eFlag,
		password:    *pFlag,
		localImages: *imgFlag,
	}

	if arg.moduleUrl == "" || arg.email == "" || arg.password == "" {
		fmt.Println("Missing arguments, please use the -h option to display the arguments required to run this application!")
		os.Exit(1)
	}

	return arg
}
