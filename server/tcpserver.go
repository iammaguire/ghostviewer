package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"ghostviewer/ui"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

const headerSize = 5
const packetPrefix = '%'

type TCPGServer struct {
	Ip   string
	Port int
	Conn net.Conn
}

func (h *TCPGServer) IsConnected() bool {
	return h.Conn != nil && h.Conn.RemoteAddr() != nil
}

func (h *TCPGServer) Listen() error {
	fmt.Println("Waiting for TCP client to connect...")
	l, err := net.Listen("tcp", h.Ip+":"+strconv.Itoa(h.Port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Listen error: %s\n", err)
		return err
	}

	h.Conn, err = l.Accept()
	l.Close()
	return err
}

func (h *TCPGServer) Close() {
	h.Conn.Close()
	h.Conn = nil
}

func (h *TCPGServer) ProcessInput(output chan ui.Message, data *bytes.Buffer) (disconnect bool) {
	disconnect = false

	for {
		if len(data.Bytes()) == 0 {
			return
		}

		if data.Bytes()[0] != packetPrefix { // check for prefix
			disconnect = true
			return
		}

		var dataSize int
		if data.Len() > headerSize {
			dataSize = int(binary.LittleEndian.Uint32(data.Bytes()[1:5]))
		} else {
			return
		}

		packetSize := headerSize + dataSize
		if data.Len() < packetSize {
			return
		}

		packet := data.Next(packetSize)[5:]
		//fmt.Println(len(packet))
		tmpbuff := bytes.NewBuffer(packet)
		msg := new(ui.Message)
		gobobjdec := gob.NewDecoder(tmpbuff)
		err := gobobjdec.Decode(msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't decode packet: %s", err)
		}

		output <- *msg

		// send UI events back

		uiMsgHeader := <-output
		numUIEvents, _ := strconv.Atoi(strings.Split(uiMsgHeader.Cmd, ":")[1])
		if numUIEvents != 0 {
			packetPrefix := []byte{'%'}
			encodedEvents := []byte{}
			uiPacketSize := make([]byte, 4)
			for i := 0; i < numUIEvents; i++ {
				encodedEvents = append(encodedEvents, ui.EncodeEvent(<-output)[:]...)
			}
			binary.LittleEndian.PutUint32(uiPacketSize, uint32(len(encodedEvents)))
			packet = append(append(packetPrefix, uiPacketSize[:]...), encodedEvents[:]...)
			h.Conn.Write(packet)
		}
	}
}

func (h *TCPGServer) Receive(output chan ui.Message) {
	localBuf := new(bytes.Buffer)
	readBuf := make([]byte, 4096)
	defer h.Close()

	for {
		dataLen, err := h.Conn.Read(readBuf)
		if err != nil {
			if err == io.EOF {
				h.Close()
			}
		}

		localBuf.Write(readBuf[:dataLen])
		if h.ProcessInput(output, localBuf) {
			break
		}
	}
}
