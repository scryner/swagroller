package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/scryner/swagroller/static"
	"github.com/fsnotify/fsnotify"
)

func doServer(usage func(), inputFilepath string, port int) {
	// check file existed
	if inputFilepath == "" {
		usage()
		os.Exit(1)
	}

	err := refreshIndex(inputFilepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build index.html: %v\n", err)
		os.Exit(1)
	}

	// start watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to make file notifier: %v\n", err)
		os.Exit(1)
	}

	baseDir, filename := filepath.Split(inputFilepath)

	err = watcher.Add(baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to add dir to notify: %v\n", err)
		os.Exit(1)
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				if ev.Name != filename {
					continue
				}

				if ev.Op&fsnotify.Create == fsnotify.Create ||
					ev.Op&fsnotify.Write == fsnotify.Write {

					if err := refreshIndex(inputFilepath); err != nil {
						log.Println("fsnotify refresh:", err)
					} else {
						log.Println("fsnotify: refresh 'index.html' is completed")
						watcher.Add(inputFilepath)
					}
				}

			case err := <-watcher.Errors:
				log.Println("fsnotify:", err)
			}
		}
	}()

	// start server
	addr := fmt.Sprintf(":%d", port)
	log.Printf("starting server at '%s'", addr)

	http.Handle("/", http.FileServer(static.FS(false)))
	log.Fatal(http.ListenAndServe(addr, nil))
}

func refreshIndex(inputFilepath string) error {
	title, jsonb, err := readYAMLtoJSON(inputFilepath)
	if err != nil {
		return err
	}

	// make index.html
	buf := new(bytes.Buffer)
	err = static.MakeIndexHTML(title, jsonb, buf)
	if err != nil {
		return fmt.Errorf("failed to make index.html: %v", err)
	}

	// refresh index.html
	err = static.AddFile("index.html", buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to add file: %v", err)
	}

	return nil
}
