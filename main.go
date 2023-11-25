package main

import (
	"embed"
	"flag"
	"io/fs"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rnetx/server-status-go/core"
)

//go:embed dist
var webUIFiles embed.FS

type fsFunc func(name string) (fs.File, error)

func (f fsFunc) Open(name string) (fs.File, error) {
	return f(name)
}

var (
	listen = ":8080"
)

func init() {
	flag.StringVar(&listen, "listen", listen, "listen address")
}

func main() {
	flag.Parse()

	serverMux := http.NewServeMux()
	serverMux.Handle("/ws", http.HandlerFunc(handle))
	serverMux.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(fsFunc(func(name string) (fs.File, error) {
		assetPath := path.Join("dist", name)
		file, err := webUIFiles.Open(assetPath)
		if err != nil {
			return nil, err
		}
		return file, err
	})))))
	server := &http.Server{
		Addr:    listen,
		Handler: serverMux,
	}
	server.ListenAndServe()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := conn.WriteJSON(core.GetAll(r.Context()))
			if err != nil {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}
