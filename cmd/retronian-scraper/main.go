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
	fmt.Println(`Retronian Scraper — 多言語 ROM メタデータツール

Usage:
  retronian-scraper                          GUI を起動
  retronian-scraper scan <dir>               CLI スキャン (詳細は -h を付けて実行)
  retronian-scraper normalize <parent-dir>   ROM フォルダ名を frontend 規約に正規化
  retronian-scraper help                     このヘルプ`)
}
