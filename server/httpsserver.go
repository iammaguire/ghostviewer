package server

import (
	"crypto/tls"
	"fmt"
	"ghostviewer/ui"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/acme/autocert"
)

type HTTPSGServer struct {
	Ip        string
	Port      int
	Connected bool
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
}

func (h *HTTPSGServer) IsConnected() bool {
	return h.Connected
}

func ProcessInput(conn *websocket.Conn) {
	for {
		messageType, p, err := conn.ReadMessage()

		if err != nil {
			fmt.Fprintf(os.Stdout, "Websocket read error: %s", err)
			return
		}

		fmt.Println("Got message!")

		if err := conn.WriteMessage(messageType, p); err != nil {
			fmt.Fprintf(os.Stdout, "Websocket write error: %s", err)
			return
		}

	}
}

func (h *HTTPSGServer) Endpoint(w http.ResponseWriter, r *http.Request) {
	h.Connected = true
	upgrader.CheckOrigin = func(r *http.Request) bool { return true } // remove CORS error
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Fprintf(os.Stdout, "Failed to upgrade client to websocket: %s", err)
	}

	fmt.Println("Client connected")

	ProcessInput(ws)
}

func (h *HTTPSGServer) Listen() error {
	fmt.Println("Waiting for HTTPS client to connect...")

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		h.Endpoint(w, r)
	})

	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("certs"),
	}

	certManager.TLSConfig().ServerName = "ghostclient"

	server := &http.Server{
		Addr:    h.Ip + ":" + strconv.Itoa(h.Port),
		Handler: mux,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	go func() {
		if ex, err := os.Executable(); err == nil {
			exPath := filepath.Dir(ex)

			if err := server.ListenAndServeTLS(exPath+"/certs/localhost.crt", exPath+"/certs/localhost.key"); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			fmt.Println(err)
		}
	}()

	return nil
}

func (h *HTTPSGServer) Close() {
	h.Connected = false
}

func (h *HTTPSGServer) Receive(output chan ui.Message) {

}
