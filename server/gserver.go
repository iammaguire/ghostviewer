package server

import (
	"ghostviewer/ui"
	"image"
	"strconv"
	"strings"
)

type GServer interface {
	Receive(chan ui.Message)
	Listen() error
	Close()
	IsConnected() bool
}

func ServerViewer(ghostserver GServer, grenderer *ui.GRenderer) {
	var uiMsgStack []ui.Message
	messages := make(chan ui.Message)
	go ghostserver.Receive(messages)
	go func() {
		for {
			uiMsgStack = append(uiMsgStack, <-grenderer.Messages)
		}
	}()

	for {
		msg := <-messages
		messages <- ui.Message{"UI:" + strconv.Itoa(len(uiMsgStack)), nil}
		for _, uiMsg := range uiMsgStack {
			messages <- uiMsg
		}
		uiMsgStack = []ui.Message{}

		cmdlist := strings.Split(msg.Cmd, ":")
		cmd := cmdlist[0]
		args := cmdlist[1:]

		if cmd == "FRAME" {
			w, _ := strconv.Atoi(args[0])
			h, _ := strconv.Atoi(args[1])
			imageBytes := make([]byte, len(msg.Data))

			for i := 0; i < len(imageBytes); i += 4 {
				imageBytes[i], imageBytes[i+2], imageBytes[i+1], imageBytes[i+3] = msg.Data[i+2], msg.Data[i], msg.Data[i+1], msg.Data[i+3]
			}

			grenderer.UpdateFrame(&image.RGBA{Pix: imageBytes, Stride: w * 4, Rect: image.Rect(0, 0, w, h)})
		}

	}
}
