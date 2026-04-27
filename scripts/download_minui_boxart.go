package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/retronian/retronian-scraper/internal/export"
	"github.com/retronian/retronian-scraper/internal/match"
	"github.com/retronian/retronian-scraper/internal/pipeline"
)

type platform struct {
	id     string
	folder string
}

var platforms = []platform{
	{"gb", "ゲームボーイ (GB)"},
	{"gbc", "ゲームボーイカラー (GBC)"},
	{"gba", "ゲームボーイアドバンス (GBA)"},
	{"fc", "ファミリーコンピュータ (FC)"},
	{"sfc", "スーパーファミコン (SFC)"},
	{"pce", "PCエンジン (PCE)"},
	{"ps1", "プレイステーション (PS)"},
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: go run ./scripts/download_minui_boxart.go <rom-root> <api-base-url>")
		os.Exit(2)
	}
	root := os.Args[1]
	api := os.Args[2]
	client := &http.Client{Timeout: 30 * time.Second}

	var total, downloaded, exists, noMedia, unmatched, failed int
	for _, p := range platforms {
		dir := filepath.Join(root, p.folder)
		out, err := pipeline.Run(context.Background(), pipeline.Options{
			ROMDir:   dir,
			Platform: p.id,
			BaseURL:  api,
		}, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", p.id, err)
			failed++
			continue
		}

		resDir := filepath.Join(dir, ".res")
		if err := os.MkdirAll(resDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", p.id, err)
			failed++
			continue
		}

		var tasks []downloadTask
		var pExists, pNoMedia, pUnmatched int
		for _, r := range out.Results {
			total++
			if r.Game == nil || r.Tier == match.TierNone {
				unmatched++
				pUnmatched++
				continue
			}
			url := export.PickBoxart(r.Game.Media)
			if url == "" {
				noMedia++
				pNoMedia++
				continue
			}
			target := filepath.Join(resDir, filepath.Base(r.Path)+".png")
			if _, err := os.Stat(target); err == nil {
				exists++
				pExists++
				continue
			}
			tasks = append(tasks, downloadTask{platform: p.id, name: filepath.Base(r.Path), url: url, target: target})
		}
		pDownloaded, pFailed := downloadAll(client, tasks)
		downloaded += pDownloaded
		failed += pFailed
		fmt.Printf("%s: downloaded=%d exists=%d no_media=%d unmatched=%d failed=%d\n", p.id, pDownloaded, pExists, pNoMedia, pUnmatched, pFailed)
	}

	fmt.Printf("summary: total=%d downloaded=%d exists=%d no_media=%d unmatched=%d failed=%d\n", total, downloaded, exists, noMedia, unmatched, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

type downloadTask struct {
	platform string
	name     string
	url      string
	target   string
}

func downloadAll(client *http.Client, tasks []downloadTask) (int, int) {
	if len(tasks) == 0 {
		return 0, 0
	}
	const workers = 2
	work := make(chan downloadTask)
	var done, okCount, failCount int64
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range work {
				if err := download(client, task.url, task.target); err != nil {
					fmt.Fprintf(os.Stderr, "%s: %s: %v\n", task.platform, task.name, err)
					atomic.AddInt64(&failCount, 1)
				} else {
					atomic.AddInt64(&okCount, 1)
				}
				current := atomic.AddInt64(&done, 1)
				if current%50 == 0 || current == int64(len(tasks)) {
					fmt.Fprintf(os.Stderr, "downloaded %d/%d pending images\n", current, len(tasks))
				}
			}
		}()
	}
	for _, task := range tasks {
		work <- task
	}
	close(work)
	wg.Wait()
	return int(okCount), int(failCount)
}

func download(client *http.Client, url, target string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("unsupported URL %q", url)
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "retronian-scraper/boxart")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %s", resp.Status)
	}

	tmp := target + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(f, resp.Body)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}
	return os.Rename(tmp, target)
}
