package main

import (
	"flag"
	"fmt"
	"os"
)

type Args struct {
	moduleUrl string
	cookies   string
	//email       string
	//password    string
	localImages bool
}

func getArguments() Args {
	var mFlag = flag.String("m", "", "(REQUIRED) Academy Module URL to the first page.")
	var cFlag = flag.String("c", "", "(REQUIRED) Academy Cookies for authorization.")
	//var eFlag = flag.String("e", "", "(REQUIRED) Email for your HTB account.")
	//var pFlag = flag.String("p", "", "Password for your HTB account.")
	var imgFlag = flag.Bool("local_images", false, "Save images locally rather than referencing the URL location.")
	flag.Parse()
	arg := Args{
		moduleUrl: *mFlag,
		cookies:   *cFlag,
		//email:       *eFlag,
		//password:    *pFlag,
		localImages: *imgFlag,
	}

	//if arg.moduleUrl == "" || arg.email == "" {
	//	fmt.Println("Missing required arguments for module URL and HTB email. Please use the -h option for help.")
	//	os.Exit(1)
	//}
	//
	//if arg.password == "" {
	//	fmt.Print("Enter Password: ")
	//	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	//	if err != nil {
	//		fmt.Println("\nError reading password:", err)
	//		os.Exit(1)
	//	}
	//	arg.password = string(passwordBytes)
	//	fmt.Println()
	//}

	if arg.moduleUrl == "" || arg.cookies == "" {
		fmt.Println("Missing required arguments for module URL and HTB Academy Cookies. Please use the -h option for help.")
		os.Exit(1)
	}

	return arg
}
