package client

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"ghostviewer/io"
	"net"
	"os"
	"strconv"
	"time"
)

const headerSize = 5
const packetPrefix = '%'

type Message struct {
	Cmd  string
	Data []byte
}

type TCPGClient struct {
	Ip   string
	Port int
	Conn *net.TCPConn
}

func (h *TCPGClient) Connect() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", h.Ip+":"+strconv.Itoa(h.Port))

	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		h.Conn, err = net.DialTCP("tcp", nil, tcpAddr)
		if err == nil {
			break
		}
		fmt.Println(err)
		fmt.Println("Connect failed, sleeping 5 seconds")
		time.Sleep(5 * time.Second)
	}

	return nil
}

func (h *TCPGClient) ProcessRead(data *bytes.Buffer) (done bool) {
	done = false

	for {
		if len(data.Bytes()) == 0 {
			return
		}

		if data.Bytes()[0] != packetPrefix { // check for prefix
			done = true
			return
		}

		var dataSize int
		if data.Len() > headerSize {
			dataSize = int(binary.LittleEndian.Uint32(data.Bytes()[1:5])) // All UI events follow the partten Left Mouse Click -> ML:x:y;, Key A Press -> K:char:idx;
		} else {
			return
		}

		packetSize := headerSize + dataSize
		if data.Len() < packetSize {
			return
		}

		packet := data.Next(packetSize)[5:]
		encodedMsgs := bytes.Split(packet, []byte{';'})

		for _, msg := range encodedMsgs {
			if len(msg) == 0 {
				continue
			}

			io.PassMessageToIODriver(msg)
		}
	}
}
func (h *TCPGClient) Receive() {
	localBuf := new(bytes.Buffer)
	readBuf := make([]byte, 256)

	for {
		dataLen, _ := h.Conn.Read(readBuf)
		localBuf.Write(readBuf[:dataLen])
		h.ProcessRead(localBuf)
	}
}

func (h *TCPGClient) SendMessage(msg Message) error {
	bin_buf := new(bytes.Buffer)
	gobobj := gob.NewEncoder(bin_buf)
	gobobj.Encode(msg)
	databytes := bin_buf.Bytes()
	packetPrefix := []byte{'%'}
	packetSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(packetSize, uint32(len(databytes)))
	packet := append(append(packetPrefix, packetSize[:]...), databytes[:]...)
	_, err := h.Conn.Write(packet)
	return err
}

func (h *TCPGClient) SendFrame(img []byte, width int, height int) error {
	if h.Conn == nil {
		fmt.Println("Invalid connection")
		os.Exit(1)
	}
	msg := Message{"FRAME:" + strconv.Itoa(width) + ":" + strconv.Itoa(height), img}
	return h.SendMessage(msg)
}

func (h *TCPGClient) Disconnect(message string) {
	h.Conn.Close()
}
