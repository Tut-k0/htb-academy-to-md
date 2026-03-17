package main

import (
	"flag"
	"fmt"
	"os"
)

type Args struct {
	moduleUrl   string
	cookies     string
	localImages bool
	outputDir   string
}

func getArguments() Args {
	var mFlag = flag.String("m", "", "(REQUIRED) Academy Module URL to the first page.")
	var cFlag = flag.String("c", "", "(REQUIRED) Academy Cookies for authorization.")
	var imgFlag = flag.Bool("local_images", false, "Save images locally rather than referencing the URL location.")
	var oFlag = flag.String("o", "", "Output directory for the generated Markdown file (defaults to current directory).")
	flag.Parse()
	arg := Args{
		moduleUrl:   *mFlag,
		cookies:     *cFlag,
		localImages: *imgFlag,
		outputDir:   *oFlag,
	}

	if arg.moduleUrl == "" || arg.cookies == "" {
		fmt.Println("Missing required arguments for module URL and HTB Academy Cookies. Please use the -h option for help.")
		os.Exit(1)
	}

	return arg
}
