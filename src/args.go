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
    dirTree     bool
}

func getArguments() Args {
	var mFlag = flag.String("m", "", "(REQUIRED) Academy Module URL to the first page.")
	var cFlag = flag.String("c", "", "(REQUIRED) Academy Cookies for authorization.")
	var imgFlag = flag.Bool("local_images", false, "Save images locally rather than referencing the URL location.")
	var directoryTreeFlag = flag.Bool("dir_tree", false, "Save sections and images in a directory Tree.")
	flag.Parse()
	arg := Args{
		moduleUrl:   *mFlag,
		cookies:     *cFlag,
		localImages: *imgFlag,
        dirTree:     *directoryTreeFlag,

	}

	if arg.moduleUrl == "" || arg.cookies == "" {
		fmt.Println("Missing required arguments for module URL and HTB Academy Cookies. Please use the -h option for help.")
		os.Exit(1)
	}

	return arg
}
