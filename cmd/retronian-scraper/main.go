package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.NewWithID("com.retronian.retronian-scraper")
	w := a.NewWindow("Retronian Scraper")

	title := widget.NewLabelWithStyle(
		"Retronian Scraper",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	subtitle := widget.NewLabel("多言語 ROM スクレイパー — Phase 0 (Fyne 動作検証)")
	body := widget.NewLabel("こんにちは、レトロニアン。\nFyne での日本語表示を確認しています。")

	w.SetContent(container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		body,
	))
	w.Resize(fyne.NewSize(480, 240))
	w.ShowAndRun()
}
