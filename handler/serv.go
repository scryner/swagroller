package handler

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/browser"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/scryner/swagroller/static"
)

func DoServer(inputFilepath string, port int, openBrowser bool) error {
	err := refreshIndex(inputFilepath, port)
	if err != nil {
		return fmt.Errorf("failed to build index.html: %v", err)
	}

	// make websockets to refresh web browser when contents are updated
	ws := newWebSockets()

	// start watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to make file notifier: %v", err)
	}

	baseDir, filename := filepath.Split(inputFilepath)

	err = watcher.Add(baseDir)
	if err != nil {
		return fmt.Errorf("failed to add dir to notify: %v", err)
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

					if err := refreshIndex(inputFilepath, port); err != nil {
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

	if openBrowser {
		go func() {
			url := fmt.Sprintf("http://localhost:%d", port)

			// sleep while to start server
			time.Sleep(time.Millisecond * 200)
			_ = browser.OpenURL(url)
		}()
	}

	return http.ListenAndServe(addr, nil)
}

func refreshIndex(inputFilepath string, port int) error {
	title, jsonb, err := readYAMLtoJSON(inputFilepath)
	if err != nil {
		return err
	}

	// make index.html
	buf := new(bytes.Buffer)
	err = static.MakeIndexHTML(title, jsonb, buf, true, port)
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
