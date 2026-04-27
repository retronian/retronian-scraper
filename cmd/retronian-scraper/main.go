package main

import (
	"fmt"
	"os"

	"github.com/retronian/retronian-scraper/internal/cli"
	"github.com/retronian/retronian-scraper/internal/gui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "scan":
			os.Exit(cli.Scan(os.Args[2:]))
		case "normalize":
			os.Exit(cli.Normalize(os.Args[2:]))
		case "-h", "--help", "help":
			usage()
			return
		}
	}
	gui.Run()
}

func usage() {
	fmt.Println(`Retronian Scraper - multilingual ROM metadata tool

Usage:
  retronian-scraper                          Start the GUI
  retronian-scraper scan <dir>               Run a CLI scan; pass -h for details
  retronian-scraper normalize <parent-dir>   Normalize ROM folder names for a frontend
  retronian-scraper help                     Show this help`)
}
