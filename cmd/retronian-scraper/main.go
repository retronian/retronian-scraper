package main

import (
	"fmt"
	"os"

	"github.com/retronian/retronian-scraper/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "scan":
		os.Exit(cli.Scan(os.Args[2:]))
	case "-h", "--help", "help":
		usage()
		return
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println(`Retronian Scraper — 多言語 ROM メタデータツール (CLI)

Usage:
  retronian-scraper scan <rom-dir> --platform <id> [--out gamelist.xml]
  retronian-scraper help`)
}
