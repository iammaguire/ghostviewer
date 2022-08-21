package ui

import (
	"ghostviewer/io"
	"image"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

type Message struct {
	Cmd  string
	Data []byte
}

type GRenderer struct {
	CurFrame     *ebiten.Image
	KBHandler    *io.KbInputHandler
	Messages     chan Message
	RemoteWidth  int
	RemoteHeight int
	LocalWidth   int
	LocalHeight  int
	RemoteMouseX int
	RemoteMouseY int
	LocalMouseX  int
	LocalMouseY  int
}

func EncodeEvent(msg Message) []byte {
	return append([]byte(msg.Cmd), ';')
}

var keyCounter = 0

func (gr *GRenderer) Update() error {
	if gr.CurFrame != nil {
		gr.RemoteWidth = gr.CurFrame.Bounds().Dx()
		gr.RemoteHeight = gr.CurFrame.Bounds().Dy()
		x, y := ebiten.CursorPosition()
		gr.RemoteMouseY = y * gr.RemoteHeight / gr.LocalHeight
		gr.RemoteMouseX = x * gr.RemoteWidth / gr.LocalWidth
		if gr.LocalMouseX != x || gr.LocalMouseY != y {
			go func() {
				gr.Messages <- Message{"MOUSE:" + strconv.Itoa(gr.RemoteMouseX) + ":" + strconv.Itoa(gr.RemoteMouseY), nil}
			}()
		}
		gr.LocalMouseX, gr.LocalMouseY = x, y

		for _, e := range gr.KBHandler.GetEvents() {
			go func() {
				gr.Messages <- Message{"KEY:" + strconv.Itoa(int(e.Kind)) + ":" + strconv.Itoa(keyCounter) + ":" + string(e.Keychar), nil}
			}()
			keyCounter++
		}
	}

	return nil
}

var leftPressedLast = false
var rightPressedLast = false

func (gr *GRenderer) HandleMouse() {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if !leftPressedLast {
			leftPressedLast = true
			go func() {
				gr.Messages <- Message{"LDMOUSE:" + strconv.Itoa(gr.RemoteMouseX) + ":" + strconv.Itoa(gr.RemoteMouseY), nil}
			}()
		}
	} else {
		if leftPressedLast {
			go func() {
				gr.Messages <- Message{"LUMOUSE:" + strconv.Itoa(gr.RemoteMouseX) + ":" + strconv.Itoa(gr.RemoteMouseY), nil}
			}()
		}
		leftPressedLast = false
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		if !rightPressedLast {
			rightPressedLast = true
			go func() {
				gr.Messages <- Message{"RDMOUSE:" + strconv.Itoa(gr.RemoteMouseX) + ":" + strconv.Itoa(gr.RemoteMouseY), nil}
			}()
		}
	} else {
		if rightPressedLast {
			go func() {
				gr.Messages <- Message{"RUMOUSE:" + strconv.Itoa(gr.RemoteMouseX) + ":" + strconv.Itoa(gr.RemoteMouseY), nil}
			}()
		}
		rightPressedLast = false
	}
}

func (gr *GRenderer) Draw(screen *ebiten.Image) {
	if gr.CurFrame != nil {
		screen.DrawImage(gr.CurFrame, nil)
		gr.HandleMouse()
	}
}

func (gr *GRenderer) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	if gr.RemoteWidth <= 0 || gr.RemoteHeight <= 0 {
		return 1280, 720
	}
	return gr.RemoteWidth, gr.RemoteHeight
}

func NewGRenderer() *GRenderer {
	gr := &GRenderer{LocalWidth: 1920, LocalHeight: 1080, Messages: make(chan Message), KBHandler: &io.KbInputHandler{}}
	return gr
}

func (gr *GRenderer) UpdateFrame(frame *image.RGBA) {
	img := ebiten.NewImageFromImage(frame)
	gr.CurFrame = img
}
