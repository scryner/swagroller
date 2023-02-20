package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/scryner/swagroller/static"
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

	// make websockets to refresh web browser when contents are updated
	ws := newWebSockets()

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

						// notify all channels to wait update
						ws.notifyAll()
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
	http.Handle("/websocket", handleWebSocket(ws))
	log.Fatal(http.ListenAndServe(addr, nil))
}

func refreshIndex(inputFilepath string) error {
	title, jsonb, err := readYAMLtoJSON(inputFilepath)
	if err != nil {
		return err
	}

	// make index.html
	buf := new(bytes.Buffer)
	err = static.MakeIndexHTML(title, jsonb, buf, true)
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

type webSockets struct {
	lock *sync.RWMutex
	m    map[uint32]chan int
}

func newWebSockets() *webSockets {
	return &webSockets{
		lock: new(sync.RWMutex),
		m:    make(map[uint32]chan int),
	}
}

func (ws *webSockets) register() (id uint32, ch chan int) {
	ws.lock.Lock()
	defer ws.lock.Unlock()

	// make id
	for {
		id = rand.Uint32()
		if _, ok := ws.m[id]; !ok {
			break
		}
	}

	ch = make(chan int)
	ws.m[id] = ch

	return
}

func (ws *webSockets) notifyAll() {
	ws.lock.RLock()
	defer ws.lock.RUnlock()

	for _, ch := range ws.m {
		ch <- 1
	}
}

func (ws *webSockets) unregister(id uint32) {
	ws.lock.Lock()
	defer ws.lock.Unlock()

	delete(ws.m, id)
}

type command struct {
	Command string `json:"command"`
}

func handleWebSocket(ws *webSockets) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("failed to upgrade websocket:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// register web socket to receive update signal
		id, ch := ws.register()
		defer func() {
			ws.unregister(id)
		}()

		// wait signal to refresh
		for {
			select {
			case <-ch:
				// send to client
				if err := conn.WriteJSON(command{"UPDATE"}); err != nil {
					break
				}
			}
		}
	}
}
