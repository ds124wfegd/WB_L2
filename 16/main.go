package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ds124wfegd/WB_L2/16/config"
	"github.com/ds124wfegd/WB_L2/16/downloader"
)

func main() {
	url := flag.String("url", "", "URL to download (required)")
	output := flag.String("output", "./download", "Output directory")
	depth := flag.Int("depth", 2, "Max recursion depth")
	workers := flag.Int("workers", 3, "Number of workers")
	flag.Parse()

	if *url == "" {
		fmt.Println("Ошибка: необходим URL")
		flag.Usage()
		os.Exit(1)
	}

	cfg := config.NewConfig(*url, *output, *depth, *workers)

	dl := downloader.NewDownloader(cfg)
	log.Printf("Начало загрузки c %s в дирректорию %s", cfg.URL, cfg.Output)

	if err := dl.Start(); err != nil {
		log.Fatal("Ошибка:", err)
	}

	log.Println("Загрузка завершена!")
}
