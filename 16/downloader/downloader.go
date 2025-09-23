package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ds124wfegd/WB_L2/16/config"
	"github.com/ds124wfegd/WB_L2/16/parser"
)

type Downloader struct {
	config  *config.Config
	client  *http.Client
	visited sync.Map
	queue   chan *Task
	wg      sync.WaitGroup
	mu      sync.Mutex
	closed  bool
}

func NewDownloader(cfg *config.Config) *Downloader {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Downloader{
		config: cfg,
		client: client,
		queue:  make(chan *Task, 1000),
	}
}

func (d *Downloader) Start() error {
	if err := os.MkdirAll(d.config.Output, 0755); err != nil {
		return err
	}

	for i := 0; i < d.config.Workers; i++ {
		d.wg.Add(1)
		go d.worker(i)
	}

	d.addTask(&Task{URL: d.config.URL, Depth: 0})

	go d.monitorCompletion()

	d.wg.Wait()

	return nil
}

func (d *Downloader) worker(id int) {
	defer d.wg.Done()

	for task := range d.queue {
		log.Printf("Воркер %d: %s (глубина рекурсии %d)", id, task.URL, task.Depth)

		if err := d.download(task); err != nil {
			log.Printf("Ошибка скачивания %s: %v", task.URL, err)
		}

		time.Sleep(100 * time.Millisecond)
	}
	log.Printf("Воркер %d закончил работу", id)
}

func (d *Downloader) addTask(task *Task) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}

	select {
	case d.queue <- task:
	default:
		log.Printf("Очередь заполнена: %s", task.URL)
	}
}

func (d *Downloader) monitorCompletion() {
	time.Sleep(2 * time.Second)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	inactiveCount := 0
	lastActivity := time.Now()

	for range ticker.C {
		queueLen := len(d.queue)

		if queueLen == 0 {
			inactiveCount++

			if inactiveCount >= 5 || time.Since(lastActivity) > 30*time.Second {
				d.mu.Lock()
				if !d.closed {
					d.closed = true
					close(d.queue)
					log.Println("Очередь закрыта")
				}
				d.mu.Unlock()
				return
			}
		} else {
			inactiveCount = 0
			lastActivity = time.Now()
		}
	}
}

func (d *Downloader) download(task *Task) error {

	if _, visited := d.visited.Load(task.URL); visited {
		return nil
	}
	d.visited.Store(task.URL, true)

	req, err := http.NewRequest("GET", task.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MyDownloader/1.0)")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	filename := d.getFilename(task.URL, resp.Header.Get("Content-Type"))
	fullPath := filepath.Join(d.config.Output, filename)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	log.Printf("Скачано: %s", filename)

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") && task.Depth < d.config.MaxDepth {
		if err := d.parseHTML(fullPath, task.URL, task.Depth+1); err != nil {
			log.Printf("Ошибка парсинга HTML: %v", err)
		}
	}

	return nil
}

func (d *Downloader) getFilename(rawURL, contentType string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Sprintf("file_%d.html", time.Now().UnixNano())
	}

	host := strings.ReplaceAll(parsed.Host, ":", "_")
	urlPath := parsed.Path

	if urlPath == "" || urlPath == "/" {
		urlPath = "/index.html"
	}

	urlPath = strings.TrimPrefix(urlPath, "/")

	ext := ".html"
	if strings.Contains(contentType, "css") {
		ext = ".css"
	} else if strings.Contains(contentType, "javascript") {
		ext = ".js"
	} else if strings.Contains(contentType, "image") {
		ext = ".jpg"
	}

	if strings.Contains(filepath.Base(urlPath), ".") {
		ext = ""
	}

	filename := filepath.Join(host, urlPath) + ext

	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "&", "_")
	filename = strings.ReplaceAll(filename, "=", "_")

	return filename
}

func (d *Downloader) parseHTML(filepath, baseURL string, depth int) error {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	html := string(content)
	base, _ := url.Parse(baseURL)

	links := parser.ExtractLinks(html, base)

	count := 0
	for _, link := range links {
		if d.isSameDomain(link, baseURL) {
			if _, visited := d.visited.Load(link); !visited {
				d.addTask(&Task{URL: link, Depth: depth})
				count++
				if count >= 50 {
					break
				}
			}
		}
	}

	log.Printf("Найдена %d новая ссылка в %s", count, baseURL)
	return nil
}

func (d *Downloader) isSameDomain(link, baseURL string) bool {
	linkURL, err1 := url.Parse(link)
	base, err2 := url.Parse(baseURL)

	if err1 != nil || err2 != nil {
		return false
	}

	return linkURL.Host == base.Host
}
