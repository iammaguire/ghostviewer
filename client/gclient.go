package client

import (
	"fmt"
	"ghostviewer/screenshot"
	"os"
	"runtime"
	"time"
)

var fps = 30

type GClient interface {
	Connect() error
	Receive()
	Disconnect(string)
	SendFrame([]byte, int, int) error
}

func ClientCommunicate(ghostclient GClient) {
	runtime.LockOSThread() // lock so windows/dxgi/d3d11 can use threadlocal caches, if any
	screenshot.InitializeScreenshotInterface()
	defer func() {
		if screenshot.DDUP != nil {
			screenshot.DDUP.Release()
			screenshot.DDUP = nil
		}

		screenshot.Device.Release()
		screenshot.DeviceCtx.Release()
	}()

	go func() {
		j := 0
		t := time.Now()
		for {
			if time.Since(t) >= time.Second {
				j = 0
				t = time.Now()
			}
			cap, _ := screenshot.CaptureScreen()
			j++

			if cap == nil {
				continue
			}

			if err := ghostclient.SendFrame(cap.Pix, cap.Bounds().Dx(), cap.Bounds().Dy()); err != nil {
				fmt.Fprintf(os.Stderr, "Send error: %s\nAttempting reconnect...", err)
				time.Sleep(5 * time.Second)
				fmt.Println(err)
				os.Exit(1)
			}
			time.Sleep(time.Millisecond * 0)
		}
	}()

	for {
		ghostclient.Receive()
	}
}
