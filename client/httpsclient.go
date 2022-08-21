package client

import (
	"bytes"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
)

type HTTPSGClient struct {
	Ip   string
	Port int
	Conn *websocket.Conn
}

func (h *HTTPSGClient) Connect() error {
	u := url.URL{Scheme: "wss", Host: h.Ip + ":" + strconv.Itoa(h.Port), Path: "/ws"}
	fmt.Println("Connecting to " + u.String())
	dialer := *websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c, _, err := dialer.Dial(u.String(), nil)
	h.Conn = c

	if err != nil {
		fmt.Printf("Handshake failed with server: %s", err)
		os.Exit(1)
	}

	return nil
}

func (h *HTTPSGClient) Receive() {
	if _, _, err := h.Conn.ReadMessage(); err != nil {
		fmt.Println(err)
		return
	} else {
		//fmt.Println(messageType)
	}
}

func (h *HTTPSGClient) SendMessage(msg Message) error {
	bin_buf := new(bytes.Buffer)
	gobobj := gob.NewEncoder(bin_buf)
	gobobj.Encode(msg)
	databytes := bin_buf.Bytes()
	err := h.Conn.WriteMessage(websocket.BinaryMessage, databytes)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return err
}

func (h *HTTPSGClient) SendFrame(img []byte, width int, height int) error {
	if h.Conn == nil {
		fmt.Println("Invalid connection")
		os.Exit(1)
	}
	msg := Message{"FRAME:" + strconv.Itoa(width) + ":" + strconv.Itoa(height), img}
	return h.SendMessage(msg)
}

func (h *HTTPSGClient) Disconnect(message string) {
}
